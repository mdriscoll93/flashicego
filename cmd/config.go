package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage flashicego configuration (reserved)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("config support will land in the next milestone")
			return nil
		},
	}
	return cmd
}
