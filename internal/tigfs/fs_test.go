package tigfs

import (
	"bytes"
	"path"
	"strconv"
	"testing"
	"tig/internal/tigfile"
)

// generateFakeFs generate a known FS, no files are written
func generateFakeFs(t *testing.T, fileList []string) (*TigFS, string) {
	tmpDirPath := t.TempDir()
	fileContent := []byte("Hello world! xxx")
	newFs, err := New(tmpDirPath)
	if err != nil {
		t.Fatalf("Error GenerateFakeFs New TigFS: %s", err)
	}
	var indexFile []string
	for i, filePath := range fileList {
		indexFile = append(indexFile, "#"+filePath)
		tigFile := &TigFile{Path: filePath, Head: nil}
		newFs.Files[filePath] = tigFile
		// Generate multiple snapshots
		for nOfSnaps := i; nOfSnaps < i+1; nOfSnaps++ {
			tmpFileContent := bytes.Replace(fileContent, []byte("xxx"), []byte(strconv.Itoa(nOfSnaps)), -1)
			hash := tigfile.HashBytes(tmpFileContent)
			indexFile = append(indexFile, hash+";"+hash)
			fileSnap := &TigFileSnapshot{Hash: hash, Path: hash, File: tigFile, Previous: tigFile.Head}
			tigFile.Head = fileSnap
		}
	}
	err = tigfile.WriteFileLines(newFs.IndexPath, indexFile)
	if err != nil {
		t.Fatalf("Error GenerateFakeFs write index: %s", err)
	}

	return newFs, tmpDirPath
}

var fakeFileListMulti = []string{
	"file1.py",
	"a/file2.c",
	"a/file3.zig",
	"b/file4.go",
	"c/d/file5.gleam",
	"a/c/d/file2.c",
}

func TestFSLoad(t *testing.T) {
	fs_generated, fs_path := generateFakeFs(t, fakeFileListMulti)

	fs_to_test, err := New(fs_path)
	if err != nil {
		t.Fatalf("Error New fs_to_test: %s", err)
	}
	if err := fs_to_test.Load(); err != nil {
		t.Fatalf("Error LoadTigFs: %s", err)
	}

	if len(fs_generated.Files) != len(fs_to_test.Files) {
		t.Fatalf("Map mismatch in len: %d != %d", len(fs_generated.Files), len(fs_to_test.Files))
	}
	for k, tgGen := range fs_generated.Files {
		if tgTest, ok := fs_to_test.Files[k]; ok {
			if tgGen.Path != tgTest.Path {
				t.Fatalf("Map mismatch in value Path: %s != %s", tgGen.Path, tgTest.Path)
			}
			snapGen := tgGen.Head
			snapTest := tgTest.Head
			if snapTest == nil {
				t.Fatalf("Map mismatch in snapshot: no snapshot for %s", tgGen.Path)
			}
			for {
				if snapTest == nil && snapGen != nil {
					t.Fatalf("Map mismatch in snapshot: missing some snap for %s", tgGen.Path)
				}
				if snapTest != nil && snapGen == nil {
					t.Fatalf("Map mismatch in snapshot: too much snap for %s", tgGen.Path)
				}
				if snapTest == nil && snapGen == nil {
					break
				}
				if snapGen.Hash != snapTest.Hash {
					t.Fatalf("Map mismatch in snapshot hash: %s != %s", snapGen.Hash, snapTest.Hash)
				}
				if snapGen.Path != snapTest.Path {
					t.Fatalf("Map mismatch in snapshot path: %s != %s", snapGen.Path, snapTest.Path)
				}
				snapGen = snapGen.Previous
				snapTest = snapTest.Previous
			}
		} else {
			t.Fatalf("Map mismatch in key: %s", k)
		}
	}
}

func TestFSAdd(t *testing.T) {
	tmpDirPath := t.TempDir()
	// We can skip fs.Load() since it's an empty FS for now
	fs_to_test, err := New(tmpDirPath)
	if err != nil {
		t.Fatalf("Error New fs_to_test: %s", err)
	}
	fileToAdd := "hello.go"
	fullFilePath := path.Join(tmpDirPath, fileToAdd)
	if err := tigfile.WriteFileString(fullFilePath, "Hello world"); err != nil {
		t.Fatalf("Error file WriteString: %s", err)
	}
	if _, err = fs_to_test.Add(fullFilePath); err != nil {
		t.Fatalf("Error AddFile: %s", err)
	}
	if len(fs_to_test.Files) != 1 {
		t.Fatalf("FS must be of size 1, not %d (1)", len(fs_to_test.Files))
	}

	// Reset FS
	for k := range fs_to_test.Files {
		delete(fs_to_test.Files, k)
	}
	// Now Load should show the previous added file
	if err := fs_to_test.Load(); err != nil {
		t.Fatalf("Error Load: %s", err)
	}
	if len(fs_to_test.Files) != 1 {
		t.Fatalf("FS must be of size 1, not %d (2)", len(fs_to_test.Files))
	}

	v, ok := fs_to_test.Files[fullFilePath]
	if !ok {
		t.Fatalf("File must have key: %s", fullFilePath)
	} else {
		if v.Path != fullFilePath {
			t.Fatalf("File must have Path: %s", fullFilePath)
		}
	}

}
