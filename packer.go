package packer

import (
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jakobvarmose/packer/internal/zip"
)

func OpenRoot(path string) (http.FileSystem, error) {
	f, err := os.Open(os.Args[0])
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	size := stat.Size()

	zipsize, err := zip.ZipSize(f, size)
	if err != nil {
		if err == zip.ErrFormat {
			return http.Dir(path), nil
		}
		return nil, err
	}

	r := &readerAt{f, size - zipsize}

	z, err := zip.NewReader(r, zipsize)
	if err != nil {
		return nil, err
	}

	return &fileSystem{z}, nil
}

type readerAt struct {
	source io.ReaderAt
	start  int64
}

func (r *readerAt) ReadAt(p []byte, off int64) (int, error) {
	return r.source.ReadAt(p, r.start+off)
}

type fileSystem struct {
	zip *zip.Reader
}

func (fs *fileSystem) Open(name string) (http.File, error) {
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}

	if name == "" {
		return &file{
			zip:      fs.zip,
			file:     nil,
			filename: "",
			rc:       nil,
			index:    0,
		}, nil
	}

	for _, f := range fs.zip.File {
		if f.Name == name || f.Name == name+"/" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			return &file{
				zip:      fs.zip,
				file:     f,
				filename: f.Name,
				rc:       rc,
				index:    0,
			}, nil
		}
	}
	return nil, os.ErrNotExist
}

type file struct {
	zip      *zip.Reader
	file     *zip.File
	filename string
	rc       io.ReadCloser
	index    int64
	index2   int
}

func (f *file) Close() error {
	if f.rc == nil {
		return nil
	}
	return f.rc.Close()
}

func (f *file) Read(p []byte) (int, error) {
	n, err := f.rc.Read(p)
	f.index += int64(n)
	return n, err
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		offset = f.index + offset
		break
	case io.SeekEnd:
		offset = f.file.FileInfo().Size() + offset
		break
	}

	f.rc.Close()
	f.rc = nil
	f.index = 0
	rc, err := f.file.Open()
	if err != nil {
		return 0, err
	}
	for i := int64(0); i < offset; i++ {
		_, err = rc.Read([]byte{0})
		if err != nil {
			return 0, err
		}
	}
	f.rc = rc
	f.index = offset
	return offset, nil
}

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	if f.index2 >= len(f.zip.File) {
		return nil, io.EOF
	}
	var files []os.FileInfo
	for ; f.index2 < len(f.zip.File); f.index2++ {
		f2 := f.zip.File[f.index2]
		if strings.HasPrefix(f2.Name, f.filename) && len(f2.Name) > len(f.filename) && !strings.Contains(f2.Name[len(f.filename):len(f2.Name)-1], "/") {
			files = append(files, f2.FileInfo())
		}
		if count > 0 && len(files) >= count {
			break
		}
	}
	return files, nil
}

type rootFileInfo struct{}

func (r rootFileInfo) IsDir() bool {
	return true
}

func (r rootFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (r rootFileInfo) Mode() os.FileMode {
	return os.ModeDir
}

func (r rootFileInfo) Name() string {
	return ""
}

func (r rootFileInfo) Size() int64 {
	return 0
}

func (r rootFileInfo) Sys() interface{} {
	return nil
}

func (f *file) Stat() (os.FileInfo, error) {
	if f.file == nil {
		return rootFileInfo{}, nil
	}
	return f.file.FileInfo(), nil
}
