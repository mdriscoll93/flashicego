package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"flashicego/pkg/checksum"
	"flashicego/pkg/cleanup"
	"flashicego/pkg/device"
	"flashicego/pkg/source"
	"flashicego/pkg/ui"
	"flashicego/pkg/verify"
	writerpkg "flashicego/pkg/write"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newBurnCmd() *cobra.Command {
	var expectedSerial string
	var expectedModel string
	var yes bool
	var profileName string
	var storeISO string
	var checksumFile string
	var allowUnsigned bool
	var shredPasses int

	cmd := &cobra.Command{
		Use:   "burn <iso-path-or-url> <device-path>",
		Short: "Download/validate/burn an ISO to removable media",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := device.RequireLinux(); err != nil {
				return err
			}
			if strings.TrimSpace(expectedSerial) == "" && strings.TrimSpace(expectedModel) == "" {
				return fmt.Errorf("at least one identity flag is required: --serial or --model")
			}

			sourceArg := args[0]
			devicePath := args[1]
			resolved, err := resolveSource(sourceArg, storeISO)
			if err != nil {
				return err
			}

			if err := verifyChecksumPolicy(resolved, sourceArg, checksumFile, allowUnsigned); err != nil {
				return err
			}

			target, err := device.FindDiskByPath(devicePath)
			if err != nil {
				return err
			}
			if err := device.EnsureIdentity(target, expectedModel, expectedSerial); err != nil {
				return err
			}
			if err := device.EnsureUnmounted(target); err != nil {
				return err
			}
			if err := device.ConfirmDestructiveWrite(target, yes); err != nil {
				return err
			}

			isoStat, err := os.Stat(resolved.Path)
			if err != nil {
				return err
			}
			if isoStat.Size() > target.SizeBytes {
				return fmt.Errorf("image (%s) is larger than target disk (%s)", device.HumanBytes(isoStat.Size()), device.HumanBytes(target.SizeBytes))
			}

			info := color.New(color.FgCyan)
			info.Printf("Writing %s to %s\n", resolved.Path, target.Path)
			writeBar := ui.NewBar(isoStat.Size(), "write")
			written, err := writerpkg.ImageToDevice(resolved.Path, target.Path, writeBar)
			fmt.Println()
			if err != nil {
				return err
			}
			if written != isoStat.Size() {
				return fmt.Errorf("short write: wrote %d expected %d", written, isoStat.Size())
			}

			profile, err := verify.ResolveProfile(profileName)
			if err != nil {
				return err
			}
			verifyTotal := isoStat.Size()
			if !profile.Full {
				edge := min64(profile.EdgeBytes, isoStat.Size())
				verifyTotal = edge*2 + int64(profile.RandomSamples)*edge
			}
			verifyBar := ui.NewBar(verifyTotal, "verify")
			if err := verify.CompareImageAndDevice(resolved.Path, target.Path, profile, verifyBar); err != nil {
				return err
			}
			fmt.Println()
			color.New(color.FgGreen, color.Bold).Println("burn complete and verification passed")

			if resolved.FromURL && resolved.TempFile && storeISO == "" {
				if err := cleanup.ShredAndDelete(resolved.Path, shredPasses); err != nil {
					return fmt.Errorf("verification passed but temp file cleanup failed: %w", err)
				}
				color.New(color.FgGreen).Println("temporary ISO securely deleted")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&expectedSerial, "serial", "", "expected target disk serial (required in strict ops)")
	cmd.Flags().StringVar(&expectedModel, "model", "", "expected target disk model")
	cmd.Flags().BoolVar(&yes, "yes", false, "skip interactive destructive write confirmation")
	cmd.Flags().StringVar(&profileName, "verify-profile", "quick", "verify profile: quick|thorough|full")
	cmd.Flags().StringVar(&storeISO, "store-iso", "", "path to keep downloaded ISO instead of shredding temp file")
	cmd.Flags().StringVar(&checksumFile, "checksum-file", "", "path to checksum sidecar file")
	cmd.Flags().BoolVar(&allowUnsigned, "allow-unsigned", false, "allow URL burns without validated checksum")
	cmd.Flags().IntVar(&shredPasses, "shred-passes", 3, "number of secure overwrite passes for temporary ISO cleanup")
	return cmd
}

func resolveSource(input, storeISO string) (source.Resolved, error) {
	if !source.IsURL(input) {
		return source.ResolveLocal(input)
	}

	var outPath string
	if storeISO != "" {
		outPath = storeISO
	} else {
		base := filepath.Base(strings.Split(input, "?")[0])
		if base == "." || base == "/" || base == "" {
			base = "image.iso"
		}
		outPath = filepath.Join(os.TempDir(), "flashicego", base)
	}

	total := remoteContentLength(input)
	bar := ui.NewBar(total, "download")
	if _, err := source.DownloadFile(input, outPath, bar); err != nil {
		return source.Resolved{}, err
	}
	fmt.Println()

	abs, err := filepath.Abs(outPath)
	if err != nil {
		return source.Resolved{}, err
	}
	return source.Resolved{Path: abs, FromURL: true, DownloadedURL: input, TempFile: storeISO == ""}, nil
}

func verifyChecksumPolicy(resolved source.Resolved, sourceArg, checksumFile string, allowUnsigned bool) error {
	if checksumFile != "" {
		b, err := os.ReadFile(checksumFile)
		if err != nil {
			return err
		}
		expected, err := checksum.ParseSHA256Sidecar(string(b))
		if err != nil {
			return err
		}
		bar := ui.NewBar(fileSizeOrZero(resolved.Path), "checksum")
		err = checksum.VerifySHA256(resolved.Path, expected, bar)
		fmt.Println()
		return err
	}

	if source.IsURL(sourceArg) {
		sidecarURL := source.SidecarChecksumURL(sourceArg)
		b, err := source.DownloadBytes(sidecarURL)
		if err != nil {
			if allowUnsigned {
				color.New(color.FgYellow).Printf("warning: checksum sidecar unavailable (%v)\n", err)
				return nil
			}
			return fmt.Errorf("checksum sidecar unavailable at %s: %w", sidecarURL, err)
		}
		expected, err := checksum.ParseSHA256Sidecar(string(b))
		if err != nil {
			if allowUnsigned {
				color.New(color.FgYellow).Printf("warning: could not parse checksum sidecar (%v)\n", err)
				return nil
			}
			return err
		}
		bar := ui.NewBar(fileSizeOrZero(resolved.Path), "checksum")
		err = checksum.VerifySHA256(resolved.Path, expected, bar)
		fmt.Println()
		return err
	}

	color.New(color.FgYellow).Println("warning: no checksum verification requested for local image")
	return nil
}

func remoteContentLength(uri string) int64 {
	resp, err := http.Head(uri)
	if err != nil {
		return -1
	}
	defer resp.Body.Close()
	if resp.ContentLength <= 0 {
		return -1
	}
	return resp.ContentLength
}

func fileSizeOrZero(path string) int64 {
	st, err := os.Stat(path)
	if err != nil {
		return -1
	}
	return st.Size()
}
