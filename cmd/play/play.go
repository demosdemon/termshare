package play

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/demosdemon/termshare/cmd"
	"github.com/demosdemon/termshare/pkg/termshare"
)

var (
	flagIdleTimeLimit time.Duration
	flagSpeed         float64
)

var root = cobra.Command{
	Use:   "play FILENAME",
	Short: "replay terminal session",
	Args:  cobra.ExactArgs(1),
	RunE: func(x *cobra.Command, args []string) error {
		filename := args[0]
		fp, err := os.Open(filename)
		if err != nil {
			return errors.Wrapf(err, "error opening file %q", filename)
		}

		player, err := termshare.NewPlayer(fp, flagIdleTimeLimit)
		if err != nil {
			return errors.Wrap(err, "error creating player")
		}

		x.Println(cmd.Info("beginning playback of %s", filename))
		if err := player.Play(context.TODO(), x.OutOrStdout(), flagSpeed); err != nil {
			x.PrintErrln(cmd.Error("error during playback: %v", err))
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	flags := root.Flags()
	flags.DurationVarP(&flagIdleTimeLimit, "idle-time-limit", "i", 0, "limit idle time during playback to given duration")
	flags.Float64VarP(&flagSpeed, "speed", "s", 1.0, "playback speedup (can be fractional)")
	cmd.AddCommand(&root)
}
