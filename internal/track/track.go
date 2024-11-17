package track

import (
	"bytes"
	"errors"
	"os"
	"path"
	"tig/internal/commit"
	"tig/internal/context"
	"tig/internal/tgfile"
)

type TigTrackStatus int

const (
	NotTracked TigTrackStatus = iota
	Tracked
	Deleted
	Modified
	Added
)

type TigTrackCtx struct {
	IsTrack  bool
	Status   TigTrackStatus
	FilePath string // Relative to [TigRootPath]
}

func GetTrackedFile(ctx context.TigCtx) ([]string, error) {
	var fileList []string

	fileBytes, err := tgfile.ReadFileLimitBytes(
		path.Join(ctx.RootPath, context.TigTrackFileName), context.TigMaxFileRead)
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

func GetFilesToProcessTrack(ctx context.TigCtx, filesReq []string, mode string) error {
	if len(filesReq) == 0 {
		return errors.New("No file to process")
	}

	filesTracked, err := GetTrackedFile(ctx)
	if err != nil {
		return err
	}
	_, err = commit.GetOrCreateCommit(ctx)
	if err != nil {
		return err
	}

	filesToProcess := make(map[string]TigTrackStatus)
	for _, file := range filesTracked {
		filesToProcess[path.Clean(file)] = Tracked
	}
	for _, file := range filesReq {
		file = path.Clean(file)
		if mode == "add" {
			_, err := os.Stat(file)
			if errors.Is(err, os.ErrNotExist) {
				return errors.New("File " + file + " does not exist")
			} else {
				if _, ok := filesToProcess[file]; !ok {
					// First time we see it
					filesToProcess[file] = Added
				}
				// Default = Tracked

				// Commit in both case
			}
		} else {
			if _, ok := filesToProcess[file]; !ok {
				return errors.New("tig don't know about " + file)
			}
			// if added/same -> untrack (delete)
			// if modified/deleted -> tracked (add to file)
			filesToProcess[file] = Tracked
			delete(filesToProcess, file)

		}
	}

	var fileList []string
	for k := range filesToProcess {
		fileList = append(fileList, k)
	}
	if err := tgfile.WriteStrings(
		path.Join(ctx.RootPath, context.TigTrackFileName), fileList); err != nil {
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
