package fs

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"time"

	"github.com/xbugio/aliyundrive-go-sdk"
)

type Fs struct {
	c    *aliyundrive.Drive
	root string
}

func New(c *aliyundrive.Drive, root string) fs.FS {
	return &Fs{c: c, root: root}
}

func (f *Fs) Open(name string) (fs.File, error) {
	item, err := f.pathToItem(name)
	if err != nil {
		return nil, err
	}
	return &File{fs: f, item: item}, nil
}

func (f *Fs) pathToItem(p string) (item *aliyundrive.Item, err error) {
	p = path.Join(f.root, p)
	paths := splitPath(p)

	ctx := context.Background()
	parentFileId := "root"
	exists := false
	for _, name := range paths {
		if name == "/" {
			parentFileId = "root"
			continue
		}
		nextmarker := ""
		for {
			resp, err := f.c.DoListRequest(ctx, aliyundrive.ListRequest{
				ParentFileId: parentFileId,
				Limit:        200,
				NextMarker:   nextmarker,
			})
			if err != nil {
				return nil, err
			}
			item, exists = aliyundrive.ItemQuery(resp.Items).ByName(name)
			if exists || resp.NextMarker == "" {
				break
			}
			nextmarker = resp.NextMarker
		}
		if !exists {
			return nil, fs.ErrNotExist
		}
		parentFileId = item.FileId
	}
	return
}

type File struct {
	fs   *Fs
	item *aliyundrive.Item

	cancel context.CancelFunc
	body   io.ReadCloser
	offset int64
}

func (f *File) Name() string {
	return f.item.Name
}
func (f *File) Size() int64 {
	return int64(f.item.Size)
}
func (f *File) Mode() fs.FileMode {
	if f.item.Type == "folder" {
		return fs.ModeDir
	}
	return 0
}

func (f *File) ModTime() time.Time {
	return f.item.UpdatedAt
}

func (f *File) IsDir() bool {
	return f.item.Type == "folder"
}

func (f *File) Sys() any {
	return nil
}

func (f *File) Stat() (fs.FileInfo, error) {
	return f, nil
}
func (f *File) Read(p []byte) (int, error) {
	// 第一次读，初始化
	if f.body == nil {
		if err := f.prepareReader(0); err != nil {
			return 0, err
		}
	}
	n, err := f.body.Read(p)
	f.offset += int64(n)
	return n, err
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if whence != io.SeekStart {
		return 0, fs.ErrInvalid
	}

	if f.offset == offset {
		return offset, nil
	}

	f.close()
	if err := f.prepareReader(offset); err != nil {
		return 0, err
	}
	return offset, nil
}

func (f *File) Close() error {
	return f.close()
}

func (f *File) prepareReader(offset int64) error {
	ctx := context.Background()

	// 获取下载链接
	getDownloadUrlResp, err := f.fs.c.DoGetDownloadUrlRequest(ctx, aliyundrive.GetDownloadUrlRequest{
		FileId: f.item.FileId,
	})
	if err != nil {
		return err
	}

	// 获取数据流
	header := make(http.Header)
	if offset > 0 {
		header.Set("Range", fmt.Sprintf("bytes=%v-", offset))
	}
	ctx, cancel := context.WithCancel(context.Background())
	downloadResp, err := f.fs.c.DoDownloadFileRequest(ctx, aliyundrive.DownloadFileRequest{
		Url:    getDownloadUrlResp.Url,
		Header: header,
	})
	if err != nil {
		cancel()
		return err
	}
	f.body = downloadResp.Reader
	f.cancel = cancel
	f.offset = offset
	return nil
}

func (f *File) close() error {
	if f.body == nil {
		return nil
	}
	f.cancel()
	io.Copy(io.Discard, f.body)
	err := f.body.Close()
	f.cancel = nil
	f.body = nil
	f.offset = 0
	return err
}

func splitPath(p string) []string {
	p = path.Clean(p)
	if p == "." || p == "/" {
		return []string{"/"}
	}

	dir, file := path.Split(p)
	parents := splitPath(dir)
	return append(parents, file)
}
