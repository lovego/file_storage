package filestorage

import (
	"fmt"
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
		b.cleanDB(tx, cleanAfter)

		return nil
	})
}

func (b *Bucket) cleanDB(tx DB, cleanAfter time.Duration) ([]string, error) {
	rows, err := tx.Query(fmt.Sprintf(`
	DELETE FROM %s
	WHERE NOT EXISTS (
	  SELECT 1 FROM %s WHERE file = hash
	) AND created_at < %s
	RETURNING hash
	`, b.FilesTable, b.LinksTable, fmtTime(time.Now().Add(-cleanAfter)),
	))
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
