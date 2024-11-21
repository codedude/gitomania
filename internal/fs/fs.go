package fs

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"tig/internal/tgfile"
)

/*
How to store file snapshot history:
- Line oriented, order matters
- File snapshots in chronological order (first = oldest, last = latest)
	- Line starts with "#xxx" = start the "xxx" file declaration
	- n next lines: file snapshots in chronological order -> "hash,path"
- hash = path for now, path = relative to .tig/blobs

###FILE START
#main.go
ab42cd64ef01;ab42cd64ef01
a98bd8be9ae8;a98bd8be9ae8
da8db98b4a09;da8db98b4a09
#internal/commit/commit.go
a0e9720b207e;a0e9720b207e
###FILE END

*/

const SEPARATOR string = ";"

const tigFSIndexPath string = "index"     // File
const tigFSObjectsPath string = "objects" // Directory

type TigFileSnapshot struct {
	Hash     string           // Content hash of the file at snapshot
	Path     string           // Path of the snapshot in Tig (based on hash)
	File     *TigFile         // Never nil
	Previous *TigFileSnapshot // Never nil
}

type TigFile struct {
	Path string           // Path of the file in the client project
	Head *TigFileSnapshot // Never nil
}

type TigFileMap = map[string]*TigFile

type TigFS struct {
	Files       TigFileMap
	IndexPath   string
	ObjectsPath string
}

func New(rootDir string) (*TigFS, error) {
	cleanRootDir := path.Clean(rootDir)
	fs := &TigFS{
		Files:       make(TigFileMap, 32),
		IndexPath:   path.Join(cleanRootDir, tigFSIndexPath),
		ObjectsPath: path.Join(cleanRootDir, tigFSObjectsPath),
	}
	if err := os.Mkdir(fs.ObjectsPath, tgfile.FILE_PERM); err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}
	f, err := os.OpenFile(fs.IndexPath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fs, nil
}

func (fs TigFS) Load() error {
	lines, err := tgfile.ReadFileLimitLines(fs.IndexPath, tgfile.MAX_FILE_SIZE)
	if err != nil {
		return err
	}
	var currentFile string
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if line[0] == '#' {
			if len(line) == 1 {
				return errors.New("Bad formatting in file declaration")
			}
			currentFile = line[1:]
			fs.Files[currentFile] = &TigFile{Path: currentFile, Head: nil}
		} else {
			if len(currentFile) == 0 {
				return errors.New("File must begin with file declaration")
			}
			data := strings.Split(line, SEPARATOR)
			if len(data) != 2 {
				return errors.New("Bad hash;path formatting in snapshot list")
			}
			tigFile := fs.Files[currentFile]
			if tigFile == nil {
				return errors.New("File snapshot list must be preceeded by file declaration")
			}
			tigFile.Head = &TigFileSnapshot{
				Hash: data[0], Path: data[1], File: tigFile, Previous: tigFile.Head,
			}
		}
	}
	return nil
}

func (fs *TigFS) Add(filepath string) error {
	cleanPath := path.Clean(filepath)
	if _, ok := fs.Files[cleanPath]; ok {
		// Will be done later, since we keep track of all file history, even if not track by tig
		return errors.New("FS.ADD(): File " + cleanPath + " already exists in FS")
	} else {
		hash, err := tgfile.HashFile(cleanPath)
		if err != nil {
			return fmt.Errorf("Add getting hash: %w", err)
		}
		newTigFile := &TigFile{Path: cleanPath, Head: nil}
		newFileSnap := &TigFileSnapshot{
			Hash:     hash,
			Path:     cleanPath,
			File:     newTigFile,
			Previous: nil,
		}
		newTigFile.Head = newFileSnap

		newFileSnapPath := path.Join(fs.ObjectsPath, hash)
		err = tgfile.CopyFile(cleanPath, newFileSnapPath)
		if err != nil {
			return fmt.Errorf("Add create copy: %w", err)
		}
		fs.Files[cleanPath] = newTigFile
	}
	return nil
}

func (tgFile *TigFile) Add() error {
	return nil
}