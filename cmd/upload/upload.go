package upload

import (
	"github.com/spf13/cobra"

	"github.com/demosdemon/termshare/cmd"
)

var root = cobra.Command{
	Use:   "upload FILENAME",
	Short: "upload locally saved terminal session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	cmd.AddCommand(&root)
}
