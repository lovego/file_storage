package file_storage

import (
	"fmt"

	"github.com/lovego/bsql"
)

/*
1. upload, download
2. owner check
2. clean unused files
*/

func (s *Storage) createFilesTable(db bsql.DbOrTx) error {
	if s.FilesTable == "" {
		s.FilesTable = "files"
	}
	return db.Exec(fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		file           text NOT NULL UNIQUE,
		type           text NOT NULL,
		created_at     timestamptz NOT NULL
	)`, s.FilesTable,
	))
}

func (s *Storage) createFiles(db bsql.DbOrTx, owner, files ...string) error {
}
