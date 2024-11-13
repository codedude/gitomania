package status

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"tig/internal/context"
	"tig/internal/file"
)

func GetFilesToProcessTrack(ctx context.TigCtx, files []string, mode string) error {
	var err error

	if len(files) == 0 {
		return errors.New("No file to process")
	}

	filesTracked, err := GetTrackedFileFromFile(ctx)
	if err != nil {
		return err
	}
	filesToProcess := make(map[string]bool)
	for _, file := range filesTracked {
		filesToProcess[path.Clean(file)] = true
	}
	for _, file := range files {
		file = path.Clean(file)
		if mode == "add" {
			_, err := os.Stat(file)
			if errors.Is(err, os.ErrNotExist) {
				return errors.New("File " + file + " does not exist")
			} else {
				filesToProcess[file] = true
			}
		} else {
			if !filesToProcess[file] {
				return errors.New("tig don't know about " + file)
			}
			delete(filesToProcess, file)
		}
	}

	var fileList []string
	for k := range filesToProcess {
		fileList = append(fileList, k)
	}
	if err := file.WriteStrings(path.Join(ctx.RootPath, context.TigTrackFileName), fileList); err != nil {
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

func GetStatus(ctx *context.TigCtx) error {
	cwdFileList, err := file.GetDirTreeFileList(".")
	if err != nil {
		return err
	}
	trackFileList, err := GetTrackedFileFromFile(*ctx)
	if err != nil {
		return err
	}

	trackFiles := make(map[string]*context.TigTrackStatus)
	untrackFiles := make(map[string]*context.TigTrackStatus)

	for _, v := range trackFileList {
		trackFiles[v] = &context.TigTrackStatus{FilePath: v, IsTrack: true, Exists: false}
	}
	for _, v := range cwdFileList {
		tmpFile, ok := trackFiles[v]
		if !ok {
			untrackFiles[v] = &context.TigTrackStatus{FilePath: v, IsTrack: false, Exists: true}
		} else {
			tmpFile.Exists = true
		}
	}
	fmt.Println("Track files:")
	for k, v := range trackFiles {
		var fileState string
		if v.Exists {
			fileState = "tracked"
		} else {
			fileState = "deleted"
		}
		fmt.Println("\t", fileState, "\t", k)
	}
	fmt.Println("\nUntrack files:")
	for k := range untrackFiles {
		fmt.Println("\t" + k)
	}

	return nil
}

func GetTrackedFileFromFile(ctx context.TigCtx) ([]string, error) {
	var fileList []string

	fileBytes, err := file.ReadFileLimitBytes(
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
