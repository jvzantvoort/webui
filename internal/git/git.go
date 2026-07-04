// Package git provides read-only access to a local git repository.
package git

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Commit represents a single commit that touched a file.
type Commit struct {
	Hash    string
	Short   string
	Author  string
	Date    time.Time
	Subject string
}

// RelativeDate returns a human-readable relative time string (e.g. "3 hours ago").
func (c Commit) RelativeDate() string {
	return relativeTime(c.Date)
}

// DateFmt returns the commit date formatted for display.
func (c Commit) DateFmt() string {
	if c.Date.IsZero() {
		return ""
	}
	return c.Date.Format("Mon 02 Jan 2006 15:04 MST")
}

// CommitFile stages path in the git index and records a commit with message.
// The file is already saved before this is called; errors here are non-fatal.
func CommitFile(dir, path, message string) error {
	if err := runGit(dir, "add", "--", path); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	if err := runGit(dir, "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

// FileLog returns up to limit commits that touched path, following renames.
// dir is the working directory for the git invocation (typically the repo root).
func FileLog(dir, path string, limit int) ([]Commit, error) {
	// \x1e (record separator) between commits, \x1f (unit separator) between fields.
	args := []string{
		"log", "--follow",
		fmt.Sprintf("-n%d", limit),
		"--pretty=format:\x1e%H\x1f%h\x1f%an\x1f%ai\x1f%s",
		"--", path,
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	return parseLog(string(out))
}

func parseLog(raw string) ([]Commit, error) {
	var commits []Commit
	for _, record := range strings.Split(raw, "\x1e") {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}
		// Fields: hash, short, author, date (ISO 8601), subject
		fields := strings.SplitN(record, "\x1f", 5)
		if len(fields) < 5 {
			continue
		}
		t, err := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(fields[3]))
		if err != nil {
			t = time.Time{}
		}
		commits = append(commits, Commit{
			Hash:    fields[0],
			Short:   fields[1],
			Author:  fields[2],
			Date:    t,
			Subject: fields[4],
		})
	}
	return commits, nil
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		n := int(d.Minutes())
		if n == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", n)
	case d < 24*time.Hour:
		n := int(d.Hours())
		if n == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", n)
	case d < 48*time.Hour:
		return "yesterday"
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		n := int(d.Hours() / (24 * 30))
		if n == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", n)
	default:
		n := int(d.Hours() / (24 * 365))
		if n == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", n)
	}
}
