package main

/*
	Steps 1 (init and track): DONE
	- init: create a directory .tig with data
	- status: show tracked and untracked files
	- add: track file
	- rm: untrack file

	Steps 2 (commit):
	- Add modified/created files to the commit with "stage"
	- Remove staged files with "unstaged"
	- Commit changes with "commit"
	- List commit

	Steps 3 (revert, head):
	- Revert to a specific commit
	- Delete a commit
	- Reset head

*/

import (
	"errors"
	"fmt"
	"log"
	"os"
	"tig/internal/commit"
	"tig/internal/context"
	"tig/internal/loader"
	"tig/internal/status"
	"tig/internal/track"
)

func main() {
	fmt.Println("### Start")
	os.Exit(run(os.Args))
	fmt.Println("### Done")
}

func run(args []string) int {
	var (
		err    error
		tigCtx context.TigCtx
	)

	if len(args) < 2 {
		fmt.Println("No command")
		return 0
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Can't get cwd")
	}

	tigCtx.Cwd = cwd
	err = loader.InitSystem(&tigCtx)
	if err != nil {
		fmt.Println("Error during tig initialization: ", err)
		return 1
	}

	var arg string = args[1]
	if arg == "init" {
		err = loader.CreateTig(&tigCtx)
	} else if arg == "clear" {
		err = loader.DeleteTig(&tigCtx)
	} else if arg == "status" {
		err = status.GetStatus(&tigCtx)
	} else if arg == "add" {
		err = track.AddFileTrack(tigCtx, args[2:])
	} else if arg == "rm" {
		err = track.RmFileTrack(tigCtx, args[2:])
	} else if arg == "commit" {
		err = commit.Commit(tigCtx)
	} else {
		err = errors.New("Unknown command")
	}
	if err != nil {
		fmt.Println("Error in command ", arg, ": ", err)
		return 1
	}

	return 0
}
