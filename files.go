package file_storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
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
	now := time.Now().Format("2006-01-02T15:04:05.999999Z07:00")
	for _, record := range records {
		values = append(values, fmt.Sprintf("('%s', '%s', %d, '%s')",
			record.Hash, record.Type, record.Size, now,
		))
	}

	_, err := db.Exec(fmt.Sprintf(`
	INSERT INTO %s (hash, type, size, created_at)
	VALUES %s
	`, s.FilesTable))
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
	return hex.EncodeToString(h.Sum(nil)), nil
}
