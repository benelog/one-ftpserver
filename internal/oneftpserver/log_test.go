package oneftpserver

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTheLogFileCarriesTheDayItCovers(t *testing.T) {
	day := time.Date(2026, time.August, 16, 23, 59, 0, 0, time.Local)

	cases := []struct {
		base string
		want string
	}{
		{base: "one-ftpserver.log", want: "one-ftpserver-2026-08-16.log"},
		{base: filepath.Join("var", "ftp.log"), want: filepath.Join("var", "ftp-2026-08-16.log")},
		{base: "activity", want: "activity-2026-08-16"},
	}

	for _, testCase := range cases {
		if got := dailyPath(testCase.base, day); got != testCase.want {
			t.Errorf("dailyPath(%q) = %q, want %q", testCase.base, got, testCase.want)
		}
	}
}

func TestTheLogMovesOnToTheFileOfTheNextDay(t *testing.T) {
	base := filepath.Join(t.TempDir(), "one-ftpserver.log")
	day := time.Date(2026, time.August, 16, 23, 59, 0, 0, time.Local)

	writer, err := newLogWriter(base, func() time.Time { return day })
	if err != nil {
		t.Fatalf("newLogWriter failed: %v", err)
	}

	defer func() { _ = writer.Close() }()

	if _, err := writer.Write([]byte("before midnight\n")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	day = day.Add(2 * time.Minute)

	if _, err := writer.Write([]byte("after midnight\n")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if got := readFile(t, dailyPath(base, day.Add(-2*time.Minute))); got != "before midnight\n" {
		t.Errorf("the file of the 16th holds %q, want the line written that day", got)
	}

	if got := readFile(t, dailyPath(base, day)); got != "after midnight\n" {
		t.Errorf("the file of the 17th holds %q, want the line written that day", got)
	}
}

func TestTheLogFileIsAppendedToRatherThanTruncated(t *testing.T) {
	base := filepath.Join(t.TempDir(), "one-ftpserver.log")
	day := time.Date(2026, time.August, 16, 10, 0, 0, 0, time.Local)
	clock := func() time.Time { return day }

	for _, line := range []string{"first run\n", "second run\n"} {
		writer, err := newLogWriter(base, clock)
		if err != nil {
			t.Fatalf("newLogWriter failed: %v", err)
		}

		if _, err := writer.Write([]byte(line)); err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		if err := writer.Close(); err != nil {
			t.Fatalf("Close failed: %v", err)
		}
	}

	if got := readFile(t, dailyPath(base, day)); got != "first run\nsecond run\n" {
		t.Errorf("the file holds %q, want both runs; a restart must not lose the earlier one", got)
	}
}

func TestALogFileThatCannotBeWrittenIsRefusedAtStartup(t *testing.T) {
	config := &Config{
		Port: 0,
		ID:   AnonymousID,
		Home: t.TempDir(),
		Log:  filepath.Join(t.TempDir(), "no-such-directory", "one-ftpserver.log"),
	}

	if err := config.Prepare(); err == nil {
		t.Error("a log file that cannot be written must be refused before the server starts")
	}
}

func TestTheLogFileCanBeTurnedOff(t *testing.T) {
	home := t.TempDir()
	config := &Config{Port: 0, ID: AnonymousID, Home: home, Log: "off"}

	if err := config.Prepare(); err != nil {
		t.Fatalf("Prepare failed: %v", err)
	}

	if config.Logging() {
		t.Error("--log=off must leave no log file")
	}

	if config.LogFile() != "" {
		t.Errorf("the summary reports %q as the log file, want none", config.LogFile())
	}
}

func TestTheActivityOfAClientReachesTheLogFile(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, "shared.txt"), "my works")

	config := &Config{
		ID:       "benelog",
		Password: "1234",
		Home:     home,
		Log:      filepath.Join(t.TempDir(), "one-ftpserver.log"),
	}
	summary := startServer(t, config)

	client := dial(t, summary.Port, false)
	client.login("benelog", "1234")

	if got := client.retrieve("shared.txt"); got != "my works" {
		t.Fatalf("downloaded %q, want %q", got, "my works")
	}

	client.store("uploaded.txt", "sent over")

	logged := readFile(t, summary.Log)
	for _, want := range []string{"login", "download", "shared.txt", "upload", "uploaded.txt", "bytes=8"} {
		if !strings.Contains(logged, want) {
			t.Errorf("the log does not mention %q:\n%s", want, logged)
		}
	}
}

func TestARefusedLoginIsLoggedWithoutItsPassword(t *testing.T) {
	config := &Config{
		ID:       "benelog",
		Password: "1234",
		Home:     t.TempDir(),
		Log:      filepath.Join(t.TempDir(), "one-ftpserver.log"),
	}
	summary := startServer(t, config)

	client := dial(t, summary.Port, false)
	client.send("USER benelog")
	client.expect(331)
	client.send("PASS s3cret-typo")
	client.reply()

	logged := readFile(t, summary.Log)
	if !strings.Contains(logged, "login refused") {
		t.Errorf("a refused login must be logged:\n%s", logged)
	}

	if strings.Contains(logged, "s3cret-typo") {
		t.Errorf("a password must never be written to the log:\n%s", logged)
	}
}

// discardLogger is the logger of a test that has nothing to say about the log.
func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path) //nolint:gosec // the path is built by the test
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}

	return string(content)
}
