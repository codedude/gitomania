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
	FS   *TigFS           // FS link
	Path string           // Path of the file in the client project
	Head *TigFileSnapshot // Never nil
}

type TigFileMap = map[string]*TigFile

type TigFS struct {
	Files       TigFileMap
	IndexPath   string
	ObjectsPath string
}

// New initialise a new/existing FS in directory rootDir.
// Subsequent call of New wont erase local files, but the in memory representation
// of the FS can diverge, so don't call New multiple times.
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

// Load read the index file to popuplate the FS.
// Load is idempotent.
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

// Save write the FS to the index file.
// Any Add/Delete action on File or Snapshot must end with a TigFile.Save() call
func (fs *TigFS) save() error {
	var builder strings.Builder
	for _, v := range fs.Files {
		builder.WriteByte('#')
		builder.WriteString(v.Path)
		builder.WriteByte('\n')
		ptr := v.Head
		for ; ptr != nil; ptr = ptr.Previous {
			builder.WriteString(ptr.Hash)
			builder.WriteByte(';')
			builder.WriteString(ptr.Path)
			builder.WriteByte('\n')
		}
	}
	err := tgfile.WriteString(fs.IndexPath, builder.String())
	if err != nil {
		return err
	}
	return nil
}

// DeleteAll delete all data and files of the FS.
func (fs *TigFS) DeleteAll() error {
	if err := os.RemoveAll(fs.ObjectsPath); err != nil {
		return err
	}
	f, err := os.Open(fs.IndexPath)
	if err != nil {
		return err
	}
	f.Close()
	if err := os.Mkdir(fs.ObjectsPath, tgfile.FILE_PERM); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	fs.Files = make(TigFileMap, len(fs.Files)+1)
	return nil
}

// Add add a file to the FS. It also create a snapshot of the file in the FS objects directory
func (fs *TigFS) Add(filepath string) error {
	cleanPath := path.Clean(filepath)
	if _, ok := fs.Files[cleanPath]; ok {
		// Will be done later, since we keep track of all file history, even if not track by tig
		return errors.New("FS.ADD(): File " + cleanPath + " already exists in FS")
	} else {
		newTigFile := &TigFile{FS: fs, Path: cleanPath, Head: nil}
		err := newTigFile.Add()
		if err != nil {
			return fmt.Errorf("Add adding snapshot: %w", err)
		}
		fs.Files[cleanPath] = newTigFile
	}
	return fs.save()
}

// Add add a snapshot to a [TigFile]
func (tgFile *TigFile) Add() error {
	hash, err := tgfile.HashFile(tgFile.Path)
	if err != nil {
		return fmt.Errorf("Add getting hash: %w", err)
	}
	newFileSnap := &TigFileSnapshot{
		Hash:     hash,
		Path:     tgFile.Path,
		File:     tgFile,
		Previous: tgFile.Head,
	}
	err = tgfile.CopyFile(tgFile.Path, path.Join(tgFile.FS.ObjectsPath, hash))
	if err != nil {
		return fmt.Errorf("Add create copy: %w", err)
	}
	// Last so GC can clean if any error
	tgFile.Head = newFileSnap

	return nil
}
