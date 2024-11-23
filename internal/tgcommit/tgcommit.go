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
	"time"
)

// Path relative to TigRootPath
const TigCommitFileName = "commit"

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
	Date     int64
	Id       string
	ParentId string      // '-' on first commit, no parent
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

// GetCurrentCommit read the commit file, or create it if it does not exists
func GetCurrentCommit(ctx context.TigCtx) (*TigCommit, error) {
	f, err := os.OpenFile(
		path.Join(ctx.RootPath, TigCommitFileName),
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

// Commit get the current commit and commit it
func Commit(ctx context.TigCtx, msg string) error {
	commit, err := GetCurrentCommit(ctx)
	if err != nil {
		return err
	}
	return commit.Commit(ctx, msg)
}

func (c *TigCommit) Save(ctx context.TigCtx) error {
	var fileLines []string
	for _, change := range c.Changes {
		fileLines = append(fileLines, fmt.Sprintf("%d;%s;%s",
			change.Action, change.FileSnapshot.File.Path, change.FileSnapshot.Hash))
	}
	err := tgfile.WriteStrings(
		path.Join(ctx.RootPath, TigCommitFileName), fileLines)
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

func (c *TigCommit) Unstage(filepath string) {
	var i int
	for k, v := range c.Changes {
		if v.FileSnapshot.File.Path == filepath {
			i = k
			break
		}
	}
	c.Changes[i] = c.Changes[len(c.Changes)-1]
	c.Changes = c.Changes[:len(c.Changes)-1]
}

func (c *TigCommit) HasFile(filepath string) bool {
	for _, v := range c.Changes {
		if v.FileSnapshot.File.Path == filepath {
			return true
		}
	}
	return false
}

func (c *TigCommit) Commit(ctx context.TigCtx, msg string) error {
	// X 1- Remplir les infos du commit
	// 2- Ecrire dans le fichiers des commits
	// 3- Reset le fichier de commit en cours
	c.Author = tgfile.B64Str(ctx.AuthorName)
	c.Date = time.Now().Unix()
	c.Msg = tgfile.B64Str(msg)
	c.ParentId = "-"
	c.Id = tgfile.HashBytes(tgfile.StrToBytes(fmt.Sprintf("%s;%d", c.Author, c.Date)))

	return nil
}
