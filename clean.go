package filestorage

import (
	"fmt"
	"os"
	"os/exec"
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
			if err := b.deleteFile(file); err != nil {
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

func (b *Bucket) deleteFile(hash string) error {
	var dir = b.FileDir(hash)
	var path = filepath.Join(dir, hash)
	var script = fmt.Sprintf(
		`cd %s; test -f %s && rm -f %s && rmdir -p --ignore-fail-on-non-empty %s || true`,
		b.Dir, path, path, dir,
	)

	if b.localMachine {
		var cmd = exec.Command("bash", "-c", script)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	for _, addr := range b.otherMachines {
		var cmd = exec.Command("ssh", addr, script)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
