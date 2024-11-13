package main

/*
	Steps 1 (init and track): DONE
	- init: create a directory .tig with data
	- status: show tracked and untracked files
	- add: track file
	- rm: untrack file

	Steps 2 (commit):
	-

	Steps 3 (revert):
	-

*/

import (
	"errors"
	"fmt"
	"log"
	"os"
	"tig/internal/context"
	"tig/internal/loader"
	"tig/internal/status"
)

func main() { os.Exit(run(os.Args)) }

func run(args []string) int {
	fmt.Println("### Start")

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
		err = loader.CreateTigDir(&tigCtx)
	} else if arg == "clear" {
		err = loader.ClearSystem(&tigCtx)
	} else if arg == "status" {
		err = status.GetStatus(&tigCtx)
	} else if arg == "add" {
		err = status.AddFileTrack(tigCtx, args[2:])
	} else if arg == "rm" {
		err = status.RmFileTrack(tigCtx, args[2:])
	} else {
		err = errors.New("Unknown command")
	}
	if err != nil {
		fmt.Println("Error in command ", arg, ": ", err)
		return 1
	}

	fmt.Println("### Done")
	return 0
}
