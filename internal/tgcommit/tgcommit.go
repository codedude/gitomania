package tgcommit

/*
How to store the current commit (not the final one):
- Line oriented, order does not matter
- action;filepath;hash

###FILE END

*/

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"tig/internal/context"
	"tig/internal/fs"
	"tig/internal/tgfile"
)

type ChangeAction int

// We need persistent id, so no iota
const (
	ADD    ChangeAction = 1
	MODIFY              = 2
	DELETE              = 3
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

func ChangeActionToStr(action ChangeAction) string {
	if action == ADD {
		return "new"
	} else if action == MODIFY {
		return "modified"
	} else if action == DELETE {
		return "deleted"
	} else {
		return ""
	}
}

// GetOrCreateCommit read the commit file, create it if it does not exists
func GetOrCreateCommit(ctx context.TigCtx) (*TigCommit, error) {
	f, err := os.OpenFile(
		path.Join(ctx.RootPath, context.TigCommitFileName),
		os.O_CREATE|os.O_RDONLY,
		0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fileLines, err := tgfile.ReadFdLimitLines(f, tgfile.MAX_FILE_SIZE)
	if err != nil {
		return nil, err
	}

	commit := TigCommit{}
	for _, line := range fileLines {
		lineSplit := strings.Split(line, ";")
		if len(lineSplit) != 3 {
			return nil, errors.New("Commit file bad format, must contains 3 elem on each line")
		}
		fileAction, filePath, fileHash := lineSplit[0], lineSplit[1], lineSplit[2]
		action, err := strconv.Atoi(fileAction)
		if err != nil {
			return nil, fmt.Errorf("Commit file bad format: %w", err)
		}
		file, ok := ctx.FS.Get(filePath)
		if !ok {
			return nil, fmt.Errorf("File %s in current commit does not exist in FS", filePath)
		}
		snapshot := file.Search(fileHash)
		if snapshot == nil {
			return nil, fmt.Errorf("Bad snapshot declared for file %s", filePath)
		}
		commit.Changes = append(commit.Changes,
			TigChange{Action: ChangeAction(action), FileSnapshot: snapshot})
	}

	return &commit, nil
}

func (c *TigCommit) Save(ctx context.TigCtx) error {
	var builder strings.Builder
	for _, change := range c.Changes {
		builder.WriteString(fmt.Sprintf("%d;%s;%s\n",
			change.Action, change.FileSnapshot.File.Path, change.FileSnapshot.Hash))
	}
	err := tgfile.WriteString(
		path.Join(ctx.RootPath, context.TigCommitFileName), builder.String())
	if err != nil {
		return fmt.Errorf("Cannot save commit: %w", err)
	}
	return nil
}

func (c *TigCommit) Stage(ctx context.TigCtx, filepath string) error {
	var action ChangeAction
	var err error
	var snapshot *fs.TigFileSnapshot

	file, ok := ctx.FS.Get(filepath)
	if !ok {
		file, err := ctx.FS.Add(filepath)
		if err != nil {
			return err
		}
		snapshot = file.Head
		action = ADD
	} else {
		action = MODIFY
		snapshot, err = file.Add()
		if err != nil {
			return err
		}
	}
	c.Changes = append(c.Changes, TigChange{Action: action, FileSnapshot: snapshot})
	return nil
}
