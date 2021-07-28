package filestorage

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

var hashRegexp = regexp.MustCompile(`^[\w-]{43}$`)

// IsHash returns if string s is in file hash format(43 urlsafe base64 characters).
func IsHash(s string) bool {
	return hashRegexp.MatchString(s)
}

/*
Download file, if object is not empty, the file must be linked to it, otherwise an error is returned.
If XAccelRedirectPrefix is not empty, an location like following is required in nginx virtual server config.
	location /fs/ {
	  internal;
	  alias /data/file-storage;
	}
The location prefix and alias path should be set according to XAccelRedirectPrefix and ScpPath.
*/
func (s *Storage) Download(db DB, resp http.ResponseWriter, file string, object string) error {
	if err := CheckHash(file); err != nil {
		return err
	}
	if object != "" {
		if err := s.EnsureLinked(db, object, file); err != nil {
			return err
		}
	}
	if s.XAccelRedirectPrefix != "" {
		resp.Header().Set("X-Accel-Redirect", path.Join(s.XAccelRedirectPrefix, s.FilePath(file)))
		return nil
	}

	f, err := os.Open(filepath.Join(s.ScpPath, s.FilePath(file)))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(resp, f)
	return err
}

var errInvalidHash = errors.New("invalid file hash")

// CheckHash checks if hashes is in file hash format(43 urlsafe base64 characters).
func CheckHash(hashes ...string) error {
	for _, hash := range hashes {
		if !IsHash(hash) {
			return errInvalidHash
		}
	}
	return nil
}
