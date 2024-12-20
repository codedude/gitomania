package tigfs

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
	"tig/internal/tigfile"
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
#internal/commit/tighistory.go
a0e9720b207e;a0e9720b207e
###FILE END

*/

const SEPARATOR string = ";"

const tigFSIndexFileName string = "_index" // File (_file so it's first in filetree list)
const tigFSPath string = "fs"              // Directory

type TigFileSnapshot struct {
	Hash     string   // Content hash of the file at snapshot
	Path     string   // Path of the snapshot in Tig (based on hash)
	File     *TigFile // Never nil
	Previous *TigFileSnapshot
	Next     *TigFileSnapshot
}

type TigFile struct {
	FS   *TigFS           // FS link
	Path string           // Path of the file in the client project
	Head *TigFileSnapshot // Never nil
}

type TigFileMap = map[string]*TigFile

type TigFS struct {
	Files     TigFileMap
	IndexPath string
	DirPath   string
}

// New initialise a new/existing FS in directory rootDir.
// Do not call multiple times
func New(rootDir string) (*TigFS, error) {
	cleanRootDir := path.Clean(rootDir)
	cleanFSPath := path.Join(cleanRootDir, tigFSPath)
	fs := &TigFS{
		Files:     make(TigFileMap, 32),
		IndexPath: path.Join(cleanFSPath, tigFSIndexFileName),
		DirPath:   cleanFSPath,
	}
	if err := os.Mkdir(fs.DirPath, tigfile.FILE_PERM); err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}
	fd, err := tigfile.Create(fs.IndexPath, 0)
	if err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}
	defer fd.Close()
	return fs, nil
}

// Load read the index file to popuplate the FS.
// Load is idempotent.
func (fs *TigFS) Load() error {
	lines, err := tigfile.ReadFileLines(fs.IndexPath, tigfile.MAX_FILE_SIZE)
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
			fs.Files[currentFile] = &TigFile{Path: currentFile, FS: fs, Head: nil}
		} else {
			if len(currentFile) == 0 {
				return errors.New("File must begin with file declaration")
			}
			data := strings.Split(line, SEPARATOR)
			if len(data) != 2 {
				return errors.New("Bad hash;path formatting in snapshot list")
			}
			tigFile, ok := fs.Files[currentFile]
			if !ok {
				return errors.New("File snapshot list must be preceeded by file declaration")
			}
			tigFile.Head = &TigFileSnapshot{
				Hash: data[0], Path: data[1], File: tigFile, Previous: tigFile.Head,
			}
			if tigFile.Head.Previous != nil {
				tigFile.Head.Previous.Next = tigFile.Head
			}
		}
	}
	return nil
}

// Save write the FS to the index file.
// Any Add/Delete action on File or Snapshot must end with a TigFile.Save() call
func (fs *TigFS) save() error {
	var fileLines []string
	for _, v := range fs.Files {
		if v.Head == nil {
			continue // No snapshot, dont save it
		}
		fileLines = append(fileLines, fmt.Sprintf("#%s", v.Path))
		ptr := v.Head
		var snapLines []string
		for ; ptr != nil; ptr = ptr.Previous {
			snapLines = append(snapLines, fmt.Sprintf("%s;%s", ptr.Hash, ptr.Path))
		}
		slices.Reverse(snapLines)
		fileLines = append(fileLines, snapLines...)
	}
	err := tigfile.WriteFileLines(fs.IndexPath, fileLines)
	if err != nil {
		return err
	}
	return nil
}

// Get return a File in the FS
func (fs *TigFS) Get(filepath string) (*TigFile, bool) {
	file, ok := fs.Files[path.Clean(filepath)]
	return file, ok
}

// FileIsModified check if the file filepath has been modified/added since the last snapshot
func (fs *TigFS) HasChanged(filepath string) (bool, error) {
	file, ok := fs.Get(filepath)
	if !ok {
		return true, nil
	}
	newHash, err := tigfile.HashFile(filepath)
	if err != nil {
		return false, err
	}
	return newHash != file.Head.Hash, nil
}

// Add add a file to the FS. It also create a snapshot of the file in the FS objects directory
func (fs *TigFS) Add(filepath string) (*TigFile, error) {
	cleanPath := path.Clean(filepath)
	var newTigFile *TigFile
	if _, ok := fs.Files[cleanPath]; ok {
		// Will be done later, since we keep track of all file history, even if not track by tig
		return nil, errors.New("FS.ADD(): File " + cleanPath + " already exists in FS")
	} else {
		newTigFile = &TigFile{FS: fs, Path: cleanPath, Head: nil}
		_, err := newTigFile.Add()
		if err != nil {
			return nil, fmt.Errorf("Add adding snapshot: %w", err)
		}
		fs.Files[cleanPath] = newTigFile
	}
	return newTigFile, fs.save()
}

// Add add a snapshot to a [TigFile]
func (file *TigFile) Add() (*TigFileSnapshot, error) {
	hash, err := tigfile.HashFile(file.Path)
	if err != nil {
		return nil, fmt.Errorf("Add getting hash: %w", err)
	}
	newFileSnap := &TigFileSnapshot{
		Hash:     hash,
		Path:     hash, // Path = hash for now
		File:     file,
		Previous: file.Head,
	}
	err = tigfile.CopyFile(file.Path, path.Join(file.FS.DirPath, hash))
	if err != nil {
		return nil, fmt.Errorf("Add create copy: %w", err)
	}
	// Last so GC can clean if any error
	file.Head = newFileSnap
	if newFileSnap.Previous != nil {
		newFileSnap.Previous.Next = newFileSnap
	}
	return newFileSnap, file.FS.save()
}

// Delete delete a snapshot of a [File]
func (file *TigFile) Delete(hash string) error {
	snapshot := file.Search(hash)
	if snapshot == nil {
		return errors.New("The snapshot does not exist for deletion")
	}
	if snapshot.Previous != nil {
		snapshot.Previous.Next = snapshot.Next
	} else {
		if snapshot.Next != nil {
			snapshot.Next.Previous = nil
		}
	}
	if snapshot.Next == nil {
		file.Head = snapshot.Previous
	}
	snapshot.Next = nil
	snapshot.Previous = nil
	err := os.Remove(path.Join(file.FS.DirPath, snapshot.Path))
	if err != nil {
		return fmt.Errorf("Cannot delete snapshot: %w", err)
	}
	// Dont delete file object if no snapshot remain, we delete in on save
	return file.FS.save()
}

// Search search for a specifi snapshot
func (file *TigFile) Search(hash string) *TigFileSnapshot {
	for ptr := file.Head; ptr != nil; ptr = ptr.Previous {
		if ptr.Hash == hash {
			return ptr
		}
	}
	return nil
}
