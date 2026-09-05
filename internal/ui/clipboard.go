package ui

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/muesli/termenv"
)

// clipboardTools are tried in order; the first one present on the system wins.
var clipboardTools = [][]string{
	{"wl-copy"},
	{"xclip", "-selection", "clipboard"},
	{"xsel", "--clipboard", "--input"},
	{"pbcopy"},
}

// copyToClipboard writes text to the system clipboard, falling back to the
// terminal's own OSC 52 copy so the feature still works over SSH where no
// clipboard helper is installed.
func copyToClipboard(text string) error {
	if text == "" {
		return errors.New("nothing to copy")
	}
	for _, tool := range clipboardTools {
		if _, err := exec.LookPath(tool[0]); err != nil {
			continue
		}
		cmd := exec.Command(tool[0], tool[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			return err
		}
		return nil
	}

	if !termenv.NewOutput(os.Stdout).EnvNoColor() {
		termenv.NewOutput(os.Stdout).Copy(text)
		return nil
	}
	return errors.New("no clipboard tool found (install wl-clipboard, xclip or xsel)")
}
