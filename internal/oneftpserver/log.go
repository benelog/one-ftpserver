package oneftpserver

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Defaults of the logging settings.
const (
	// DefaultLogFile is the file the activity is written to when --log is left
	// alone. The date of the day is added to the name, so a server that runs
	// for weeks leaves one file per day instead of one that never stops
	// growing.
	DefaultLogFile = "one-ftpserver.log"

	// LogOff is the --log value that asks for no log file at all.
	LogOff = "off"
)

// logTimeFormat drops the sub second precision and the time zone offset of the
// default one, which are noise in a file that covers a single day.
const logTimeFormat = "2006-01-02T15:04:05"

// newLogger builds the logger the server writes its activity with, and the
// closer that gives the log file back. Both sinks are optional and independent:
// --log writes a file, --console writes to the terminal, and asking for neither
// leaves a logger that drops everything.
func newLogger(config *Config) (*slog.Logger, io.Closer, error) {
	var (
		sinks  []io.Writer
		closer io.Closer = nopCloser{}
	)

	if config.Logging() {
		writer, err := newLogWriter(config.Log, time.Now)
		if err != nil {
			return nil, nil, err
		}

		sinks, closer = append(sinks, writer), writer
	}

	if config.Console {
		// The activity is diagnostic output, so it goes to the standard error
		// stream. That leaves the standard output stream to the summary alone,
		// which --json needs: it has to stay one object a script can parse.
		sinks = append(sinks, os.Stderr)
	}

	if len(sinks) == 0 {
		return slog.New(slog.DiscardHandler), closer, nil
	}

	handler := slog.NewTextHandler(io.MultiWriter(sinks...), &slog.HandlerOptions{ReplaceAttr: shortenTime})

	return slog.New(handler), closer, nil
}

// shortenTime writes the timestamp of a line the way logTimeFormat spells it.
func shortenTime(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) == 0 && attr.Key == slog.TimeKey {
		attr.Value = slog.StringValue(attr.Value.Time().Format(logTimeFormat))
	}

	return attr
}

// logWriter appends to the log file of the current day and moves on to the next
// one by itself, on the first line written after midnight. Nothing is renamed:
// the date is part of the name from the start, so a file that is being written
// to is never moved, which a Windows host would refuse anyway.
type logWriter struct {
	base string
	now  func() time.Time

	mu   sync.Mutex
	file *os.File
	day  string
}

// newLogWriter opens the file of the current day, so a path that cannot be
// written is reported before the server starts rather than on the first client.
func newLogWriter(base string, now func() time.Time) (*logWriter, error) {
	writer := &logWriter{base: base, now: now}

	if err := writer.rotate(now()); err != nil {
		return nil, err
	}

	return writer, nil
}

// Write appends one line, opening the file of the new day when the date has
// changed since the previous one.
func (w *logWriter) Write(line []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if day := w.now(); w.file == nil || day.Format(time.DateOnly) != w.day {
		if err := w.rotate(day); err != nil {
			return 0, err
		}
	}

	return w.file.Write(line)
}

// Close gives the file of the day back. Writing afterwards opens it again.
func (w *logWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	file := w.file
	w.file, w.day = nil, ""

	return file.Close()
}

// rotate opens the file of day and closes the one that was open before it. It
// is called either with the lock held or before the writer is shared.
func (w *logWriter) rotate(day time.Time) error {
	file, err := openLogFile(dailyPath(w.base, day))
	if err != nil {
		return err
	}

	if w.file != nil {
		_ = w.file.Close()
	}

	w.file, w.day = file, day.Format(time.DateOnly)

	return nil
}

// dailyPath puts the date between the name of base and its extension:
// one-ftpserver.log becomes one-ftpserver-2026-08-16.log. The date is the local
// one, which is the day the person reading the file lived through.
func dailyPath(base string, day time.Time) string {
	ext := filepath.Ext(base)

	return strings.TrimSuffix(base, ext) + "-" + day.Format(time.DateOnly) + ext
}

// openLogFile opens a log file for appending, creating it when it is the first
// line of the day. The file records who logged in and which files moved, so it
// is readable by its owner alone.
func openLogFile(path string) (*os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("cannot write the log file %q: %w", path, err)
	}

	return file, nil
}

// nopCloser stands in for the log file when there is none to close.
type nopCloser struct{}

func (nopCloser) Close() error { return nil }
