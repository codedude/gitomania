package tighistory

/*
How to store the current commit (not the final one):
- Line oriented, order does not matter
- List of file staged : action;filepath;hash

###FILE END

*/

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"tig/internal/tigconfig"
	"tig/internal/tigfile"
	"tig/internal/tigfs"
	"time"
)

// TigCommitFileName Path relative to TigRootPath
const TigCommitFileName = "commit"

// TigConfigFileName Path relative to TigRootPath
const TigTreeFileName = "tree"

type ChangeAction int

// We need persistent id, so no iota
const (
	ADD    ChangeAction = 1
	MODIFY              = 2
	DELETE              = 3
)

type TigChange struct {
	Action       ChangeAction           `json:"action"`
	FileSnapshot *tigfs.TigFileSnapshot `json:"file_snapshot"` // contains last snapshot if DELETE
}

type TigCommit struct {
	Author   string      `json:"author"`
	Msg      string      `json:"msg"`
	Date     int64       `json:"date"`
	Id       string      `json:"id"`
	ParentId string      `json:"parent_id"` // '-' on first commit, no parent
	Changes  []TigChange `json:"changes"`   // contains always at least 1 Change
}

type TigCommitTree struct {
	Head *NTree[*TigCommit]
	Tree NTree[*TigCommit]
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
func GetCurrentCommit(ctx tigconfig.TigCtx) (*TigCommit, error) {
	fd, err := tigfile.Create(path.Join(ctx.TigPath, TigCommitFileName), os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	fileLines, err := tigfile.ReadFdLines(fd, tigfile.MAX_FILE_SIZE)
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
func Commit(ctx tigconfig.TigCtx, msg string) error {
	commit, err := GetCurrentCommit(ctx)
	if err != nil {
		return err
	}
	return commit.Commit(ctx, msg)
}

func (c *TigCommit) Save(ctx tigconfig.TigCtx) error {
	var fileLines []string
	for _, change := range c.Changes {
		fileLines = append(fileLines, fmt.Sprintf("%d;%s;%s",
			change.Action, change.FileSnapshot.File.Path, change.FileSnapshot.Hash))
	}
	err := tigfile.WriteFileLines(
		path.Join(ctx.TigPath, TigCommitFileName), fileLines)
	if err != nil {
		return fmt.Errorf("Cannot save commit: %w", err)
	}
	return nil
}

func (c *TigCommit) Stage(ctx tigconfig.TigCtx, filepath string) error {
	var action ChangeAction
	var err error
	var snapshot *tigfs.TigFileSnapshot

	filepathClean := path.Clean(filepath)
	file, ok := ctx.FS.Get(filepathClean)
	if !ok {
		file, err := ctx.FS.Add(filepathClean)
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
	// In case of Stage a file already Stage (but an older version)
	// Replace it in current commit
	if c.HasFile(filepathClean) {
		c.Unstage(filepathClean)
	}
	c.Changes = append(c.Changes, TigChange{Action: action, FileSnapshot: snapshot})
	return nil
}

func (c *TigCommit) Unstage(filepath string) error {
	var i int = -1
	for k, v := range c.Changes {
		if v.FileSnapshot.File.Path == filepath {
			i = k
			break
		}
	}
	if i == -1 {
		return errors.New("Unknown file to unstage: " + filepath)
	}
	// Should we really delete it now?
	c.Changes[i].FileSnapshot.File.Delete(c.Changes[i].FileSnapshot.Hash)
	c.Changes[i] = c.Changes[len(c.Changes)-1]
	c.Changes = c.Changes[:len(c.Changes)-1]
	return nil
}

func (c *TigCommit) HasFile(filepath string) bool {
	for _, v := range c.Changes {
		if v.FileSnapshot.File.Path == filepath {
			return true
		}
	}
	return false
}

func (c *TigCommit) Reset(ctx tigconfig.TigCtx) error {
	fd, err := tigfile.Open(path.Join(ctx.TigPath, TigCommitFileName), os.O_TRUNC)
	if err != nil {
		return err
	}
	fd.Close()
	return nil
}

func (c *TigCommit) Commit(ctx tigconfig.TigCtx, msg string) error {
	// X 1- Remplir les infos du commit
	// 2- Ecrire dans le fichiers des commits
	// X 3- Reset le fichier de commit en cours
	c.Author = tigfile.B64Str(ctx.AuthorName)
	c.Date = time.Now().Unix()
	c.Msg = tigfile.B64Str(msg)
	c.ParentId = "-"
	c.Id = tigfile.HashBytes(tigfile.StrToBytes(fmt.Sprintf("%s;%d", c.Author, c.Date)))

	err := c.Reset(ctx)
	if err != nil {
		return fmt.Errorf("Commit: cannont reset commit : %w", err)
	}
	return nil
}

func LoadCommits(ctx tigconfig.TigCtx) (*TigCommitTree, error) {
	tree := TigCommitTree{}
	err := tree.Tree.Load(path.Join(ctx.TigPath, TigTreeFileName))
	if err != nil {
		return nil, fmt.Errorf("LoadCommits: %w", err)
	}
	return &tree, nil
}
