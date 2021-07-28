// Package filestorage upload/store/download files, and clean files that are not linked to any objects.
package filestorage

import (
	"database/sql"
	"errors"
	"net/url"
	"path"
)

// Storage do file storage on disk and infomation in database tables.
type Storage struct {
	ScpUser     string
	ScpMachines []string
	ScpPath     string
	DirDepth    uint8

	DownloadURLPrefix string

	// Path prefix for "X-Accel-Redirect" response header when downloading.
	// If this path prefix is present, only file path is sent in the "X-Accel-Redirect" header,
	// and nginx is responsible for file downloading for a better performance.
	// Otherwise, file is sent directly in the response body.
	RedirectPathPrefix string

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
func (s *Storage) Init(db DB) error {
	if len(s.ScpMachines) == 0 {
		return errors.New("ScpMachines is empty")
	}
	if s.ScpPath == "" {
		return errors.New("ScpPath is empty")
	}
	if s.ScpPath[0] != '/' {
		return errors.New("ScpPath is not an absolute path")
	}
	if s.DirDepth == 0 {
		s.DirDepth = 3
	} else if s.DirDepth > 8 {
		return errors.New("DirDepth at most be 8")
	}
	if s.RedirectPathPrefix != "" && s.RedirectPathPrefix[0] != '/' {
		s.RedirectPathPrefix = "/" + s.RedirectPathPrefix
	}

	if err := s.createFilesTable(db); err != nil {
		return err
	}
	if err := s.createLinksTable(db); err != nil {
		return err
	}

	return s.parseMachines()
}

// FileHash returns file hash from a url or file hash
func FileHash(str string) (string, error) {
	if IsHash(str) {
		return str, nil
	}
	uri, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	hash := path.Base(uri.Path)
	if err := CheckHash(hash); err != nil {
		return "", err
	}
	return hash, nil
}

// FileHash returns file hashes from urls or file hashes
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
