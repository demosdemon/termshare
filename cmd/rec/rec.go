package rec

import (
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"

	"github.com/demosdemon/termshare/cmd"
	"github.com/demosdemon/termshare/pkg/termshare"
)

var (
	flagStdin         bool
	flagAppend        bool
	flagRaw           bool
	flagOverwrite     bool
	flagCommand       string
	flagEnv           []string
	flagTitle         string
	flagIdleTimeLimit time.Duration
	//flagYes           bool
	//flagQuiet         bool
)

var root = cobra.Command{
	Use:   "rec [filename]",
	Short: "record terminal session",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(x *cobra.Command, args []string) error {
		var (
			fp  *os.File
			err error
		)

		if len(args) > 0 {
			flag := os.O_WRONLY | os.O_CREATE
			if flagAppend {
				flag |= os.O_APPEND
			} else if flagOverwrite {
				flag |= os.O_TRUNC
			} else {
				flag |= os.O_EXCL
			}
			fp, err = os.OpenFile(args[0], flag, 0666)
		} else {
			fp, err = ioutil.TempFile("", "termcast-*.cast")
		}

		if err != nil {
			return err
		}

		if flagAppend {
			x.Println(cmd.Info("appending to asciicast at %s", fp.Name()))
		} else {
			x.Println(cmd.Info("recording asciicast to %s", fp.Name()))
		}

		if flagCommand == "" {
			x.Println(cmd.Info(`press <ctrl-d> or type "exit" when you're done'`))
		} else {
			x.Println(cmd.Info("exit opened program when you're done"))
		}

		defer func() { _ = fp.Close() }()

		var rec termshare.Recorder

		if flagRaw {
			rec = &termshare.RawRecorder{
				Command: flagCommand,
			}
		} else {
			rec = &termshare.CastRecorder{
				Command:            flagCommand,
				IdleTimeLimit:      flagIdleTimeLimit,
				CaptureStdin:       flagStdin,
				Title:              flagTitle,
				Environment:        nil,
				CaptureEnvironment: flagEnv,
			}
		}

		err = rec.Record(fp)
		if exitErr, ok := err.(*exec.ExitError); ok {
			x.Println(cmd.Warning("process %s", exitErr.String()))
			err = nil
		}

		if err != nil {
			x.PrintErrln(cmd.Error("recording error: %v", err))
			os.Exit(1)
		}

		x.Println(cmd.Info("recording saved to %s", fp.Name()))
		return nil
	},
}

func init() {
	flags := root.Flags()
	flags.BoolVar(&flagStdin, "stdin", false, "enable stdin recording")
	flags.BoolVar(&flagAppend, "append", false, "append to existing recording")
	flags.BoolVar(&flagRaw, "raw", false, "save only the raw stdout output")
	flags.BoolVar(&flagOverwrite, "overwrite", false, "overwrite the file if it already exists")
	flags.StringVarP(&flagCommand, "command", "c", "", "command to record (default $SHELL)")
	flags.StringSliceVarP(&flagEnv, "env", "e", []string{"`SHELL`", "TERM"}, "list of environment variables to capture")
	flags.StringVarP(&flagTitle, "title", "t", "", "title of the termshare")
	flags.DurationVarP(&flagIdleTimeLimit, "idle-time-limit", "i", 0, "limit recorded idle time to a given number of seconds")
	//flags.BoolVarP(&flagYes, "yes", "y", false, "answer \"yes\" to all prompts")
	//flags.BoolVarP(&flagQuiet, "quiet", "q", false, "be quiet, suppress all notices/warnings (implies --yes)")

	cmd.AddCommand(&root)
}
