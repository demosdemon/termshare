package termshare

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/demosdemon/termshare/pkg/utils"
)

type Recorder interface {
	Record(w io.Writer) error
}

type RawRecorder struct {
	Command string
}

func (r *RawRecorder) Record(w io.Writer) error {
	var args []string
	if r.Command != "" {
		args = []string{"-c", r.Command}
	}
	cmd := exec.Command(utils.Shell(), args...)

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		_ = pw.Close()
		_ = pr.Close()
		return err
	}

	_ = pw.Close()

	done := make(chan struct{})
	defer func() { <-done }()
	go func() {
		defer close(done)
		_, _ = io.Copy(io.MultiWriter(w, os.Stdout), pr)
	}()

	defer func() { _ = pr.Close() }()
	return cmd.Wait()
}

type CastRecorder struct {
	Command            string
	IdleTimeLimit      time.Duration
	CaptureStdin       bool
	Title              string
	Environment        []string
	CaptureEnvironment []string
}

func getSize() *pty.Winsize {
	var sz *pty.Winsize

	size := func(f *os.File) func() error {
		return func() (err error) {
			sz, err = pty.GetsizeFull(f)
			return
		}
	}

	_ = utils.FirstSuccess(
		size(os.Stdin),
		size(os.Stdout),
		size(os.Stderr),
		func() (err error) {
			w, err := strconv.Atoi(os.Getenv("COLUMNS"))
			if err != nil {
				return err
			}
			h, err := strconv.Atoi(os.Getenv("LINES"))
			if err != nil {
				return err
			}
			sz = &pty.Winsize{Cols: uint16(w), Rows: uint16(h)}
			return nil
		},
		func() error {
			sz = &pty.Winsize{Cols: 80, Rows: 24}
			return nil
		},
	)

	return sz
}

func (r *CastRecorder) Record(w io.Writer) error {
	sz := getSize()
	env, envMap := r.environment()

	ppty, tty, err := pty.Open()
	if err != nil {
		return err
	}

	if err := pty.Setsize(ppty, sz); err != nil {
		_ = tty.Close()
		_ = ppty.Close()
		return err
	}

	winCh := make(chan os.Signal, 1)
	defer close(winCh)

	signal.Notify(winCh, syscall.SIGWINCH)
	defer signal.Stop(winCh)

	go func() {
		for range winCh {
			sz = getSize()
			_ = pty.Setsize(ppty, sz)
		}
	}()

	cmd := r.command(env, tty)
	now := time.Now()
	if err := cmd.Start(); err != nil {
		_ = tty.Close()
		_ = ppty.Close()
		return err
	}

	_ = tty.Close()

	enc := json.NewEncoder(w)

	if err := enc.Encode(r.header(sz, now, envMap)); err != nil {
		_ = cmd.Process.Kill()
		_ = ppty.Close()
		return err
	}

	ch := make(chan *Message)
	defer close(ch)
	go forwardMessages(ch, enc)

	if state, err := terminal.MakeRaw(int(os.Stdin.Fd())); err == nil {
		state := state
		defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), state) }()
	}

	stdinDone := make(chan struct{})
	go r.forwardStdin(ppty, &writer{now, StdinMessage, ch}, stdinDone)
	defer func() { <-stdinDone }()

	stdoutDone := make(chan struct{})
	go r.forwardStdout(ppty, &writer{now, StdoutMessage, ch}, stdoutDone)
	defer func() { <-stdoutDone }()

	defer func() { _ = ppty.Close() }()

	return cmd.Wait()
}

func (r *CastRecorder) environment() ([]string, map[string]string) {
	env := r.Environment
	if env == nil {
		env = os.Environ()
	}

	var captureEnv map[string]string
	if l := len(r.CaptureEnvironment); l > 0 {
		captureEnv = make(map[string]string, l)
		envMap := make(map[string]string, len(env))
		for _, v := range env {
			s := strings.SplitN(v, "=", 2)
			envMap[s[0]] = strings.Join(s[1:], "")
		}
		for _, k := range r.CaptureEnvironment {
			captureEnv[k] = envMap[k]
		}
	}

	return env, captureEnv
}

func (r *CastRecorder) header(sz *pty.Winsize, now time.Time, envMap map[string]string) Header {
	return Header{
		Version:       2,
		Width:         int(sz.Cols),
		Height:        int(sz.Rows),
		Timestamp:     Time{Time: now},
		IdleTimeLimit: Duration(r.IdleTimeLimit),
		Environment:   envMap,
		Title:         r.Title,
	}
}

func (r *CastRecorder) command(env []string, tty *os.File) *exec.Cmd {
	var args []string
	if r.Command != "" {
		args = []string{"-c", r.Command}
	}
	cmd := exec.Command(utils.Shell(), args...)
	cmd.Env = env
	cmd.Stdout = tty
	cmd.Stderr = tty
	cmd.Stdin = tty
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setctty: true,
		Setsid:  true,
		Ctty:    int(tty.Fd()),
	}
	return cmd
}

func (r *CastRecorder) forwardStdin(pty *os.File, w io.Writer, done chan<- struct{}) {
	var stdin io.Reader
	if r.CaptureStdin {
		stdin = io.TeeReader(os.Stdin, w)
		defer close(done)
	} else {
		stdin = os.Stdin
		close(done)
	}
	_, _ = io.Copy(pty, stdin)
}

func (r *CastRecorder) forwardStdout(pty *os.File, w io.Writer, done chan<- struct{}) {
	defer close(done)
	stdout := io.MultiWriter(w, os.Stdout)
	_, _ = io.Copy(stdout, pty)
	_ = pty.Close()
}

func forwardMessages(ch <-chan *Message, enc *json.Encoder) {
	for msg := range ch {
		if msg != nil {
			_ = enc.Encode(msg)
		}
	}
}
