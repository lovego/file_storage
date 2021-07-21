package file_storage

import (
	"fmt"
)

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
	`, s.LinksTable,
	))
	return err
}

// link some files to an object.
func (s *Storage) Link(db DB, object string, files ...string) error {
	return nil
}

// unlink some files from an object.
func (s *Storage) Unlink(db DB, object string, files ...string) error {
	return nil
}

// unlink all linked files from an object.
func (s *Storage) UnlinkAllOf(db DB, object string) error {
	return nil
}

// get all files linked to an object.
func (s *Storage) AllOf(db DB, owner string) ([]string, error) {
	return nil, nil
}
