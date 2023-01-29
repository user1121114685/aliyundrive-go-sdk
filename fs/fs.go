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
	return f.open(context.Background(), name)
}

func (f *Fs) ReadDir(name string) ([]fs.DirEntry, error) {
	file, err := f.open(context.Background(), name)
	if err != nil {
		return nil, err
	}
	defer file.close()
	return file.ReadDir(-1)
}

func (f *Fs) Stat(name string) (fs.FileInfo, error) {
	file, err := f.open(context.Background(), name)
	if err != nil {
		return nil, err
	}
	defer file.close()
	return file, nil
}

func (f *Fs) Sub(dir string) (fs.FS, error) {
	root := path.Join(f.root, dir)
	return New(f.c, root), nil
}

func (f *Fs) open(ctx context.Context, p string) (*File, error) {
	root, err := f.c.DoGetRequest(ctx,
		aliyundrive.GetRequest{FileId: aliyundrive.RootFileId})
	if err != nil {
		return nil, err
	}

	p = path.Join(f.root, p)
	paths := splitPath(p)
	file := &File{fs: f, item: &root.Item}

	for _, name := range paths {
		if name == "/" {
			continue
		}

		next := ""
		var targetItem *aliyundrive.Item
		for {
			var items []*aliyundrive.Item
			items, next, err = file.list(ctx, aliyundrive.LimitMax, next)
			if err != nil {
				return nil, err
			}
			for _, item := range items {
				if item.Name == name {
					targetItem = item
					break
				}
			}
			if targetItem != nil || next == "" {
				break
			}
		}

		if targetItem == nil {
			return nil, fs.ErrNotExist
		}
		file = &File{fs: f, item: targetItem}
	}

	return file, nil
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

func (f *File) Type() fs.FileMode {
	return f.Mode()
}

func (f *File) ModTime() time.Time {
	return f.item.UpdatedAt
}

func (f *File) IsDir() bool {
	return f.item.Type == "folder"
}

func (f *File) Info() (fs.FileInfo, error) {
	return f, nil
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
		err := f.prepareReader(context.Background(), 0)
		if err != nil {
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
	err := f.prepareReader(context.Background(), offset)
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func (f *File) ReadDir(n int) (entries []fs.DirEntry, err error) {
	if n > 1 {
		return nil, fs.ErrInvalid
	}
	ctx := context.Background()
	next := ""
	var items []*aliyundrive.Item
	for {
		items, next, err = f.list(ctx, aliyundrive.LimitMax, next)
		if err != nil {
			return
		}
		for _, item := range items {
			entries = append(entries, &File{fs: f.fs, item: item})
		}
		if next == "" {
			break
		}
	}
	return
}

func (f *File) Close() error {
	return f.close()
}

func (f *File) prepareReader(ctx context.Context, offset int64) error {
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

func (f *File) list(ctx context.Context, limit int, next string) (items []*aliyundrive.Item, nextMarker string, err error) {
	if !f.IsDir() {
		err = fs.ErrInvalid
		return
	}

	resp, err := f.fs.c.DoListRequest(ctx, aliyundrive.ListRequest{
		ParentFileId:   f.item.FileId,
		OrderBy:        aliyundrive.OrderByName,
		OrderDirection: aliyundrive.OrderDirectionAsc,
		Limit:          limit,
		NextMarker:     next,
	})
	if err != nil {
		return
	}
	nextMarker = resp.NextMarker
	items = resp.Items
	return
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
