package ui

import (
	"io"

	"github.com/schollz/progressbar/v3"
)

func NewBar(total int64, desc string) io.Writer {
	if total <= 0 {
		total = -1
	}
	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(desc),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(28),
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(),
	)
	return bar
}
