package tigindex

/*
How to store tracked files:
- Line-oriented, orderless
- A line equal to a tracked file -> "hash,path"
- The hash from latest commit is cached, so we can check modification

###FILE START
ab42cd64ef01;main.go
a0e9720b207e;internal/commit/tighistory.go
###FILE END

*/

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"tig/internal/tigconfig"
	"tig/internal/tigfile"
	"tig/internal/tighistory"
)

// TigConfigFileName Path relative to TigRootPath
const TigTrackFileName = "track"

func getTrackedFiles(ctx tigconfig.TigCtx) ([]string, error) {
	var fileList []string

	fd, err := tigfile.Create(path.Join(ctx.TigPath, TigTrackFileName), os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	fileBytes, err := tigfile.ReadFdBytes(fd, tigfile.MAX_FILE_SIZE)
	if err != nil {
		return nil, err
	}
	for _, v := range bytes.Split(fileBytes, []byte("\n")) {
		if len(v) > 0 {
			fileList = append(fileList, string(v))
		}
	}

	return fileList, nil
}

func beforeAddRemoveFile(ctx tigconfig.TigCtx, fileList []string) (map[string]bool, *tighistory.TigCommit, error) {
	if len(fileList) == 0 {
		return nil, nil, errors.New("No file to process")
	}
	filesTracked, err := getTrackedFiles(ctx)
	if err != nil {
		return nil, nil, err
	}
	filesAll := make(map[string]bool, 32)
	for _, file := range filesTracked {
		filesAll[path.Clean(file)] = true
	}
	commit, err := tighistory.GetCurrentCommit(ctx)
	if err != nil {
		return nil, nil, err
	}
	return filesAll, commit, nil
}

func afterAddRemoveFile(ctx tigconfig.TigCtx, commit *tighistory.TigCommit, fileMap map[string]bool) error {
	err := commit.Save(ctx)
	if err != nil {
		return err
	}
	var newTrackList []string
	for k := range fileMap {
		newTrackList = append(newTrackList, k)
	}
	if err := tigfile.WriteFileLines(
		path.Join(ctx.TigPath, TigTrackFileName), newTrackList); err != nil {
		return err
	}
	return nil
}

func AddFile(ctx tigconfig.TigCtx, fileList []string) error {
	filesMap, commit, err := beforeAddRemoveFile(ctx, fileList)
	if err != nil {
		return fmt.Errorf("AddFile: %w", err)
	}
	for _, file := range fileList {
		file = path.Clean(file)
		_, err := os.Stat(file)

		if errors.Is(err, os.ErrNotExist) {
			return errors.New("File " + file + " does not exist")
		} else {
			mustStage := false
			if _, ok := filesMap[file]; !ok {
				// First time we see it, add it to track list
				filesMap[file] = true
				mustStage = true
			} else {
				fileIsModified, err := ctx.FS.HasChanged(file)
				if err != nil {
					return fmt.Errorf("AddFile: %w", err)
				}
				mustStage = fileIsModified
			}
			// Commit in both case, only if file has changed
			if mustStage {
				err = commit.Stage(ctx, file)
				if err != nil {
					return fmt.Errorf("AddFile: Cannot stage file %s: %w", file, err)
				}
			}
		}
	}
	err = afterAddRemoveFile(ctx, commit, filesMap)
	if err != nil {
		return fmt.Errorf("AddFile: %w", err)
	}
	return nil
}

func RemoveFile(ctx tigconfig.TigCtx, fileList []string) error {
	filesMap, commit, err := beforeAddRemoveFile(ctx, fileList)
	if err != nil {
		return fmt.Errorf("RemoveFile: %w", err)
	}
	for _, file := range fileList {
		file = path.Clean(file)
		if _, ok := filesMap[file]; !ok {
			return errors.New("Tig don't know about " + file)
		}
		if commit.HasFile(file) {
			err = commit.Unstage(file)
			if err != nil {
				continue
			}
		} else {
			delete(filesMap, file)
		}

	}
	err = afterAddRemoveFile(ctx, commit, filesMap)
	if err != nil {
		return fmt.Errorf("RemoveFile: %w", err)
	}
	return nil
}
