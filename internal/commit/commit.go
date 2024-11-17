package commit

import (
	"os"
	"path"
	"tig/internal/context"
	"tig/internal/fs"
	"tig/internal/tgfile"
)

type ChangeAction int

const (
	ADD ChangeAction = iota
	MODIFY
	DELETE
)

type TigChange struct {
	Action       ChangeAction
	FileSnapshot *fs.TigFileSnapshot // contains last snapshot if DELETE
}

type TigCommit struct {
	Author   string
	Msg      string
	Date     int
	Id       int
	ParentId int         // -1 on first commit, no parent
	Changes  []TigChange // contains always at least 1 Change
}

func GetOrCreateCommit(ctx context.TigCtx) (*TigCommit, error) {
	f, err := os.OpenFile(
		path.Join(ctx.RootPath, context.TigCommitFileName),
		os.O_CREATE|os.O_RDONLY,
		0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = tgfile.ReadFdLimitBytes(f, tgfile.MAX_FILE_SIZE)
	if err != nil {
		return nil, err
	}

	tigCommit := TigCommit{}
	return &tigCommit, nil
}

func Commit(ctx context.TigCtx) error {
	return nil
}
