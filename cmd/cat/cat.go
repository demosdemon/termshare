package cat

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/demosdemon/termshare/cmd"
	"github.com/demosdemon/termshare/pkg/termshare"
)

var root = cobra.Command{
	Use:   "cat FILENAME",
	Short: "print full output of terminal session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		fp, err := os.Open(filename)
		if err != nil {
			return errors.Wrapf(err, "error opening file %q", filename)
		}

		player, err := termshare.NewPlayer(fp, 1)
		if err != nil {
			return errors.Wrap(err, "error creating player")
		}

		return player.Play(context.Background(), cmd.OutOrStdout(), 1.0)
	},
}

func init() {
	cmd.AddCommand(&root)
}
