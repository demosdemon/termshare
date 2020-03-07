package termshare

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Player struct {
	scanner *bufio.Scanner
	header  Header

	pausedMu sync.RWMutex
	pausedCh chan struct{}
}

func NewPlayer(input io.Reader, idleTimeLimit time.Duration) (*Player, error) {
	r := new(Player)
	r.scanner = bufio.NewScanner(input)
	if !r.scanner.Scan() {
		err := r.scanner.Err()
		if err == nil {
			err = io.ErrUnexpectedEOF
		}
		return nil, errors.Wrap(err, "error looking for header")
	}
	if err := json.Unmarshal(r.scanner.Bytes(), &r.header); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling header")
	}
	if idleTimeLimit > 0 {
		r.header.IdleTimeLimit = Duration(idleTimeLimit)
	}
	return r, nil
}

func (r *Player) Play(ctx context.Context, w io.Writer, speed float64) error {
	t := time.NewTimer(0)
	defer t.Stop()

	ws, ok := w.(io.StringWriter)
	if !ok {
		ws = &stringWriter{w}
	}

	p := Processors{
		ToRelativeTime(),
		CapRelativeTime(r.header.IdleTimeLimit),
		ToAbsoluteTime(),
		AdjustSpeed(speed),
	}

	start := time.Now()
	var msg Message
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-r.paused():
		}

		if msg.MessageType == StdoutMessage {
			_, _ = ws.WriteString(msg.Message)
		}

		if r.scanner.Scan() {
			if err := msg.UnmarshalJSON(r.scanner.Bytes()); err != nil {
				return errors.Wrap(err, "error unmarshaling message")
			}
		} else {
			return errors.Wrap(r.scanner.Err(), "error reading input")
		}

		msg, _ = p.Apply(msg)

		d := msg.Offset
		d -= Duration(time.Since(start))
		if i := r.header.IdleTimeLimit; i > 0 && d > i {
			d = i
		}
		t.Reset(time.Duration(d))
	}
}

func (r *Player) Next() {
	r.pausedMu.Lock()
	r.unpauseLocked()
	r.pauseLocked()
	r.pausedMu.Unlock()
}

func (r *Player) Pause() {
	r.pausedMu.Lock()
	r.pauseLocked()
	r.pausedMu.Unlock()
}

func (r *Player) Unpause() {
	r.pausedMu.Lock()
	r.unpauseLocked()
	r.pausedMu.Unlock()
}

func (r *Player) paused() <-chan struct{} {
	r.pausedMu.RLock()
	ch := r.pausedCh
	r.pausedMu.RUnlock()

	if ch == nil {
		return preClosedChannel
	}

	return ch
}

func (r *Player) pauseLocked() {
	select {
	case <-r.pausedCh:
		r.pausedCh = nil
	default:
	}

	if r.pausedCh == nil {
		r.pausedCh = make(chan struct{})
	}
}

func (r *Player) unpauseLocked() {
	if r.pausedCh == nil {
		return
	}

	select {
	case <-r.pausedCh:
		return
	default:
		close(r.pausedCh)
	}
}

var preClosedChannel = func() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}()

type stringWriter struct {
	w io.Writer
}

func (w *stringWriter) WriteString(s string) (int, error) {
	return w.w.Write([]byte(s))
}
