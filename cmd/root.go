package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "flashicego",
	Short: "Engineer-grade terminal image burner",
	Long:  "flashicego burns ISO images to removable media with strong safety checks.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(newBurnCmd())
	rootCmd.AddCommand(newListDisksCmd())
	rootCmd.AddCommand(newVerifyCmd())
	rootCmd.AddCommand(newChecksumCmd())
	rootCmd.AddCommand(newConfigCmd())
}

func mustGetwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func failf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
