package device

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ConfirmDestructiveWrite(d Disk, nonInteractive bool) error {
	if nonInteractive {
		return nil
	}

	fmt.Printf("Target: %s (%s, serial=%s, size=%s)\n", d.Path, d.Model, d.Serial, HumanBytes(d.SizeBytes))
	fmt.Printf("Type the exact device path to confirm destructive write: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if strings.TrimSpace(input) != d.Path {
		return fmt.Errorf("confirmation mismatch; expected %s", d.Path)
	}
	return nil
}
