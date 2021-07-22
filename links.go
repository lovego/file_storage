package file_storage

import (
	"errors"
	"fmt"
)

var errNotLinked = errors.New("the file is not linked to the object")

func IsNotLinked(err error) bool {
	return err == errNotLinked
}

func (s *Storage) createLinksTable(db DB) error {
	if s.LinksTable == "" {
		s.LinksTable = "file_links"
	}

	_, err := db.Exec(fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		file       text NOT NULL,
		object     text NOT NULL,
		created_at timestamptz NOT NULL,
		unique(file, object)
	);
	CREATE INDEX IF NOT EXISTS %s_object_index ON %s(object);
	`, s.LinksTable, s.LinksTable, s.LinksTable,
	))
	return err
}

// link files to object.
func (s *Storage) Link(db DB, object string, files ...string) error {
	return nil
}

// make sure files and only files are linked to object.
func (s *Storage) LinkOnly(db DB, object string, files ...string) error {
	return nil
}

// unlink files from object.
func (s *Storage) Unlink(db DB, object string, files ...string) error {
	return nil
}

// unlink all linked files from an object.
func (s *Storage) UnlinkAllOf(db DB, object string) error {
	return nil
}

// ensure file is linked to object.
func (s *Storage) EnsureLinked(db DB, object, file string) error {
	if ok, err := s.Linked(db, object, file); err != nil {
		return err
	} else if !ok {
		return errNotLinked
	}
	return nil
}

// check if file is linked to object.
func (s *Storage) Linked(db DB, object, file string) (bool, error) {
	return false, nil
}

// get all files linked to an object.
func (s *Storage) FilesOf(db DB, object string) ([]string, error) {
	return nil, nil
}
