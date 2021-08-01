package filestorage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lovego/errs"
)

var errEmptyObject = errs.New("args-err", "object is empty")
var errNotLinked = errs.New("args-err", "the file is not linked to the object")

// IsNotLinked check if an error is the not linked Error.
func IsNotLinked(err error) bool {
	return err == errNotLinked
}

func (b *Bucket) createLinksTable(db DB) error {
	if b.LinksTable == "" {
		b.LinksTable = "file_links"
	}

	_, err := b.getDB(db).Exec(fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		file       text NOT NULL,
		object     text NOT NULL,
		created_at timestamptz NOT NULL,
		unique(file, object)
	);
	CREATE INDEX IF NOT EXISTS %s_object_index ON %s(object);
	`, b.LinksTable, b.LinksTable, b.LinksTable,
	))
	return err
}

// Link files to object.
func (b *Bucket) Link(db DB, object string, files ...string) error {
	if object == "" {
		return errEmptyObject
	}
	if len(files) == 0 {
		return nil
	}
	if err := CheckHash(files...); err != nil {
		return err
	}

	var values []string
	now := nowTime()
	for _, file := range files {
		values = append(values, fmt.Sprintf("(%s, %s, %s)", quote(file), quote(object), now))
	}
	_, err := b.getDB(db).Exec(fmt.Sprintf(`
	INSERT INTO %s (file, object, created_at)
	VALUES %s
	ON CONFLICT (file, object) DO NOTHING
	`, b.LinksTable, strings.Join(values, ", "),
	))
	return err
}

// LinkOnly make sure these files and only these files are linked to object.
func (b *Bucket) LinkOnly(db DB, object string, files ...string) error {
	if object == "" {
		return errEmptyObject
	}
	if err := b.Link(db, object, files...); err != nil {
		return err
	}
	if len(files) == 0 {
		return b.unlink(db, object, "")
	}
	return runInTx(db, func(tx DB) error {
		return b.linkOnly(db, object, files)
	})
}

func (b *Bucket) linkOnly(db DB, object string, files []string) error {
	if err := CheckHash(files...); err != nil {
		return err
	}
	return b.unlink(db, object, filesCond(files, "NOT"))
}

// UnlinkAllOf unlink all linked files from an object.
func (b *Bucket) UnlinkAllOf(db DB, object string) error {
	if object == "" {
		return errEmptyObject
	}
	return b.unlink(db, object, "")
}

// Unlink files from object.
func (b *Bucket) Unlink(db DB, object string, files ...string) error {
	if object == "" {
		return errEmptyObject
	}
	if len(files) == 0 {
		return nil
	}
	if err := CheckHash(files...); err != nil {
		return err
	}
	return b.unlink(db, object, filesCond(files, ""))
}

func (b *Bucket) unlink(db DB, object string, conds string) error {
	_, err := b.getDB(db).Exec(
		fmt.Sprintf(`DELETE FROM %s WHERE object = %s `, b.LinksTable, quote(object)) + conds,
	)
	return err
}

// EnsureLinked ensure file is linked to object.
func (b *Bucket) EnsureLinked(db DB, object, file string) error {
	if ok, err := b.Linked(db, object, file); err != nil {
		return err
	} else if !ok {
		return errNotLinked
	}
	return nil
}

// Linked check if file is linked to object.
func (b *Bucket) Linked(db DB, object, file string) (bool, error) {
	if err := CheckHash(file); err != nil {
		return false, err
	}
	row := b.getDB(db).QueryRow(fmt.Sprintf(`
	SELECT true FROM %s WHERE object = %s AND file = %s
	`, b.LinksTable, quote(object), quote(file),
	))
	var linked bool
	if err := row.Scan(&linked); err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return linked, nil
}

// FilesOf get all files linked to an object.
func (b *Bucket) FilesOf(db DB, object string) ([]string, error) {
	rows, err := b.getDB(db).Query(fmt.Sprintf(`
	SELECT file FROM %s WHERE object = %s ORDER BY created_at
	`, b.LinksTable, quote(object),
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
