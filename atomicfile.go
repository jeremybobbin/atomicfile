package atomicfile

import (
	"context"
	"os"
	"path/filepath"
)

type File struct {
	file *os.File
	err  chan chan error
}

func New(ctx context.Context, path string) (*File, error) {
	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return nil, err
	}

	outer := make(chan chan error)

	go func() {
		var err error
		inner := make(chan error)
		defer func() {
			inner <- err
			close(inner)
		}()
		select {
		case outer <- inner:
			if err = file.Close(); err != nil {
				// no-op
			} else if err = os.Rename(file.Name(), path); err != nil {
				// no-op
			} else {
				return
			}
		case <-ctx.Done():
		}
		os.Remove(file.Name())
	}()

	return &File{
		file: file,
		err:  outer,
	}, nil
}

func (f *File) Write(p []byte) (int, error) {
	return f.file.Write(p)
}

func (f *File) Close() (err error) {
	return <-<-f.err
}

func WithParents(ctx context.Context, path string) (*File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return New(ctx, path)
}
