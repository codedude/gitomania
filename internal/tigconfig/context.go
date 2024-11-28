// Package tigconfig contains tig init functions
package tigconfig

import (
	"errors"
	"fmt"
	"os"
	"path"
	"tig/internal/tigfile"
	"tig/internal/tigfs"
)

// TigRootPath path relative to the current directory
const TigRootPath = ".tig"

var ErrAlreadyInit = errors.New("Tig already initialized")
var ErrNotInit = errors.New("Tig is not configured for this folder")

// TigCtx
type TigCtx struct {
	ProjectPath string
	TigPath     string
	AuthorName  string
	FS          *tigfs.TigFS
}

// InitTig initialize tig paths, must be called first.
func (ctx *TigCtx) LoadPaths() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("LoadPaths: %w", err)
	}
	ctx.ProjectPath = cwd
	ctx.TigPath = path.Join(cwd, TigRootPath)
	return nil
}

// Init create directories and files needed by tig
func (ctx *TigCtx) Init() error {
	var err error
	if err = os.Mkdir(ctx.TigPath, tigfile.DIR_PERM); err != nil {
		if os.IsExist(err) {
			return ErrAlreadyInit
		}
		return fmt.Errorf("Init: %w", err)
	}
	fd, err := tigfile.Create(path.Join(ctx.TigPath, TigConfigFileName), 0)
	if err != nil {
		return fmt.Errorf("Init: %w", err)
	}
	fd.Close()
	return nil
}

// Delete removes tig root folder .tig (DELETE ALL FILES)
func (ctx *TigCtx) Delete() error {
	return os.RemoveAll(ctx.TigPath)
}

// LoadFS initialize tig FS
func (ctx *TigCtx) LoadFS() error {
	var err error
	ctx.FS, err = tigfs.New(ctx.TigPath)
	if err != nil {
		return fmt.Errorf("LoadFS: %w", err)
	}
	err = ctx.FS.Load()
	if err != nil {
		return fmt.Errorf("LoadFS: %w", err)
	}
	return nil
}
