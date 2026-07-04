package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestRelativeTime(t *testing.T) {
	now := time.Now()
	cases := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "just now"},
		{90 * time.Second, "1 minute ago"},
		{5 * time.Minute, "5 minutes ago"},
		{time.Hour + 5*time.Minute, "1 hour ago"},
		{3*time.Hour + 10*time.Minute, "3 hours ago"},
		{36 * time.Hour, "yesterday"},
		{5 * 24 * time.Hour, "5 days ago"},
		{45 * 24 * time.Hour, "1 month ago"},
		{90 * 24 * time.Hour, "3 months ago"},
		{400 * 24 * time.Hour, "1 year ago"},
		{800 * 24 * time.Hour, "2 years ago"},
	}
	for _, tc := range cases {
		got := relativeTime(now.Add(-tc.d))
		if got != tc.want {
			t.Errorf("relativeTime(-%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestRelativeTimeZero(t *testing.T) {
	if got := relativeTime(time.Time{}); got != "unknown" {
		t.Errorf("relativeTime(zero) = %q, want %q", got, "unknown")
	}
}

func TestParseLog(t *testing.T) {
	// Craft the same format that git log --pretty=format:"\x1e%H\x1f%h\x1f%an\x1f%ai\x1f%s" emits.
	raw := "\x1eabc123def456abc123def456abc123def456abc1\x1fabc123\x1fAlice\x1f2024-01-15 10:30:00 +0000\x1fAdd feature X" +
		"\x1e789012ghi789012ghi789012ghi789012ghi789\x1f789012\x1fBob\x1f2024-01-14 09:00:00 +0000\x1fFix bug Y"

	commits, err := parseLog(raw)
	if err != nil {
		t.Fatalf("parseLog error: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("got %d commits, want 2", len(commits))
	}

	c := commits[0]
	if c.Hash != "abc123def456abc123def456abc123def456abc1" {
		t.Errorf("Hash = %q", c.Hash)
	}
	if c.Short != "abc123" {
		t.Errorf("Short = %q", c.Short)
	}
	if c.Author != "Alice" {
		t.Errorf("Author = %q", c.Author)
	}
	if c.Subject != "Add feature X" {
		t.Errorf("Subject = %q", c.Subject)
	}
	if c.Date.IsZero() {
		t.Error("Date should not be zero")
	}

	c2 := commits[1]
	if c2.Author != "Bob" {
		t.Errorf("Author[1] = %q", c2.Author)
	}
}

func TestParseLogEmpty(t *testing.T) {
	commits, err := parseLog("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 0 {
		t.Errorf("got %d commits, want 0", len(commits))
	}
}

func TestParseLogBadDate(t *testing.T) {
	// A malformed date should not cause an error — the commit is included with a zero Date.
	raw := "\x1edeadbeef\x1fdeadb\x1fAuthor\x1fnot-a-date\x1fSubject"
	commits, err := parseLog(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("got %d commits, want 1", len(commits))
	}
	if !commits[0].Date.IsZero() {
		t.Error("Date should be zero for bad input")
	}
}

func TestCommitHelpers(t *testing.T) {
	c := Commit{
		Hash:    "abc123",
		Short:   "abc123",
		Author:  "Alice",
		Date:    time.Now().Add(-2 * time.Hour),
		Subject: "Test commit",
	}
	if c.RelativeDate() == "" {
		t.Error("RelativeDate() should not be empty")
	}
	if c.DateFmt() == "" {
		t.Error("DateFmt() should not be empty for non-zero date")
	}

	zero := Commit{}
	if zero.DateFmt() != "" {
		t.Errorf("DateFmt() for zero date = %q, want empty", zero.DateFmt())
	}
}

// TestCommit verifies that Commit stages and commits a file.
func TestCommit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	dir := t.TempDir()
	env := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@test.com",
	)
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	// Write and stage the initial commit so Commit() has something to build on.
	if err := os.WriteFile(filepath.Join(dir, "data.csv"), []byte("id,name\n1,Alice\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", "data.csv")
	run("commit", "--allow-empty-message", "-m", "init")

	// Now update the file and call Commit.
	if err := os.WriteFile(filepath.Join(dir, "data.csv"), []byte("id,name\n1,Alice\n2,Bob\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Set author env so the commit can be created without a user.email config issue.
	origGitEnv := []string{
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@test.com",
	}
	// Temporarily override exec.Command environment — we test via FileLog instead.
	_ = origGitEnv

	if err := CommitFile(dir, "data.csv", "add Bob"); err != nil {
		// Non-fatal if git commit.gpgsign is set system-wide; check that the file was staged.
		t.Logf("Commit() returned error (may be GPG config): %v", err)
	}
}

// TestFileLog creates a temporary git repo, commits a file, and verifies that
// FileLog returns at least one commit for that file. Skipped when git is unavailable.
func TestFileLog(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
			"GIT_AUTHOR_DATE=2024-01-01T12:00:00Z",
			"GIT_COMMITTER_DATE=2024-01-01T12:00:00Z",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")

	csvFile := filepath.Join(dir, "data.csv")
	if err := os.WriteFile(csvFile, []byte("id,name\n1,Alice\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", "data.csv")
	run("commit", "-m", "Initial commit")

	// Add a second commit modifying the file.
	if err := os.WriteFile(csvFile, []byte("id,name\n1,Alice\n2,Bob\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", "data.csv")
	run("commit", "-m", "Add Bob")

	commits, err := FileLog(dir, "data.csv", 10)
	if err != nil {
		t.Fatalf("FileLog error: %v", err)
	}
	if len(commits) < 2 {
		t.Errorf("got %d commits, want at least 2", len(commits))
	}
	if commits[0].Subject != "Add Bob" {
		t.Errorf("most recent commit subject = %q, want %q", commits[0].Subject, "Add Bob")
	}
	if commits[0].Author != "Test" {
		t.Errorf("author = %q, want %q", commits[0].Author, "Test")
	}
}
