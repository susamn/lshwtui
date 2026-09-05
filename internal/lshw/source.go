package lshw

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Source records where a hardware tree came from, so the UI can say whether
// the data is live, complete, or read from a snapshot on disk.
type Source struct {
	Origin     string // "lshw", "sudo lshw", or a snapshot file path
	Privileged bool   // whether the capture had the rights to see everything
	Snapshot   bool   // true when read from a file rather than the live system
	CapturedAt time.Time
}

// Describe renders the source as a single status-bar phrase.
func (s Source) Describe() string {
	switch {
	case s.Snapshot:
		return "snapshot " + filepath.Base(s.Origin)
	case s.Privileged:
		return "live (root)"
	default:
		return "live (partial)"
	}
}

// IsRoot reports whether this process can see the complete hardware tree.
func IsRoot() bool { return os.Geteuid() == 0 }

// Available reports whether the lshw binary can be found.
func Available() error {
	if _, err := exec.LookPath("lshw"); err != nil {
		return errors.New("lshw not found in PATH — install it (e.g. pacman -S lshw)")
	}
	return nil
}

// Capture runs lshw directly, with whatever privileges this process holds.
func Capture(ctx context.Context) ([]byte, Source, error) {
	src := Source{Origin: "lshw", Privileged: IsRoot(), CapturedAt: time.Now()}
	if err := Available(); err != nil {
		return nil, src, err
	}
	out, err := exec.CommandContext(ctx, "lshw", "-json").Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return nil, src, fmt.Errorf("lshw failed: %s", firstLine(ee.Stderr))
		}
		return nil, src, fmt.Errorf("running lshw: %w", err)
	}
	return out, src, nil
}

// Escalator names the privilege-escalation helper available on this system.
func Escalator() (string, error) {
	for _, c := range []string{"sudo", "doas", "pkexec"} {
		if _, err := exec.LookPath(c); err == nil {
			return c, nil
		}
	}
	return "", errors.New("no sudo, doas or pkexec found to re-run lshw as root")
}

// PrivilegedCapture builds a command that re-runs lshw as root, writing JSON to
// outPath. The caller is expected to hand this to tea.ExecProcess so the
// terminal is released and any password prompt is visible; stdout goes to the
// file precisely so the prompt is not polluted by JSON.
func PrivilegedCapture(outPath string) (*exec.Cmd, Source, error) {
	src := Source{Privileged: true, CapturedAt: time.Now()}
	if err := Available(); err != nil {
		return nil, src, err
	}
	esc, err := Escalator()
	if err != nil {
		return nil, src, err
	}
	src.Origin = esc + " lshw"

	f, err := os.Create(outPath)
	if err != nil {
		return nil, src, fmt.Errorf("creating capture file: %w", err)
	}
	defer f.Close()

	cmd := exec.Command(esc, "lshw", "-json")
	cmd.Stdout = f
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd, src, nil
}

// LoadFile reads a previously saved snapshot.
func LoadFile(path string) ([]byte, Source, error) {
	src := Source{Origin: path, Snapshot: true, Privileged: true}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, src, fmt.Errorf("reading snapshot: %w", err)
	}
	if info, err := os.Stat(path); err == nil {
		src.CapturedAt = info.ModTime()
	}
	return data, src, nil
}

// SaveFile writes the current raw JSON to path, creating parent directories.
func SaveFile(path string, data []byte) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing snapshot: %w", err)
	}
	return nil
}

func firstLine(b []byte) string {
	for i, c := range b {
		if c == '\n' {
			return string(b[:i])
		}
	}
	return string(b)
}
