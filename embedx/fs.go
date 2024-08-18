package embedx

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	initTime      = time.Now()
	unixEpochTime = time.Unix(0, 0)
)

// isZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}

type ioFS struct {
	fsys fs.FS
}

type ioFile struct {
	file fs.File
}

type ioFileInfo struct {
	fs.FileInfo
}

// ModTime change modification time to the mod time of the executing binary, falling back on the init time
func (f ioFileInfo) ModTime() time.Time {
	if t := f.FileInfo.ModTime(); !isZeroTime(t) {
		return t
	}
	if exe, err := os.Executable(); err == nil {
		if exe, err := filepath.EvalSymlinks(exe); err == nil {
			if d, err := os.Stat(exe); err == nil {
				if t := d.ModTime(); !isZeroTime(t) {
					return t
				}
			}
		}
	}
	return initTime
}

func (f ioFS) Open(name string) (http.File, error) {
	if name == "/" {
		name = "."
	} else {
		name = strings.TrimPrefix(name, "/")
	}
	file, err := f.fsys.Open(name)
	if err != nil {
		return nil, mapOpenError(err, name, '/', func(path string) (fs.FileInfo, error) {
			return fs.Stat(f.fsys, path)
		})
	}
	return ioFile{file}, nil
}

func (f ioFile) Close() error               { return f.file.Close() }
func (f ioFile) Read(b []byte) (int, error) { return f.file.Read(b) }
func (f ioFile) Stat() (fs.FileInfo, error) {
	d, err := f.file.Stat()
	return ioFileInfo{d}, err
}

var errMissingSeek = errors.New("io.File missing Seek method")
var errMissingReadDir = errors.New("io.File directory missing ReadDir method")

func (f ioFile) Seek(offset int64, whence int) (int64, error) {
	s, ok := f.file.(io.Seeker)
	if !ok {
		return 0, errMissingSeek
	}
	return s.Seek(offset, whence)
}

func (f ioFile) ReadDir(count int) ([]fs.DirEntry, error) {
	d, ok := f.file.(fs.ReadDirFile)
	if !ok {
		return nil, errMissingReadDir
	}
	return d.ReadDir(count)
}

func (f ioFile) Readdir(count int) ([]fs.FileInfo, error) {
	d, ok := f.file.(fs.ReadDirFile)
	if !ok {
		return nil, errMissingReadDir
	}
	var list []fs.FileInfo
	for {
		dirs, err := d.ReadDir(count - len(list))
		for _, dir := range dirs {
			info, err := dir.Info()
			if err != nil {
				// Pretend it doesn't exist, like (*os.File).Readdir does.
				continue
			}
			list = append(list, info)
		}
		if err != nil {
			return list, err
		}
		if count < 0 || len(list) >= count {
			break
		}
	}
	return list, nil
}

// FS converts fsys to a [FileSystem] implementation,
// for use with [FileServer] and [NewFileTransport].
// The files provided by fsys must implement [io.Seeker].
// This custom FS implement solves the caching problem
// when serving embedded files. The idea comes from
// https://www.reddit.com/r/golang/comments/118pqb2/comment/j9io84v/
func FS(fsys fs.FS) http.FileSystem {
	return ioFS{fsys}
}

// mapOpenError maps the provided non-nil error from opening name
// to a possibly better non-nil error. In particular, it turns OS-specific errors
// about opening files in non-directories into fs.ErrNotExist. See Issues 18984 and 49552.
func mapOpenError(originalErr error, name string, sep rune, stat func(string) (fs.FileInfo, error)) error {
	if errors.Is(originalErr, fs.ErrNotExist) || errors.Is(originalErr, fs.ErrPermission) {
		return originalErr
	}

	parts := strings.Split(name, string(sep))
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		fi, err := stat(strings.Join(parts[:i+1], string(sep)))
		if err != nil {
			return originalErr
		}
		if !fi.IsDir() {
			return fs.ErrNotExist
		}
	}
	return originalErr
}
