package filestorage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var errEmptyObject = errors.New("object is empty")
var errNotLinked = errors.New("the file is not linked to the object")

// IsNotLinked check if an error is the not linked Error.
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

// Link files to object.
func (s *Storage) Link(db DB, object string, files ...string) error {
	if object == "" {
		return errEmptyObject
	}
	if len(files) == 0 {
		return nil
	}

	var values []string
	now := nowTime()
	for _, file := range files {
		values = append(values, fmt.Sprintf("(%s, %s, %s)", quote(file), quote(object), now))
	}
	_, err := db.Exec(fmt.Sprintf(`
	INSERT INTO %s (file, object, created_at)
	VALUES %s
	ON CONFLICT (file, object) DO NOTHING
	`, s.LinksTable, strings.Join(values, ", "),
	))
	return err
}

// LinkOnly make sure these files and only these files are linked to object.
func (s *Storage) LinkOnly(db DB, object string, files ...string) error {
	if err := s.Link(db, object, files...); err != nil {
		return err
	}
	if len(files) == 0 {
		return s.unlink(db, object, "")
	}
	return s.unlink(db, object, filesCond(files, "NOT"))
}

// UnlinkAllOf unlink all linked files from an object.
func (s *Storage) UnlinkAllOf(db DB, object string) error {
	if object == "" {
		return errEmptyObject
	}
	return s.unlink(db, object, "")
}

// Unlink files from object.
func (s *Storage) Unlink(db DB, object string, files ...string) error {
	if object == "" {
		return errEmptyObject
	}
	if len(files) == 0 {
		return nil
	}
	return s.unlink(db, object, filesCond(files, ""))
}

func (s *Storage) unlink(db DB, object string, conds string) error {
	_, err := db.Exec(
		fmt.Sprintf(`DELETE FROM %s WHERE object = %s `, s.LinksTable, quote(object)) + conds,
	)
	return err
}

// EnsureLinked ensure file is linked to object.
func (s *Storage) EnsureLinked(db DB, object, file string) error {
	if ok, err := s.Linked(db, object, file); err != nil {
		return err
	} else if !ok {
		return errNotLinked
	}
	return nil
}

// Linked check if file is linked to object.
func (s *Storage) Linked(db DB, object, file string) (bool, error) {
	row := db.QueryRow(fmt.Sprintf(`
	SELECT true FROM %s WHERE object = %s AND file = %s
	`, s.LinksTable, quote(object), quote(file),
	))
	var linked bool
	if err := row.Scan(&linked); err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return linked, nil
}

// FilesOf get all files linked to an object.
func (s *Storage) FilesOf(db DB, object string) ([]string, error) {
	rows, err := db.Query(fmt.Sprintf(`
	SELECT file FROM %s WHERE object = %s ORDER BY created_at
	`, s.LinksTable, quote(object),
	))
	if err != nil {
		return nil, err
	}

	var files []string
	for rows.Next() {
		var file string
		if err := rows.Scan(&file); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func filesCond(files []string, not string) string {
	var quoted = make([]string, len(files))
	for i := range files {
		quoted[i] = quote(files[i])
	}
	return fmt.Sprintf(" AND file %s IN (%s)", not, strings.Join(quoted, ", "))
}

var objectRegexp = regexp.MustCompile(`^(\w+)\.(\d+)(\.(\w+))?$`)
var errInvalidObject = errors.New("invalid LinkObject")

// LinkObject is a structured reference implementaion for link object string.
type LinkObject struct {
	Table string
	ID    int64
	Field string
}

func (o LinkObject) String() string {
	s := o.Table + "." + strconv.FormatInt(o.ID, 10)
	if o.Field != "" {
		s += "." + o.Field
	}
	return s
}

// MarshalJSON implements json.Marshaler
func (o LinkObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (o *LinkObject) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	m := objectRegexp.FindStringSubmatch(s)
	if len(m) == 0 {
		return errInvalidObject
	}
	o.Table = m[1]
	if id, err := strconv.ParseInt(m[2], 10, 64); err != nil {
		return errInvalidObject
	} else {
		o.ID = id
	}
	o.Field = m[4]
	return nil
}
