package tgfile

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

// In bytes, last number = Mo
const MAX_FILE_SIZE = 1024 * 1024 * 64
const FILE_PERM = 0o764

// ReadFdLimitBytes is the same as ReadFile, but read no more then 'limit' bytes
// cf : https://cs.opensource.google/go/go/+/refs/tags/go1.23.3:src/os/file.go;l=783
func ReadFdLimitBytes(f *os.File, limit int) ([]byte, error) {
	var size int
	if info, err := f.Stat(); err == nil {
		size64 := info.Size()
		if int64(int(size64)) == size64 {
			size = int(size64)
		}
	}
	size++ // one byte for final read at EOF
	if size > limit {
		return nil, errors.New("Can't read file bigger than " + strconv.Itoa(limit) + " bytes")
	}
	if size < 512 {
		size = 512
	}
	data := make([]byte, 0, size)
	for {
		n, err := f.Read(data[len(data):cap(data)])
		data = data[:len(data)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return data, err
		}
		if len(data) >= cap(data) {
			d := append(data[:cap(data)], 0)
			data = d[:len(data)]
		}
	}
}

func ReadFdLimitLines(f *os.File, limit int) ([]string, error) {
	fileBytes, err := ReadFdLimitBytes(f, limit)
	if err != nil {
		return nil, err
	}
	var bufStrLines = []string{}
	for _, v := range bytes.Split(fileBytes, []byte("\n")) {
		if len(v) > 0 {
			bufStrLines = append(bufStrLines, string(v))
		}
	}
	return bufStrLines, nil
}

func ReadFileLimitBytes(filePath string, limit int) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadFdLimitBytes(f, limit)
}

func ReadFileLimitLines(filePath string, limit int) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadFdLimitLines(f, limit)
}

func WriteStrings(filename string, data []string) error {
	var err error
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var builder strings.Builder
	for i := 0; i < len(data); i++ {
		builder.WriteString(data[i])
		builder.WriteString("\n")
	}
	_, err = file.WriteString(builder.String())
	if err != nil {
		return err
	}
	return nil
}

func WriteString(filename string, data string) error {
	var err error
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(data)
	if err != nil {
		return err
	}
	return nil
}

func WriteBytes(filename string, data []byte) error {
	var err error
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func GetDirTreeFileList(rootDirPath string) ([]string, error) {
	var fileList []string

	dirToWalk := []string{rootDirPath}
	for {
		lenDirToWalk := len(dirToWalk)
		if lenDirToWalk == 0 {
			break
		}
		currentDir := dirToWalk[lenDirToWalk-1]
		dirToWalk = dirToWalk[:lenDirToWalk-1]
		dirEntries, err := os.ReadDir(currentDir)
		if err != nil {
			return nil, err
		}
		for _, v := range dirEntries {
			tmpFileName := v.Name()
			// Skip tig system files
			if tmpFileName == ".tig" || tmpFileName == ".tigignore" || tmpFileName == ".git" {
				continue
			}
			if v.IsDir() {
				dirToWalk = append(dirToWalk, path.Join(currentDir, tmpFileName))
			} else {
				fileList = append(fileList, path.Join(currentDir, tmpFileName))
			}

		}
	}
	return fileList, nil
}

func HashBytes(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func HashFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func CopyFile(fileSrc, fileDest string) error {
	fSrc, err := os.Open(fileSrc)
	if err != nil {
		return err
	}
	fDest, err := os.Create(fileDest)
	if err != nil {
		return err
	}
	_, err = io.Copy(fDest, fSrc)
	if err != nil {
		return err
	}
	return nil
}