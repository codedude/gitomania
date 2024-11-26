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

// Default permission when creating a file
const FILE_PERM = 0o764

// Default permission when creating a directory
const DIR_PERM = 0o764

// Open is a wrapper around [os.OpenFile]. The file is open with no flag by default.
// Use when you want to Open a file for reading or writting, not for creation.
func Open(filepath string, flag int) (*os.File, error) {
	return os.OpenFile(
		filepath,
		flag,
		FILE_PERM)
}

// Create is a wrapper around [os.OpenFile]. The file is open with at least 'O_CREATE' flag.
// Use when you want to ensure file exists or is created, not when you know the file already exists beforehand.
func Create(filepath string, flag int) (*os.File, error) {
	return os.OpenFile(
		filepath,
		os.O_CREATE|flag,
		FILE_PERM)
}

// ReadFdBytes is the same as ReadFile, but read no more then 'limit' bytes
// cf : https://cs.opensource.google/go/go/+/refs/tags/go1.23.3:src/os/file.go;l=783
func ReadFdBytes(f *os.File, limit int) ([]byte, error) {
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

func ReadFdLines(f *os.File, limit int) ([]string, error) {
	fileBytes, err := ReadFdBytes(f, limit)
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

func ReadFileBytes(filePath string, limit int) ([]byte, error) {
	f, err := Open(filePath, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadFdBytes(f, limit)
}

func ReadFileLines(filePath string, limit int) ([]string, error) {
	f, err := Open(filePath, os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadFdLines(f, limit)
}

// WriteFileString write a string to a file, create it if needed (truncate)
func WriteFileString(filename string, data string) error {
	var err error
	file, err := Open(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY)
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

// WriteFileLines write a list of string (one per line) to a file, create it if needed (truncate)
func WriteFileLines(filename string, data []string) error {
	var err error
	file, err := Open(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY)
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

// WriteFileBytes write a bytes buffer to a file, create it if needed (truncate)
func WriteFileBytes(filename string, data []byte) error {
	var err error
	file, err := Create(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY)
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

func GetDirTree(rootDirPath string) ([]string, error) {
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

// HashFile return the sha1 of a file
func HashFile(filepath string) (string, error) {
	f, err := Open(filepath, os.O_RDONLY)
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

// CopyFile copy the fileSrc to fileDest, fileDest is overwritten if it exists
func CopyFile(fileSrc, fileDest string) error {
	fSrc, err := Open(fileSrc, os.O_RDONLY)
	if err != nil {
		return err
	}
	defer fSrc.Close()
	fDest, err := Create(fileDest, os.O_TRUNC|os.O_WRONLY)
	if err != nil {
		return err
	}
	defer fDest.Close()
	_, err = io.Copy(fDest, fSrc)
	if err != nil {
		return err
	}
	return nil
}
