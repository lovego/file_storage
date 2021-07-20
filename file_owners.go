package file_storage

import (
	"fmt"

	"github.com/lovego/bsql"
)

// transfer files from oldOwner to newOwner.
func (s *Storage) Transfer(db bsql.DbOrTx, oldOwner, newOwner string, files ...string) error {
	return nil
}

// remove files of owner.
func (s *Storage) Remove(db bsql.DbOrTx, owner string, files ...string) error {
	return nil
}

// remove all files of owner.
func (s *Storage) RemoveAllOf(db bsql.DbOrTx, owner string) error {
	return nil
}

// add new owner for files
func (s *Storage) AddOwner(db bsql.DbOrTx, owner string, files ...string) error {
	return nil
}

// list files of owner.
func (s *Storage) FilesOf(db bsql.DbOrTx, owner string) ([]string, error) {
	return nil, nil
}

func (s *Storage) createFileOwnersTable(db bsql.DbOrTx) error {
	if s.FileOwnersTable == "" {
		s.FileOwnersTable = "file_owners"
	}

	return db.Exec(fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		file           text   NOT NULL,
		owner          text   NOT NULL,
		created_at     timestamptz NOT NULL,
		updated_at     timestamptz NOT NULL,
		unique(file, owner)
	)`, s.FileOwnersTable,
	))
}
