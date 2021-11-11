// Package filestorage upload/store/download files, and clean files that are not linked to any objects.
package filestorage

import (
	"database/sql"
	"errors"
	"net/url"
	"path/filepath"

	"github.com/lovego/errs"
)

var buckets = make(map[string]*Bucket)

// Bucket store file on disk and infomation in database tables.
type Bucket struct {
	Name     string
	Machines []string
	Dir      string
	DirDepth uint8
	ScpUser  string

	DownloadURLPrefix string

	// Path prefix for "X-Accel-Redirect" response header when downloading.
	// If this path prefix is present, only file path is sent in the "X-Accel-Redirect" header,
	// and nginx is responsible for file downloading for a better performance.
	// Otherwise, file is sent directly in the response body.
	RedirectPathPrefix string

	DB         DB
	FilesTable string
	LinksTable string

	localMachine  bool
	otherMachines []string
}

// DB represents *sql.DB or *sql.Tx
type DB interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Init validate storage fields and create tables if not created.
func (b *Bucket) Init(db DB) error {
	if len(b.Machines) == 0 {
		return errors.New("Machines is empty")
	}
	if b.Dir == "" {
		return errors.New("Dir is empty")
	} else {
		b.Dir = filepath.Clean(b.Dir)
	}
	if !filepath.IsAbs(b.Dir) {
		return errors.New("Dir is not an absolute path")
	}
	if b.DirDepth == 0 {
		b.DirDepth = 3
	} else if b.DirDepth > 8 {
		return errors.New("DirDepth at most be 8")
	}
	if b.RedirectPathPrefix != "" && b.RedirectPathPrefix[0] != '/' {
		b.RedirectPathPrefix = "/" + b.RedirectPathPrefix
	}
	if err := b.createFilesTable(db); err != nil {
		return err
	}
	if err := b.createLinksTable(db); err != nil {
		return err
	}
	if err := b.parseMachines(); err != nil {
		return err
	}
	buckets[b.Name] = b
	return nil
}

func (b *Bucket) getDB(db DB) DB {
	if db != nil {
		return db
	}
	return b.DB
}

var errUnknownBucket = errs.New("args-err", "unknown bucket")

func GetBucket(bucketName string) (*Bucket, error) {
	if bucket := buckets[bucketName]; bucket != nil {
		return bucket, nil
	}
	return nil, errUnknownBucket
}

// FileHash returns file hash from a url or file hash
func FileHash(str string) (string, error) {
	if str == "" {
		return "", nil
	}
	if IsHash(str) {
		return str, nil
	}
	uri, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	hash := uri.Query().Get("f")
	if err := CheckHash(hash); err != nil {
		return "", err
	}
	return hash, nil
}

// FileHashes returns file hashes from urls or file hashes
func FileHashes(strs []string) ([]string, error) {
	hashes := make([]string, len(strs))
	for i, str := range strs {
		if hash, err := FileHash(str); err != nil {
			return nil, err
		} else {
			hashes[i] = hash
		}
	}
	return hashes, nil
}

// TryFileHash try to returns file hash from a url or file hash.
func TryFileHash(str string) (string, error) {
	if str == "" {
		return "", nil
	}
	if IsHash(str) {
		return str, nil
	}
	uri, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	hash := uri.Query().Get("f")
	if !IsHash(hash) {
		return "", nil
	}
	return hash, nil
}
