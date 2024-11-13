package loader

/*
- .tig
	- .config: file metadata about author or user options
	- .tree: file containing data about the file system
	- .blobs: directory of files snapshots, filename = content hash
		- 09ba49bc09b40ab4c: a snapshot of a file
		- 029fea2ef09dfa09f
		- ...
*/

import (
	"fmt"
	"os"
	"path"
	"tig/internal/context"
)

// DeleteTig removes tig root folder '.tig', and all subsequent files
func DeleteTig(ctx *context.TigCtx) error {
	var err error

	if err = os.RemoveAll(ctx.RootPath); err != nil {
		return err
	}
	*ctx = context.TigCtx{}

	return nil
}

// InitSysten read values from system or global tig config
// Nothing is read inside .tig directory
func InitSystem(ctx *context.TigCtx) error {
	ctx.RootPath = path.Join(ctx.Cwd, context.TigRootPath)

	return nil
}

// CreateTig create directories and files needed by tig
func CreateTig(ctx *context.TigCtx) error {
	var err error

	if err = os.Mkdir(ctx.RootPath, 0o755); err != nil {
		if os.IsExist(err) {
			fmt.Println("tig already initialized")
			return nil
		}
		return err
	}

	if err = os.Mkdir(path.Join(ctx.RootPath, context.TigBlobsDirName), 0o775); err != nil {
		return err
	}

	fileConfig, err := os.Create(path.Join(ctx.RootPath, context.TigConfigFileName))
	defer fileConfig.Close()
	if err != nil {
		return err
	}

	fileTrack, err := os.Create(path.Join(ctx.RootPath, context.TigTrackFileName))
	defer fileTrack.Close()
	if err != nil {
		return err
	}

	fileTree, err := os.Create(path.Join(ctx.RootPath, context.TigTreeFileName))
	defer fileTree.Close()
	if err != nil {
		return err
	}

	return nil
}
