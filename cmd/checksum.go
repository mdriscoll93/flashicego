package cmd

import (
	"fmt"
	"os"

	"flashicego/pkg/checksum"
	"flashicego/pkg/ui"

	"github.com/spf13/cobra"
)

func newChecksumCmd() *cobra.Command {
	var expected string

	cmd := &cobra.Command{
		Use:   "checksum <image-path>",
		Short: "Calculate or verify SHA256 checksum",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			st, err := os.Stat(path)
			if err != nil {
				return err
			}

			bar := ui.NewBar(st.Size(), "checksum")
			actual, err := checksum.SHA256File(path, bar)
			fmt.Println()
			if err != nil {
				return err
			}

			if expected == "" {
				fmt.Println(actual)
				return nil
			}
			if err := checksum.VerifySHA256(path, expected, nil); err != nil {
				return err
			}
			fmt.Printf("checksum verified: %s\n", actual)
			return nil
		},
	}
	cmd.Flags().StringVar(&expected, "expected", "", "expected SHA256 hash")
	return cmd
}
