package cmd

import (
	"fmt"
	"os"

	"flashicego/pkg/ui"
	"flashicego/pkg/verify"

	"github.com/spf13/cobra"
)

func newVerifyCmd() *cobra.Command {
	var profileName string

	cmd := &cobra.Command{
		Use:   "verify <image-path> <device-path>",
		Short: "Verify written media against source image",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := verify.ResolveProfile(profileName)
			if err != nil {
				return err
			}

			st, err := os.Stat(args[0])
			if err != nil {
				return err
			}
			total := st.Size()
			if !profile.Full {
				edge := min64(profile.EdgeBytes, st.Size())
				total = edge*2 + int64(profile.RandomSamples)*edge
			}
			bar := ui.NewBar(total, "verify")
			if err := verify.CompareImageAndDevice(args[0], args[1], profile, bar); err != nil {
				return err
			}
			fmt.Println("\nverify passed")
			return nil
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "quick", "verify profile: quick|thorough|full")
	return cmd
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
