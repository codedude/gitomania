package tigconfig

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"tig/internal/tigfile"
)

// TigConfigFileName path relative to TigRootPath
const TigConfigFileName = "config"

// LoadConfig load the config file
// TODO
func (ctx *TigCtx) LoadConfig() error {
	ctx.AuthorName = "codedude"
	fd, err := tigfile.Open(path.Join(ctx.TigPath, TigConfigFileName), os.O_RDONLY)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrNotInit
		}
		return fmt.Errorf("LoadConfig: %w", err)
	}
	defer fd.Close()
	return nil
}
