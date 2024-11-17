package status

import (
	"fmt"
	"tig/internal/context"
	"tig/internal/tgfile"
	"tig/internal/track"
)

func GetStatus(ctx *context.TigCtx) error {
	cwdFileList, err := tgfile.GetDirTreeFileList(".")
	if err != nil {
		return err
	}
	trackFileList, err := track.GetTrackedFile(*ctx)
	if err != nil {
		return err
	}
	untrackFiles := make(map[string]*track.TigTrackCtx)
	trackFiles := make(map[string]*track.TigTrackCtx)

	for _, v := range trackFileList {
		// By default assum they don't exists
		// And mark them Tracked when browsing cwd
		trackFiles[v] = &track.TigTrackCtx{FilePath: v, Status: track.Deleted}
	}
	for _, v := range cwdFileList {
		tmpFile, ok := trackFiles[v]
		if !ok {
			untrackFiles[v] = &track.TigTrackCtx{FilePath: v, Status: track.NotTracked}
		} else {
			tmpFile.Status = track.Tracked
		}
	}

	fmt.Println("Commit:")

	fmt.Println("\nTrack files:")
	for k, v := range trackFiles {
		var fileState string
		if v.Status == track.Tracked {
			fileState = "tracked"
		} else if v.Status == track.Deleted {
			fileState = "deleted"
		} else if v.Status == track.Modified {
			fileState = "modified"
		} else {
			fileState = ""
		}
		fmt.Println("\t", fileState, "\t", k)
	}
	fmt.Println("\nUntrack files:")
	for k := range untrackFiles {
		fmt.Println("\t" + k)
	}

	return nil
}
