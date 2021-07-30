package filestorage

import (
	"context"
	"database/sql"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lovego/addrs"
)

// Upload files, if object is not empty, the files are linked to it.
func (s *Storage) Upload(
	db DB, contentTypeCheck func(string) error, object string, fileHeaders ...*multipart.FileHeader,
) ([]string, error) {
	var files = make([]File, len(fileHeaders))
	for i := range fileHeaders {
		f, err := fileHeaders[i].Open()
		if err != nil {
			return nil, err
		}
		defer f.Close()
		files[i].IO = f
		files[i].Size = fileHeaders[i].Size
	}
	return s.Save(db, contentTypeCheck, object, files...)
}

// File reprents the file to store.
type File struct {
	IO   io.ReadSeeker
	Size int64
}

// Save file into storage.
func (s *Storage) Save(
	db DB, contentTypeCheck func(string) error, object string, files ...File,
) (fileHashes []string, err error) {
	if len(files) == 0 {
		return nil, nil
	}
	err = runInTx(db, func(tx DB) error {
		hashes, err := s.save(tx, contentTypeCheck, object, files)
		if err != nil {
			return err
		}
		fileHashes = hashes
		return nil
	})
	return
}

func (s *Storage) save(
	db DB, contentTypeCheck func(string) error, object string, files []File,
) ([]string, error) {
	records, err := s.createFileRecords(db, files, contentTypeCheck)
	if err != nil {
		return nil, err
	}
	var hashes []string
	for i := range records {
		hashes = append(hashes, records[i].Hash)
	}
	if object != "" {
		if err := s.Link(db, object, hashes...); err != nil {
			return nil, err
		}
	}
	for i := range records {
		if records[i].New {
			if err := s.saveFile(records[i].File, records[i].Hash); err != nil {
				return nil, err
			}
		}
	}
	return hashes, nil
}

func (s *Storage) saveFile(file io.Reader, hash string) error {
	var srcPath string
	var destPath = filepath.Join(s.ScpPath, s.FilePath(hash))
	if s.localMachine {
		if err := s.writeFile(file, destPath); err != nil {
			return err
		}
		srcPath = destPath
	} else {
		tempFile, err := s.writeTempFile(file)
		if err != nil {
			return err
		}
		srcPath = tempFile
	}
	for _, addr := range s.otherMachines {
		if err := exec.Command("scp", srcPath, addr+":"+destPath).Run(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) writeFile(file io.Reader, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	defer destFile.Close()
	_, err = io.Copy(destFile, file)
	return err
}

func (s *Storage) writeTempFile(file io.Reader) (string, error) {
	temp, err := ioutil.TempFile("", "fs_")
	if err != nil {
		return "", err
	}
	defer temp.Close()
	if _, err := io.Copy(temp, file); err != nil {
		return "", err
	}
	return temp.Name(), nil
}

func (s *Storage) parseMachines() error {
	var user string
	if s.ScpUser != "" {
		user = s.ScpUser + "@"
	}
	for _, addr := range s.ScpMachines {
		if ok, err := addrs.IsLocalhost(addr); err != nil {
			return err
		} else if ok {
			s.localMachine = true
		} else {
			s.otherMachines = append(s.otherMachines, user+addr)
		}
	}
	return nil
}

// FilePath returns the file path to store on disk.
func (s *Storage) FilePath(hash string) string {
	var path string
	var i uint8
	for ; i < s.DirDepth; i++ {
		path = filepath.Join(path, hash[i:i+1])
	}
	return filepath.Join(path, hash)
}

func runInTx(db DB, work func(DB) error) error {
	if sqldb, ok := db.(*sql.DB); ok {
		tx, err := sqldb.BeginTx(context.Background(), nil)
		if err != nil {
			return err
		}
		defer func() {
			if err := recover(); err != nil {
				_ = tx.Rollback()
				panic(err)
			}
		}()
		if err := work(tx); err != nil {
			_ = tx.Rollback()
			return err
		}
		return tx.Commit()
	}
	return work(db)
}
