package track

/*
How to store tracked files:
- Line-oriented, orderless
- A line equal to a tracked file -> "hash,path"
- The hash from latest commit is cached, so we can check modification

###FILE START
ab42cd64ef01;main.go
a0e9720b207e;internal/commit/tgcommit.go
###FILE END

*/

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"tig/internal/context"
	"tig/internal/tgcommit"
	"tig/internal/tgfile"
)

// TigConfigFileName Path relative to TigRootPath
const TigTrackFileName = "track"

func ReadTrackFile(ctx context.TigCtx) ([]string, error) {
	var fileList []string

	fileBytes, err := tgfile.ReadFileLimitBytes(
		path.Join(ctx.RootPath, TigTrackFileName), context.TigMaxFileRead)
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

func GetFilesToProcessTrack(ctx context.TigCtx, filesToAdd []string, mode string) error {
	if len(filesToAdd) == 0 {
		return errors.New("No file to process")
	}

	filesTracked, err := ReadTrackFile(ctx)
	if err != nil {
		return err
	}
	commit, err := tgcommit.GetCurrentCommit(ctx)
	if err != nil {
		return err
	}

	filesAll := make(map[string]bool, 32)
	for _, file := range filesTracked {
		filesAll[path.Clean(file)] = true
	}
	for _, file := range filesToAdd {
		file = path.Clean(file)
		_, err := os.Stat(file)
		if mode == "add" {
			if errors.Is(err, os.ErrNotExist) {
				return errors.New("File " + file + " does not exist")
			} else {
				if _, ok := filesAll[file]; !ok {
					// First time we see it, add it to track list
					filesAll[file] = true
				}
				// Commit in both case, only if file has changed
				fileIsModified, err := ctx.FS.HasChanged(file)
				if err != nil {
					return err
				}
				if fileIsModified {
					err = commit.Stage(ctx, file)
					if err != nil {
						return fmt.Errorf("Cannot stage file %s: %w", file, err)
					}
				}
			}
		} else {
			if _, ok := filesAll[file]; !ok {
				return errors.New("tig don't know about " + file)
			}
			if commit.HasFile(file) {
				commit.Unstage(file)
			} else {
				delete(filesAll, file)
			}
		}
	}

	err = commit.Save(ctx)
	if err != nil {
		return err
	}

	var fileList []string
	for k := range filesAll {
		fileList = append(fileList, k)
	}
	if err := tgfile.WriteStrings(
		path.Join(ctx.RootPath, TigTrackFileName), fileList); err != nil {
		return err
	}

	return nil
}

func AddFileTrack(ctx context.TigCtx, files []string) error {
	return GetFilesToProcessTrack(ctx, files, "add")
}

func RmFileTrack(ctx context.TigCtx, files []string) error {
	return GetFilesToProcessTrack(ctx, files, "rm")
}
