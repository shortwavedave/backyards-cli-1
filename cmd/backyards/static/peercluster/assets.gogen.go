// Code generated by vfsgen; DO NOT EDIT.

package peercluster

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	pathpkg "path"
	"time"
)

// Assets statically implements the virtual filesystem provided to vfsgen.
var Assets = func() http.FileSystem {
	fs := vfsgen۰FS{
		"/": &vfsgen۰DirInfo{
			name:    "/",
			modTime: time.Date(2019, 1, 1, 0, 1, 0, 0, time.UTC),
		},
		"/Chart.yaml": &vfsgen۰CompressedFileInfo{
			name:             "Chart.yaml",
			modTime:          time.Date(2019, 1, 1, 0, 1, 0, 0, time.UTC),
			uncompressedSize: 289,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x74\x8f\x31\x6e\xc3\x30\x0c\x45\x77\x9e\x82\x17\x90\x9d\xac\x9a\x8a\x76\xee\xda\x9d\x51\x68\x85\xa8\x45\x0a\xa2\xec\xa1\xa7\x2f\xe4\x1a\x05\x5a\x20\x23\xdf\x07\xc9\xf7\xa9\xca\x07\x37\x17\xd3\x88\xfb\x15\xa8\xd6\xdf\xf1\x32\x5d\xa6\x2b\xdc\xd9\x53\x93\xda\x0f\xf4\x4e\x2a\x0b\x7b\xc7\xc5\x1a\x56\xe6\x86\x69\xdd\xbc\x73\x73\xfc\xdc\x6e\x9c\x4c\x17\xc9\x98\x59\xb9\xd1\xd8\x80\x87\x15\x8e\xf8\xe8\xbd\x7a\x9c\xe7\x1b\xe9\x17\x49\x5a\x6d\xbb\x4f\xc9\x0a\x48\x1a\x47\x9f\xa4\xb3\x94\x7c\xb2\x70\xc0\xb0\x5a\xb6\xa9\x6a\x86\x42\xa2\x9d\x44\xb9\x79\x84\x80\x5c\x48\xd6\x88\xa2\x8b\xbd\xfc\x7f\x81\xa8\x34\x14\x5e\x0f\x8e\x6f\x23\x80\x1f\x34\xfc\xc3\xe9\x1f\xca\xd9\xcc\x61\xff\xdb\xff\x3b\x00\x00\xff\xff\x77\x80\xd4\x15\x21\x01\x00\x00"),
		},
		"/templates": &vfsgen۰DirInfo{
			name:    "templates",
			modTime: time.Date(2019, 1, 1, 0, 1, 0, 0, time.UTC),
		},
		"/templates/backyards-als-service.yaml": &vfsgen۰CompressedFileInfo{
			name:             "backyards-als-service.yaml",
			modTime:          time.Date(2019, 1, 1, 0, 1, 0, 0, time.UTC),
			uncompressedSize: 372,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x8f\xb1\x6a\xc4\x40\x0c\x44\xfb\xfd\x0a\xfd\x80\x37\xa4\xdd\x3e\x6d\x9a\x40\x7a\x79\x3d\x01\xe1\xb5\x56\xac\x74\x26\xfe\xfb\xe0\x1c\x81\x3b\xc8\xc1\x75\xc3\xf0\xf4\x06\xb1\xc9\x27\x86\x4b\xd7\x42\xfb\x6b\x5a\x45\x97\x42\x1f\x18\xbb\x54\xa4\x0d\xc1\x0b\x07\x97\x44\xd4\x78\x46\xf3\x33\x11\xb1\x59\xa1\x99\xeb\x7a\xf0\x58\x7c\xe2\xe6\x7f\x75\x5e\x2f\x33\x86\x22\xe0\x59\xfa\x4b\xed\x9b\x75\x85\x46\xa1\xc7\x90\xa8\x07\x6b\xc5\x8d\xf2\x01\xa9\xbc\xe1\xb9\x61\xe3\x11\x53\xff\xba\x57\xfe\x7f\x7e\xb6\x6e\x7c\xb7\x3f\xf9\xe1\x81\x2d\xb9\xa1\x9e\x2f\xe3\x3b\x30\x94\xdb\xfb\xaf\x41\x3c\xa4\x4f\x26\xad\x47\xbe\xe6\x2b\x9e\x7d\xaf\xb9\xb6\x8b\x07\x46\x6e\xbd\x72\x4b\x44\x71\x18\x0a\xbd\xdd\x08\xd2\x4f\x00\x00\x00\xff\xff\x79\x95\x5b\x12\x74\x01\x00\x00"),
		},
		"/templates/backyards-tracing-service.yaml": &vfsgen۰CompressedFileInfo{
			name:             "backyards-tracing-service.yaml",
			modTime:          time.Date(2019, 1, 1, 0, 1, 0, 0, time.UTC),
			uncompressedSize: 365,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x8f\xb1\x4e\x2c\x31\x0c\x45\xfb\x7c\x85\x7f\x60\xf2\xf4\xda\xf4\xb4\x34\x48\xf4\xde\xcc\x65\x65\x26\xe3\x58\xb1\x77\xc4\xf0\xf5\x68\x19\x40\x6c\xb1\xa2\xbb\x89\x8f\xcf\x95\xd9\xe4\x19\xc3\xa5\x6b\xa1\xed\x7f\x5a\x44\xe7\x42\x4f\x18\x9b\x54\xa4\x15\xc1\x33\x07\x97\x44\xd4\xf8\x84\xe6\xd7\x44\xc4\x66\x85\x5e\x19\x67\x8c\xef\x77\x5e\x2e\x27\x0c\x45\xc0\xb3\xf4\x7f\xb5\xaf\xd6\x15\x1a\x85\x62\x70\x15\x3d\xdf\x01\x45\x3d\x58\x2b\x0a\x9d\xb8\x2e\x3b\x8f\xd9\xef\x90\xca\x2b\xfe\x68\x35\x1e\x31\xf5\x97\x5b\xd7\xb1\xf7\xf3\x33\xbd\x8b\x2d\xa2\x5f\x03\x37\xbe\xe9\x9e\x7c\xf7\xc0\x9a\xdc\x50\xaf\xb7\xe2\x2d\x30\x94\xdb\xe3\xa7\x44\x3c\xa4\x4f\x26\xad\x47\x3e\xf2\x81\x67\xdf\x6a\xae\xed\xe2\x81\x91\x5b\xaf\xdc\x12\x51\xec\x86\x42\x0f\xbf\x04\xe9\x23\x00\x00\xff\xff\xae\xee\xe6\x32\x6d\x01\x00\x00"),
		},
		"/templates/clusterrole.yaml": &vfsgen۰CompressedFileInfo{
			name:             "clusterrole.yaml",
			modTime:          time.Date(2019, 1, 1, 0, 1, 0, 0, time.UTC),
			uncompressedSize: 1872,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xc4\x54\xb1\x8e\xdb\x30\x0c\xdd\xfd\x15\xc2\x2d\x07\x14\x88\x0f\xdd\x0a\xaf\x1d\xba\x74\x0a\xd0\xee\xb4\xc4\x38\x44\x64\x51\x20\x29\xa7\xed\xd7\x17\xb6\x93\x22\x80\x93\xa0\x68\x13\xdf\x64\x9a\x7c\xf6\x7b\x7a\x22\x09\x99\xbe\xa3\x28\x71\x6a\x9c\xb4\xe0\x6b\x28\xb6\x67\xa1\x5f\x60\xc4\xa9\x3e\x7c\xd2\x9a\xf8\x6d\xf8\x58\x1d\x28\x85\xc6\x7d\x8e\x45\x0d\x65\xcb\x11\xab\x1e\x0d\x02\x18\x34\x95\x73\x09\x7a\x6c\x1c\xa9\x11\x6f\x38\xa3\x80\xb1\x54\x52\x22\x6a\x53\x6d\x1c\x64\xfa\x22\x5c\xb2\x8e\xd0\xf1\x35\x6b\xe5\x9c\xa0\x72\x11\x8f\xa7\x6c\xc0\x1c\xf9\x67\x8f\xc9\xc6\xe2\x80\xd2\x9e\x0a\x1d\xda\xf4\x8c\xa4\x73\x70\x04\xf3\xfb\x29\xf2\x82\x60\x38\x85\x25\x87\x73\x98\xff\xd4\x03\x46\x34\xfc\x07\x05\x6f\x6a\x60\xe5\x86\x90\x05\xd5\xe2\xff\x67\x0b\xea\xc9\x91\x9a\x78\x49\xe6\x39\xed\xa8\x7b\xfe\x51\xff\x5e\xca\x7f\x9e\x19\x42\x4f\x3a\x36\x92\x60\x47\x6a\x72\xd9\x40\x4b\xce\xbe\x18\x18\xa5\xee\x88\xed\x9e\xf9\x30\x4b\x28\xf3\x47\x3a\x21\x06\x88\x14\xee\x62\x9e\x6b\xdc\xcb\xcb\x52\xb5\xa2\x17\x5c\xa1\x3f\xaf\x73\xcb\x40\x1e\xdf\x95\x1c\xbc\xe7\xb2\xc6\x80\xde\x5c\x45\x57\xba\x77\x5e\x49\xc2\xf1\xf1\xde\x8c\x21\xaa\x87\x08\x8f\x13\xd9\x52\x0a\x94\x56\x98\xfc\x6b\xf7\x38\xcf\x50\x0f\x79\x8d\x1d\x4b\xf8\xc3\x30\x8d\x2b\x41\x6f\x1b\x53\xd4\xb8\x3f\x27\x03\xee\x28\xd1\x3a\xf3\xfd\xfa\xe1\x75\x29\x67\x4e\x5e\x10\x8f\x89\x8d\x4b\x9c\xb6\x27\xe0\xb7\xed\xd7\x7b\xd8\xdf\x01\x00\x00\xff\xff\xd8\xd0\xce\x94\x50\x07\x00\x00"),
		},
		"/templates/clusterrolebinding.yaml": &vfsgen۰CompressedFileInfo{
			name:             "clusterrolebinding.yaml",
			modTime:          time.Date(2019, 1, 1, 0, 1, 0, 0, time.UTC),
			uncompressedSize: 288,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x8d\x31\x4e\x03\x31\x10\x45\x7b\x9f\x62\x2e\x10\xa3\x74\xc8\x1d\x50\xd0\x51\x2c\x12\xfd\xac\xf7\x03\x43\x36\x1e\x6b\x66\x9c\x82\x28\x77\x47\x08\xa8\xd0\xa6\x7d\x5f\xef\x3f\xee\xf2\x02\x73\xd1\x56\xc8\x66\xae\x99\x47\xbc\xab\xc9\x27\x87\x68\xcb\x87\x5b\xcf\xa2\x37\xa7\xfd\x8c\xe0\x7d\x3a\x48\x5b\x0a\x3d\xac\xc3\x03\x36\xe9\x8a\x7b\x69\x8b\xb4\xb7\x74\x44\xf0\xc2\xc1\x25\x11\x35\x3e\xa2\x90\x78\x88\xee\xb4\xc3\x38\xd4\x92\xe9\x8a\x09\xaf\xdf\x3b\x77\x79\x34\x1d\xfd\x4a\x30\x11\xfd\x4b\x6d\x3d\xfb\x98\x3f\x50\xc3\x4b\xda\xfd\x4a\xcf\xb0\x93\x54\xdc\xd5\xaa\xa3\xc5\x96\xf7\x83\xbd\x73\x45\xa1\xf3\x99\xf2\x84\x15\xec\xc8\x4f\x7f\x98\x2e\x97\xf4\x15\x00\x00\xff\xff\x84\xc6\x87\x7d\x20\x01\x00\x00"),
		},
		"/templates/namespace.yaml": &vfsgen۰CompressedFileInfo{
			name:             "namespace.yaml",
			modTime:          time.Date(2019, 1, 1, 0, 1, 0, 0, time.UTC),
			uncompressedSize: 144,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\xcc\x31\x0e\xc2\x30\x0c\x05\xd0\xdd\xa7\xf0\x05\x5c\x89\xd5\x87\x60\x60\x60\xff\x34\x7f\x88\x8a\x43\x55\x47\x48\x55\x94\xbb\xb3\xb1\x77\x7f\x7a\xd8\xeb\x93\x47\xd6\x4f\x73\xfd\xde\x64\xab\xad\xb8\xde\x11\xcc\x1d\x2b\x25\xd8\x51\xd0\xe1\xa2\xda\x10\x74\x1d\x43\x97\x07\xdf\x44\x72\xf9\x3b\x9d\x53\xcc\x4c\x2e\x6f\x2f\xac\xdb\x89\xa3\xa4\xe5\x99\x9d\x21\xbf\x00\x00\x00\xff\xff\xbc\xe0\xfe\x9c\x90\x00\x00\x00"),
		},
		"/templates/serviceaccount.yaml": &vfsgen۰FileInfo{
			name:    "serviceaccount.yaml",
			modTime: time.Date(2019, 1, 1, 0, 1, 0, 0, time.UTC),
			content: []byte("\x61\x70\x69\x56\x65\x72\x73\x69\x6f\x6e\x3a\x20\x76\x31\x0a\x6b\x69\x6e\x64\x3a\x20\x53\x65\x72\x76\x69\x63\x65\x41\x63\x63\x6f\x75\x6e\x74\x0a\x6d\x65\x74\x61\x64\x61\x74\x61\x3a\x0a\x20\x20\x6e\x61\x6d\x65\x3a\x20\x69\x73\x74\x69\x6f\x2d\x6f\x70\x65\x72\x61\x74\x6f\x72\x0a\x20\x20\x6e\x61\x6d\x65\x73\x70\x61\x63\x65\x3a\x20\x7b\x7b\x20\x2e\x52\x65\x6c\x65\x61\x73\x65\x2e\x4e\x61\x6d\x65\x73\x70\x61\x63\x65\x20\x7d\x7d\x0a"),
		},
	}
	fs["/"].(*vfsgen۰DirInfo).entries = []os.FileInfo{
		fs["/Chart.yaml"].(os.FileInfo),
		fs["/templates"].(os.FileInfo),
	}
	fs["/templates"].(*vfsgen۰DirInfo).entries = []os.FileInfo{
		fs["/templates/backyards-als-service.yaml"].(os.FileInfo),
		fs["/templates/backyards-tracing-service.yaml"].(os.FileInfo),
		fs["/templates/clusterrole.yaml"].(os.FileInfo),
		fs["/templates/clusterrolebinding.yaml"].(os.FileInfo),
		fs["/templates/namespace.yaml"].(os.FileInfo),
		fs["/templates/serviceaccount.yaml"].(os.FileInfo),
	}

	return fs
}()

type vfsgen۰FS map[string]interface{}

func (fs vfsgen۰FS) Open(path string) (http.File, error) {
	path = pathpkg.Clean("/" + path)
	f, ok := fs[path]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}

	switch f := f.(type) {
	case *vfsgen۰CompressedFileInfo:
		gr, err := gzip.NewReader(bytes.NewReader(f.compressedContent))
		if err != nil {
			// This should never happen because we generate the gzip bytes such that they are always valid.
			panic("unexpected error reading own gzip compressed bytes: " + err.Error())
		}
		return &vfsgen۰CompressedFile{
			vfsgen۰CompressedFileInfo: f,
			gr:                        gr,
		}, nil
	case *vfsgen۰FileInfo:
		return &vfsgen۰File{
			vfsgen۰FileInfo: f,
			Reader:          bytes.NewReader(f.content),
		}, nil
	case *vfsgen۰DirInfo:
		return &vfsgen۰Dir{
			vfsgen۰DirInfo: f,
		}, nil
	default:
		// This should never happen because we generate only the above types.
		panic(fmt.Sprintf("unexpected type %T", f))
	}
}

// vfsgen۰CompressedFileInfo is a static definition of a gzip compressed file.
type vfsgen۰CompressedFileInfo struct {
	name              string
	modTime           time.Time
	compressedContent []byte
	uncompressedSize  int64
}

func (f *vfsgen۰CompressedFileInfo) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}
func (f *vfsgen۰CompressedFileInfo) Stat() (os.FileInfo, error) { return f, nil }

func (f *vfsgen۰CompressedFileInfo) GzipBytes() []byte {
	return f.compressedContent
}

func (f *vfsgen۰CompressedFileInfo) Name() string       { return f.name }
func (f *vfsgen۰CompressedFileInfo) Size() int64        { return f.uncompressedSize }
func (f *vfsgen۰CompressedFileInfo) Mode() os.FileMode  { return 0444 }
func (f *vfsgen۰CompressedFileInfo) ModTime() time.Time { return f.modTime }
func (f *vfsgen۰CompressedFileInfo) IsDir() bool        { return false }
func (f *vfsgen۰CompressedFileInfo) Sys() interface{}   { return nil }

// vfsgen۰CompressedFile is an opened compressedFile instance.
type vfsgen۰CompressedFile struct {
	*vfsgen۰CompressedFileInfo
	gr      *gzip.Reader
	grPos   int64 // Actual gr uncompressed position.
	seekPos int64 // Seek uncompressed position.
}

func (f *vfsgen۰CompressedFile) Read(p []byte) (n int, err error) {
	if f.grPos > f.seekPos {
		// Rewind to beginning.
		err = f.gr.Reset(bytes.NewReader(f.compressedContent))
		if err != nil {
			return 0, err
		}
		f.grPos = 0
	}
	if f.grPos < f.seekPos {
		// Fast-forward.
		_, err = io.CopyN(ioutil.Discard, f.gr, f.seekPos-f.grPos)
		if err != nil {
			return 0, err
		}
		f.grPos = f.seekPos
	}
	n, err = f.gr.Read(p)
	f.grPos += int64(n)
	f.seekPos = f.grPos
	return n, err
}
func (f *vfsgen۰CompressedFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.seekPos = 0 + offset
	case io.SeekCurrent:
		f.seekPos += offset
	case io.SeekEnd:
		f.seekPos = f.uncompressedSize + offset
	default:
		panic(fmt.Errorf("invalid whence value: %v", whence))
	}
	return f.seekPos, nil
}
func (f *vfsgen۰CompressedFile) Close() error {
	return f.gr.Close()
}

// vfsgen۰FileInfo is a static definition of an uncompressed file (because it's not worth gzip compressing).
type vfsgen۰FileInfo struct {
	name    string
	modTime time.Time
	content []byte
}

func (f *vfsgen۰FileInfo) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}
func (f *vfsgen۰FileInfo) Stat() (os.FileInfo, error) { return f, nil }

func (f *vfsgen۰FileInfo) NotWorthGzipCompressing() {}

func (f *vfsgen۰FileInfo) Name() string       { return f.name }
func (f *vfsgen۰FileInfo) Size() int64        { return int64(len(f.content)) }
func (f *vfsgen۰FileInfo) Mode() os.FileMode  { return 0444 }
func (f *vfsgen۰FileInfo) ModTime() time.Time { return f.modTime }
func (f *vfsgen۰FileInfo) IsDir() bool        { return false }
func (f *vfsgen۰FileInfo) Sys() interface{}   { return nil }

// vfsgen۰File is an opened file instance.
type vfsgen۰File struct {
	*vfsgen۰FileInfo
	*bytes.Reader
}

func (f *vfsgen۰File) Close() error {
	return nil
}

// vfsgen۰DirInfo is a static definition of a directory.
type vfsgen۰DirInfo struct {
	name    string
	modTime time.Time
	entries []os.FileInfo
}

func (d *vfsgen۰DirInfo) Read([]byte) (int, error) {
	return 0, fmt.Errorf("cannot Read from directory %s", d.name)
}
func (d *vfsgen۰DirInfo) Close() error               { return nil }
func (d *vfsgen۰DirInfo) Stat() (os.FileInfo, error) { return d, nil }

func (d *vfsgen۰DirInfo) Name() string       { return d.name }
func (d *vfsgen۰DirInfo) Size() int64        { return 0 }
func (d *vfsgen۰DirInfo) Mode() os.FileMode  { return 0755 | os.ModeDir }
func (d *vfsgen۰DirInfo) ModTime() time.Time { return d.modTime }
func (d *vfsgen۰DirInfo) IsDir() bool        { return true }
func (d *vfsgen۰DirInfo) Sys() interface{}   { return nil }

// vfsgen۰Dir is an opened dir instance.
type vfsgen۰Dir struct {
	*vfsgen۰DirInfo
	pos int // Position within entries for Seek and Readdir.
}

func (d *vfsgen۰Dir) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == io.SeekStart {
		d.pos = 0
		return 0, nil
	}
	return 0, fmt.Errorf("unsupported Seek in directory %s", d.name)
}

func (d *vfsgen۰Dir) Readdir(count int) ([]os.FileInfo, error) {
	if d.pos >= len(d.entries) && count > 0 {
		return nil, io.EOF
	}
	if count <= 0 || count > len(d.entries)-d.pos {
		count = len(d.entries) - d.pos
	}
	e := d.entries[d.pos : d.pos+count]
	d.pos += count
	return e, nil
}
