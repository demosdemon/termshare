package auth

import (
	"github.com/spf13/cobra"

	"github.com/demosdemon/termshare/cmd"
)

var root = cobra.Command{
	Use:   "auth",
	Short: "manage termshare recordings account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	cmd.AddCommand(&root)
}
