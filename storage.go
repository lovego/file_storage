// Package file_storage upload/store/download files, and clean unused files.
package file_storage

import (
	"database/sql"
	"mime/multipart"
	"net/http"

	"github.com/lovego/addrs"
)

type Storage struct {
	ScpUser     string
	ScpMachines []string
	ScpPath     string

	ExternalURL string

	FilesTable string
	LinksTable string

	localMachine  bool
	otherMachines []string
}

type DB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func (s *Storage) Init(db DB) error {
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
	for i := range records {
		if err := s.saveFile(records[i].File, records[i].Hash); err != nil {
			return nil, err
		}
	}
	return hashes, nil
}

// download file, if object is not empty, the file must be linked to it, otherwise an error is returned.
func (s *Storage) Download(
	db DB, resp *http.ResponseWriter, file string, object string,
) error {
	return nil
}

func (s *Storage) saveFile(file multipart.File, hash string) error {
	if s.localMachine {
	}
	for _, addr := range s.otherMachines {
		if err := s.scpFile(file, addr, path); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) scpFile(file multipart.File, addr, path string) error {
	return nil
}

func (s *Storage) parseMachines() error {
	for _, addr := range s.ScpMachines {
		if ok, err := addrs.IsLocalhost(addr); err != nil {
			return err
		} else if ok {
			s.localMachine = true
		} else {
			s.otherMachines = append(s.otherMachines, addr)
		}
	}
	return nil
}
