package oneftpserver

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// pathHidingFs keeps the real location of the home directory out of the errors
// a client gets to see. The FTP library repeats error messages to the client
// word for word, and the errors of the wrapped file system name files by their
// full path on disk, so without this every failed operation would tell the
// client where on the server the home directory lives.
type pathHidingFs struct {
	afero.Fs

	// home is the absolute path errors must not mention; it is replaced by
	// the path the client used.
	home string
}

func (p *pathHidingFs) Create(name string) (afero.File, error) {
	return p.file(p.Fs.Create(name))
}

func (p *pathHidingFs) Mkdir(name string, perm os.FileMode) error {
	return p.hide(p.Fs.Mkdir(name, perm))
}

func (p *pathHidingFs) MkdirAll(path string, perm os.FileMode) error {
	return p.hide(p.Fs.MkdirAll(path, perm))
}

func (p *pathHidingFs) Open(name string) (afero.File, error) {
	return p.file(p.Fs.Open(name))
}

func (p *pathHidingFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return p.file(p.Fs.OpenFile(name, flag, perm))
}

func (p *pathHidingFs) Remove(name string) error {
	return p.hide(p.Fs.Remove(name))
}

func (p *pathHidingFs) RemoveAll(path string) error {
	return p.hide(p.Fs.RemoveAll(path))
}

func (p *pathHidingFs) Rename(oldname, newname string) error {
	return p.hide(p.Fs.Rename(oldname, newname))
}

func (p *pathHidingFs) Stat(name string) (os.FileInfo, error) {
	info, err := p.Fs.Stat(name)

	return info, p.hide(err)
}

func (p *pathHidingFs) Chmod(name string, mode os.FileMode) error {
	return p.hide(p.Fs.Chmod(name, mode))
}

func (p *pathHidingFs) Chown(name string, uid, gid int) error {
	return p.hide(p.Fs.Chown(name, uid, gid))
}

func (p *pathHidingFs) Chtimes(name string, atime, mtime time.Time) error {
	return p.hide(p.Fs.Chtimes(name, atime, mtime))
}

// file wraps a freshly opened file, whose own errors reach the client through
// failed listings and transfers.
func (p *pathHidingFs) file(f afero.File, err error) (afero.File, error) {
	if err != nil {
		return nil, p.hide(err)
	}

	return &pathHidingFile{File: f, fs: p}, nil
}

// hide rewrites the paths an error carries. The error keeps its type and its
// cause, so os.IsNotExist and the like still recognize it.
func (p *pathHidingFs) hide(err error) error {
	switch pathErr := err.(type) {
	case *os.PathError:
		return &os.PathError{Op: pathErr.Op, Path: p.hidePath(pathErr.Path), Err: pathErr.Err}
	case *os.LinkError:
		return &os.LinkError{Op: pathErr.Op, Old: p.hidePath(pathErr.Old), New: p.hidePath(pathErr.New), Err: pathErr.Err}
	}

	return err
}

func (p *pathHidingFs) hidePath(path string) string {
	if path == p.home {
		return string(os.PathSeparator)
	}

	rel := strings.TrimPrefix(path, p.home)
	if rel == path || !os.IsPathSeparator(rel[0]) {
		return path
	}

	return rel
}

// pathHidingFile cleans the errors of an open file the same way pathHidingFs
// cleans the errors of the file system that opened it.
type pathHidingFile struct {
	afero.File

	fs *pathHidingFs
}

func (f *pathHidingFile) Close() error {
	return f.fs.hide(f.File.Close())
}

func (f *pathHidingFile) Read(b []byte) (int, error) {
	n, err := f.File.Read(b)

	return n, f.fs.hide(err)
}

func (f *pathHidingFile) ReadAt(b []byte, off int64) (int, error) {
	n, err := f.File.ReadAt(b, off)

	return n, f.fs.hide(err)
}

func (f *pathHidingFile) Seek(offset int64, whence int) (int64, error) {
	pos, err := f.File.Seek(offset, whence)

	return pos, f.fs.hide(err)
}

func (f *pathHidingFile) Write(b []byte) (int, error) {
	n, err := f.File.Write(b)

	return n, f.fs.hide(err)
}

func (f *pathHidingFile) WriteAt(b []byte, off int64) (int, error) {
	n, err := f.File.WriteAt(b, off)

	return n, f.fs.hide(err)
}

func (f *pathHidingFile) WriteString(s string) (int, error) {
	n, err := f.File.WriteString(s)

	return n, f.fs.hide(err)
}

func (f *pathHidingFile) Readdir(count int) ([]os.FileInfo, error) {
	infos, err := f.File.Readdir(count)

	return infos, f.fs.hide(err)
}

func (f *pathHidingFile) Readdirnames(n int) ([]string, error) {
	names, err := f.File.Readdirnames(n)

	return names, f.fs.hide(err)
}

func (f *pathHidingFile) Stat() (os.FileInfo, error) {
	info, err := f.File.Stat()

	return info, f.fs.hide(err)
}

func (f *pathHidingFile) Sync() error {
	return f.fs.hide(f.File.Sync())
}

func (f *pathHidingFile) Truncate(size int64) error {
	return f.fs.hide(f.File.Truncate(size))
}
