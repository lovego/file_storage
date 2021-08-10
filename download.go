package filestorage

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/lovego/errs"
)

// DownloadURL make the url for file download
func (b *Bucket) DownloadURL(linkObject interface{}, fileHash string) string {
	if fileHash == "" || !IsHash(fileHash) {
		return fileHash
	}
	q := url.Values{}
	q.Set("b", b.Name)   // bucket
	q.Set("f", fileHash) // file
	if linkObject != nil {
		q.Set("o", fmt.Sprint(linkObject)) // link object
	}
	return b.DownloadURLPrefix + "?" + q.Encode()
}

// DownloadURL make the urls for files download
func (b *Bucket) DownloadURLs(linkObject interface{}, fileHashes []string) []string {
	urls := make([]string, len(fileHashes))
	for i, hash := range fileHashes {
		urls[i] = b.DownloadURL(linkObject, hash)
	}
	return urls
}

// Download file according to the requested bucket, file, link object
func Download(req *http.Request, resp http.ResponseWriter) error {
	q := req.URL.Query()
	bucket, err := GetBucket(q.Get("b"))
	if err != nil {
		return err
	}
	return bucket.Download(nil, resp, q.Get("f"), q.Get("o"))
}

/*
Download file, if object is not empty, the file must be linked to it, otherwise an error is returned.
If RedirectPathPrefix is not empty, an location like following is required in nginx virtual server config.
	location /fs/ {
	  internal;
	  alias /data/file-storage;
	}
The location prefix and alias path should be set according to RedirectPathPrefix and Dir.
*/
func (b *Bucket) Download(db DB, resp http.ResponseWriter, file string, object string) error {
	if err := CheckHash(file); err != nil {
		return err
	}
	if object != "" {
		if err := b.EnsureLinked(db, object, file); err != nil {
			return err
		}
	}
	if err := b.writeHeader(db, resp, file); err != nil {
		return err
	}
	if b.RedirectPathPrefix != "" {
		resp.Header().Set("X-Accel-Redirect", path.Join(b.RedirectPathPrefix, b.FilePath(file)))
		return nil
	}

	f, err := os.Open(filepath.Join(b.Dir, b.FilePath(file)))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(resp, f)
	return err
}

func (b *Bucket) writeHeader(db DB, resp http.ResponseWriter, file string) error {
	row := b.getDB(db).QueryRow(
		fmt.Sprintf(`SELECT type FROM %s WHERE hash = %s`, b.FilesTable, quote(file)),
	)
	var contentType string
	if err := row.Scan(&contentType); err != nil && err != sql.ErrNoRows {
		return err
	}
	if contentType != "" {
		resp.Header().Set("Content-Type", contentType)
		resp.Header().Set("Expires", "Thu, 31 Dec 2037 23:55:55 GMT")
	}
	return nil
}

var errInvalidHash = errs.New("args-err", "invalid file hash")

// CheckHash checks if hashes is in file hash format(43 urlsafe base64 characters).
func CheckHash(hashes ...string) error {
	for _, hash := range hashes {
		if !IsHash(hash) {
			return errInvalidHash
		}
	}
	return nil
}

var hashRegexp = regexp.MustCompile(`^[\w-]{43}$`)

// IsHash returns if string s is in file hash format(43 urlsafe base64 characters).
func IsHash(s string) bool {
	return hashRegexp.MatchString(s)
}
