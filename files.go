package file_storage

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/lovego/slice"
)

func (s *Storage) createFilesTable(db DB) error {
	if s.FilesTable == "" {
		s.FilesTable = "files"
	}
	_, err := db.Exec(fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		hash           text        NOT NULL UNIQUE,
		type           text        NOT NULL,
		size           int8        NOT NULL,
		tranformations jsonb       NOT NULL DEFAULT '{}',
		created_at     timestamptz NOT NULL
	)`, s.FilesTable,
	))
	return err
}

type fileRecord struct {
	Hash string
	Type string
	Size int64
	File multipart.File
	New  bool
}

func (s *Storage) createFileRecords(
	db DB, fileHeaders []*multipart.FileHeader,
	contentTypeCheck func(*multipart.FileHeader, string) error,
) ([]fileRecord, error) {
	records := make([]fileRecord, 0, len(fileHeaders))
	for _, header := range fileHeaders {
		file, err := header.Open()
		if err != nil {
			return records, err
		}
		contentType, err := getContentType(file)
		if err != nil {
			return records, err
		}
		if contentTypeCheck != nil {
			if err := contentTypeCheck(header, contentType); err != nil {
				return records, err
			}
		}
		hash, err := getContentHash(file)
		if err != nil {
			return records, err
		}
		records = append(records, fileRecord{
			Hash: hash, Type: contentType, Size: header.Size, File: file,
		})
	}
	if err := s.insertFileRecords(db, records); err != nil {
		return records, err
	}
	return records, nil
}

func (s *Storage) insertFileRecords(db DB, records []fileRecord) error {
	var values []string
	now := nowTime()
	for _, record := range records {
		values = append(values, fmt.Sprintf("(%s, %s, %d, %s)",
			quote(record.Hash), quote(record.Type), record.Size, now,
		))
	}

	rows, err := db.Query(fmt.Sprintf(`
	INSERT INTO %s
		(hash, type, size, created_at)
	VALUES
		%s
	ON CONFLICT (hash) DO NOTHING
	RETURNING hash
	`, s.FilesTable, strings.Join(values, ",\n\t\t"),
	))
	if err != nil {
		return err
	}
	defer rows.Close()
	var inserted []string
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return err
		}
		inserted = append(inserted, hash)
	}
	for i := range records {
		records[i].New = slice.ContainsString(inserted, records[i].Hash)
	}

	return err
}

func getContentType(file multipart.File) (string, error) {
	var array [512]byte
	n, err := file.Read(array[:])
	if err != nil {
		return "", err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return http.DetectContentType(array[:n]), nil
}

func getContentHash(file multipart.File) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil)), nil
}

func nowTime() string {
	return time.Now().Format("'2006-01-02T15:04:05.999999Z07:00'")
}

func quote(s string) string {
	s = strings.Replace(s, "'", "''", -1)
	s = strings.Replace(s, "\000", "", -1)
	return "'" + s + "'"
}
