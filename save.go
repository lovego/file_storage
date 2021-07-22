package file_storage

import (
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lovego/addrs"
)

func (s *Storage) saveFile(file multipart.File, hash string) error {
	var srcPath string
	var destPath = filepath.Join(s.ScpPath, s.FilePath(hash))
	if s.localMachine {
		s.writeFile(file, destPath)
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

func (s *Storage) writeFile(file multipart.File, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer destFile.Close()
	_, err = io.Copy(destFile, file)
	return err
}

func (s *Storage) writeTempFile(file multipart.File) (string, error) {
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

func (s *Storage) FilePath(hash string) string {
	var path string
	var i uint8
	for ; i < s.DirDepth; i++ {
		path = filepath.Join(path, hash[2*i:2*i+2])
	}
	return filepath.Join(path, hash)
}
