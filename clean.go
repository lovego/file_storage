package filestorage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Logger interface {
	Error(args ...interface{})
}

func (b *Bucket) StartClean(cleanInterval, cleanAfter time.Duration, logger Logger) {
	if cleanInterval <= 0 || cleanAfter <= 0 {
		return
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(err)
			}
		}()
		for {
			if err := b.clean(cleanAfter); err != nil {
				logger.Error(err)
			}
			time.Sleep(cleanInterval)
		}
	}()
}

func (b *Bucket) clean(cleanAfter time.Duration) error {
	return runInTx(b.DB, func(tx DB) error {
		files, err := b.cleanDB(tx, cleanAfter)
		if err != nil {
			return err
		}
		for _, file := range files {
			if err := b.cleanFile(file); err != nil {
				return err
			}
		}

		return nil
	})
}

func (b *Bucket) cleanDB(tx DB, cleanAfter time.Duration) ([]string, error) {
	sql := fmt.Sprintf(`
	DELETE FROM %s
	WHERE NOT EXISTS (
	  SELECT 1 FROM %s WHERE file = hash
	) AND created_at < %s
	RETURNING hash
	`, b.FilesTable, b.LinksTable, fmtTime(time.Now().Add(-cleanAfter)),
	)
	rows, err := tx.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (b *Bucket) cleanFile(file string) error {
	var filePath = filepath.Join(b.Dir, b.FilePath(file))
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	var dir = filepath.Dir(filePath)
	for len(dir) > len(b.Dir) {
		if ok, err := emptyDir(dir); err != nil {
			return err
		} else if ok {
			if err := os.Remove(dir); err != nil && !os.IsNotExist(err) {
				return err
			}
		} else {
			break
		}
		dir = filepath.Dir(dir)
	}
	return nil
}

func emptyDir(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
