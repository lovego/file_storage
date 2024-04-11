package filestorage

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func Open(req *http.Request) ([]byte, error) {
	q := req.URL.Query()
	bucket, err := GetBucket(q.Get("b"))
	if err != nil {
		return nil, err
	}
	return bucket.ReadFile(nil, q.Get("f"), q.Get("o"))
}
func GetFile(req *http.Request) (*os.File, error) {
	q := req.URL.Query()
	bucket, err := GetBucket(q.Get("b"))
	if err != nil {
		return nil, err
	}
	return bucket.GetFile(nil, q.Get("f"), q.Get("o"))
}
func (b *Bucket) GetFile(db DB, file string, object string) (*os.File, error) {
	if err := CheckHash(file); err != nil {
		return nil, err
	}
	if object != "" {
		if err := b.EnsureLinked(db, object, file); err != nil {
			return nil, err
		}
	}

	f, err := os.Open(filepath.Join(b.Dir, b.FilePath(file)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("file not exist")
		}
		return nil, err
	}
	return f, nil
}
func (b *Bucket) ReadFile(db DB, file string, object string) ([]byte, error) {
	if err := CheckHash(file); err != nil {
		return nil, err
	}
	if object != "" {
		if err := b.EnsureLinked(db, object, file); err != nil {
			return nil, err
		}
	}

	f, err := os.Open(filepath.Join(b.Dir, b.FilePath(file)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("file not exist")
		}
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}
