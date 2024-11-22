package main

/*
	Steps 1 (init and track): DONE
	- init: create a directory .tig with data
	- status: show tracked and untracked files
	- add: track file
	- rm: untrack file

	Steps 2 (commit):
	- Add modified/created files to the commit
	- Remove staged files
	- Commit changes
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
	"tig/internal/context"
	"tig/internal/fs"
	"tig/internal/loader"
	"tig/internal/status"
	"tig/internal/track"
)

func main() {
	fmt.Println("### Start ###")
	ret := run(os.Args)
	fmt.Println("### Done ###")
	os.Exit(ret)
}

func run(args []string) int {
	var (
		err    error
		tigCtx = context.TigCtx{AuthorName: "codedude"}
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

	// Action that dont need FS
	if arg == "init" {
		err = loader.CreateTig(&tigCtx)
		if err != nil {
			fmt.Println("Error in command ", arg, ": ", err)
			return 1
		}
		return 0
	} else if arg == "clear" {
		err = loader.DeleteTig(&tigCtx)
		if err != nil {
			fmt.Println("Error in command ", arg, ": ", err)
			return 1
		}
		return 0
	} else if arg == "reset" {
		loader.DeleteTig(&tigCtx)
		loader.CreateTig(&tigCtx)
		return 0
	}

	tigCtx.FS, err = fs.New(tigCtx.RootPath)
	if err != nil {
		fmt.Println("Error during fs initialization: ", err)
		return 1
	}
	err = tigCtx.FS.Load()
	if err != nil {
		fmt.Println("Error during fs loading: ", err)
		return 1
	}

	// Action that need FS
	if arg == "status" {
		err = status.GetStatus(&tigCtx)
	} else if arg == "add" {
		err = track.AddFileTrack(tigCtx, args[2:])
	} else if arg == "rm" {
		err = track.RmFileTrack(tigCtx, args[2:])
	} else if arg == "commit" {
		// err = tgcommit.Commit(tigCtx)
	} else {
		err = errors.New("Unknown command")
	}
	if err != nil {
		fmt.Println("Error in command ", arg, ": ", err)
		return 1
	}

	return 0
}
