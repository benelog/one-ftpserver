package oneftpserver

import (
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/spf13/afero"
)

// loggingFs writes down what a client does with the files it is allowed to
// reach: transfers, listings, deletions, renames and new directories. The
// operations that carry no intent of their own are left out, because a single
// listing stats every entry it returns and the log would say nothing over the
// noise.
//
// The paths it logs are the ones the client used. The file system underneath is
// rooted at the home directory, so there is nothing of the server layout to
// give away.
type loggingFs struct {
	afero.Fs

	logger *slog.Logger
}

// OpenFile is how a transfer starts, in either direction: the library asks for
// a read only handle to send a file and a writable one to receive it.
func (l *loggingFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	action := "download"
	if flag&(os.O_WRONLY|os.O_RDWR) != 0 {
		action = "upload"
	}

	file, err := l.Fs.OpenFile(name, flag, perm)
	if err != nil {
		l.logger.Warn(action+" refused", "path", name, "err", err)

		return nil, err
	}

	l.logger.Info(action, "path", name)

	return &countingFile{File: file, logger: l.logger, action: action, path: name}, nil
}

// Open is a listing: the library opens a directory to read its entries, and
// goes through OpenFile for everything it transfers.
func (l *loggingFs) Open(name string) (afero.File, error) {
	file, err := l.Fs.Open(name)
	if err != nil {
		l.logger.Warn("list refused", "path", name, "err", err)

		return nil, err
	}

	l.logger.Info("list", "path", name)

	return file, nil
}

func (l *loggingFs) Remove(name string) error {
	return l.report("delete", name, l.Fs.Remove(name))
}

func (l *loggingFs) RemoveAll(path string) error {
	return l.report("delete", path, l.Fs.RemoveAll(path))
}

func (l *loggingFs) Mkdir(name string, perm os.FileMode) error {
	return l.report("mkdir", name, l.Fs.Mkdir(name, perm))
}

func (l *loggingFs) MkdirAll(path string, perm os.FileMode) error {
	return l.report("mkdir", path, l.Fs.MkdirAll(path, perm))
}

func (l *loggingFs) Rename(oldname, newname string) error {
	err := l.Fs.Rename(oldname, newname)
	if err != nil {
		l.logger.Warn("rename refused", "path", oldname, "to", newname, "err", err)

		return err
	}

	l.logger.Info("rename", "path", oldname, "to", newname)

	return nil
}

// report logs how an operation went and hands its error back untouched.
func (l *loggingFs) report(action, path string, err error) error {
	if err != nil {
		l.logger.Warn(action+" refused", "path", path, "err", err)

		return err
	}

	l.logger.Info(action, "path", path)

	return nil
}

// countingFile counts what actually moved and logs it once the transfer is
// over, which is where the size of an upload or a download comes from. A
// transfer that is cut short still gets its line, with what got through.
type countingFile struct {
	afero.File

	logger *slog.Logger
	action string
	path   string
	// bytes is read by Close, which an aborted transfer can reach while the
	// data is still moving, so it is counted atomically.
	bytes atomic.Int64
}

func (f *countingFile) Read(dst []byte) (int, error) {
	read, err := f.File.Read(dst)
	f.bytes.Add(int64(read))

	return read, err
}

func (f *countingFile) Write(src []byte) (int, error) {
	written, err := f.File.Write(src)
	f.bytes.Add(int64(written))

	return written, err
}

func (f *countingFile) Close() error {
	err := f.File.Close()

	f.logger.Info(f.action+" done", "path", f.path, "bytes", f.bytes.Load())

	return err
}
