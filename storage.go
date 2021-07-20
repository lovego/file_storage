package file_storage

import (
	"net/http"

	"github.com/lovego/bsql"
)

type Storage struct {
	FilesTable, FileOwnersTable string

	ScpUser     string
	ScpMachines []string
	ScpPath     string

	ExternalURL string
}

func (s *Storage) Init(db bsql.DbOrTx) error {
	if err := s.createFilesTable(); err != nil {
		return err
	}
	if err := s.createFileOwnersTable(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) Upload(
	db bsql.DbOrTx, req *http.Request, owner, files ...string,
) (string, error) {
	return nil
}

func (s *Storage) Download(
	db bsql.DbOrTx, resp *http.ResponseWriter, owner string, file string,
) error {
	return nil
}
