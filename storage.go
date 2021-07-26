// Package file_storage upload/store/download files, and clean unused files.
package file_storage

import (
	"database/sql"
	"errors"
	"mime/multipart"
	"net/http"
	"path"
)

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
type DB interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func (s *Storage) Init(db DB) error {
	if len(s.ScpMachines) == 0 {
		return errors.New("ScpMachines is empty.")
	}
	if s.ScpPath == "" {
		return errors.New("ScpPath is empty.")
	}
	if s.ScpPath[0] != '/' {
		return errors.New("ScpPath is not an absolute path.")
	}
	if s.DirDepth == 0 {
		s.DirDepth = 3
	} else if s.DirDepth > 8 {
		return errors.New("DirDepth at most be 8.")
	}
	if s.XAccelRedirectPrefix == "" {
		s.XAccelRedirectPrefix = "/fs"
	} else if s.XAccelRedirectPrefix[0] != '/' {
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

// upload files, if object is not empty, the files are linked to it.
func (s *Storage) Upload(
	db DB, contentTypeCheck func(*multipart.FileHeader, string) error, object string,
	files ...*multipart.FileHeader,
) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}
	records, err := s.createFileRecords(db, files, contentTypeCheck)
	defer func() {
		for i := range records {
			records[i].File.Close()
		}
	}()
	if err != nil {
		return nil, err
	}
	var hashes []string
	for i := range records {
		hashes = append(hashes, records[i].Hash)
	}
	if object != "" {
		if err := s.Link(db, object, hashes...); err != nil {
			return nil, err
		}
	}
	for i := range records {
		if records[i].New {
			if err := s.saveFile(records[i].File, records[i].Hash); err != nil {
				return nil, err
			}
		}
	}
	return hashes, nil
}

/* download file, if object is not empty, the file must be linked to it, otherwise an error is returned.
An location like following is required in nginx virtual server config.
	location /fs/ {
	  internal;
	  alias /data/file-storage;
	}
The location prefix should be XAccelRedirectPrefix + "/", the alias path should be ScpPath.
*/
func (s *Storage) Download(
	db DB, resp http.ResponseWriter, file string, object string,
) error {
	if object != "" {
		if err := s.EnsureLinked(db, object, file); err != nil {
			return err
		}
	}
	resp.Header().Set("X-Accel-Redirect", path.Join(s.XAccelRedirectPrefix, s.FilePath(file)))

	return nil
}
