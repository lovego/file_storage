// Package filestorage upload/store/download files, and clean unused files.
package filestorage

import (
	"database/sql"
	"errors"
)

// Storage do file storage on disk and infomation in database tables.
type Storage struct {
	ScpUser     string
	ScpMachines []string
	ScpPath     string
	DirDepth    uint8

	// path prefix for "X-Accel-Redirect" Response Header.
	XAccelRedirectPrefix string

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
	if s.XAccelRedirectPrefix != "" && s.XAccelRedirectPrefix[0] != '/' {
		s.XAccelRedirectPrefix = "/" + s.XAccelRedirectPrefix
	}

	if err := s.createFilesTable(db); err != nil {
		return err
	}
	if err := s.createLinksTable(db); err != nil {
		return err
	}

	return s.parseMachines()
}
