package cmd

import (
	"fmt"

	"flashicego/pkg/device"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newListDisksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-disks",
		Short: "List candidate removable disks",
		RunE: func(cmd *cobra.Command, args []string) error {
			disks, err := device.ListRemovableDisks()
			if err != nil {
				return err
			}
			if len(disks) == 0 {
				fmt.Println("No removable disks found")
				return nil
			}

			header := color.New(color.FgCyan, color.Bold)
			header.Println("PATH\tSIZE\tMODEL\tSERIAL\tBY-ID\tMOUNTED")
			for _, d := range disks {
				byID := "-"
				if len(d.ByIDPaths) > 0 {
					byID = d.ByIDPaths[0]
				}
				fmt.Printf("%s\t%s\t%s\t%s\t%s\t%t\n", d.Path, device.HumanBytes(d.SizeBytes), d.Model, d.Serial, byID, d.Mounted)
			}
			return nil
		},
	}
	return cmd
}
