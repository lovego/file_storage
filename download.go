package filestorage

import (
	"errors"
	"net/http"
	"path"
	"regexp"
)

var hashRegexp = regexp.MustCompile(`^[\w-]{43}$`)

func IsHash(s string) bool {
	return hashRegexp.MatchString(s)
}

/*
Download file, if object is not empty, the file must be linked to it, otherwise an error is returned.
An location like following is required in nginx virtual server config.
	location /fs/ {
	  internal;
	  alias /data/file-storage;
	}
The location prefix should be XAccelRedirectPrefix + "/", the alias path should be ScpPath.
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
	resp.Header().Set("X-Accel-Redirect", path.Join(s.XAccelRedirectPrefix, s.FilePath(file)))

	return nil
}

var errInvalidHash = errors.New("invalid file hash")

func CheckHash(hashes ...string) error {
	for _, hash := range hashes {
		if !IsHash(hash) {
			return errInvalidHash
		}
	}
	return nil
}
