package tgstatus

import (
	"fmt"
	"tig/internal/tgcommit"
	"tig/internal/tgcontext"
	"tig/internal/tgfile"
)

func GetStatus(ctx *tgcontext.TigCtx) error {
	cwdFileList, err := tgfile.GetDirTree(".")
	if err != nil {
		return fmt.Errorf("Cannot get file tree: %w", err)
	}
	trackFileList, err := ReadTrackFile(*ctx)
	if err != nil {
		return fmt.Errorf("Cannot get track files: %w", err)
	}
	commit, err := tgcommit.GetCurrentCommit(*ctx)
	if err != nil {
		return fmt.Errorf("Cannot get current commit: %w", err)
	}
	untrackFiles := make([]string, 0, 40)
	trackFiles := make(map[string]bool, 32)
	commitFiles := make(map[string]bool, 16)

	for _, v := range trackFileList {
		// By default assum they don't exists
		// And mark them track when browsing cwd
		trackFiles[v] = false
	}
	for _, filePath := range cwdFileList {
		_, ok := trackFiles[filePath]
		if !ok {
			untrackFiles = append(untrackFiles, filePath)
		} else {
			trackFiles[filePath] = true
		}
	}

	fmt.Println("Commit:")
	for _, v := range commit.Changes {
		commitFiles[v.FileSnapshot.File.Path] = true
		fmt.Println(fmt.Sprintf(
			"\t%s:\t%s", tgcommit.ChangeActionToStr(v.Action), v.FileSnapshot.File.Path))
	}

	fmt.Println("\nTrack files:")
	for k, v := range trackFiles {
		var fileState string
		if v {
			hasChanged, err := ctx.FS.HasChanged(k)
			if err != nil {
				return err
			}
			if hasChanged {
				fileState = "modified"
			} else {
				if _, ok := commitFiles[k]; ok {
					continue
				}
			}

		} else {
			fileState = "delete"
		}
		if len(fileState) > 0 {
			fmt.Println(fmt.Sprintf("\t%s:\t%s", fileState, k))
		} else {
			fmt.Println(fmt.Sprintf("\t\t%s", k))
		}
	}
	fmt.Println("\nUntrack files:")
	for _, v := range untrackFiles {
		fmt.Println("\t" + v)
	}

	return nil
}
