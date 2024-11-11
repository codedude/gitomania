package main

/*
	Steps 1:

	- init: create a directory .tig with data
	- status: show tracked and untracked files
	- add: track file
	- rm: untrack file

*/

import (
	"fmt"
	"log"
	"os"
	"tig/internal/context"
	"tig/internal/loader"
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
		if err != nil {
			fmt.Println("Error while creating system: ", err)
			return 1
		}
	} else if arg == "reset" {
		// helper for dev only
		fmt.Println("reset")
		err = loader.ClearSystem(&tigCtx)
		if err != nil {
			fmt.Println("Error while clearing system: ", err)
			return 1
		}
		err = loader.InitSystem(&tigCtx)
		if err != nil {
			fmt.Println("Error while initiating system: ", err)
			return 1
		}
		err = loader.CreateTigDir(&tigCtx)
		if err != nil {
			fmt.Println("Error while creating system: ", err)
			return 1
		}
	} else if arg == "clear" {
		err = loader.ClearSystem(&tigCtx)
		if err != nil {
			fmt.Println("Error while clearing system: ", err)
			return 1
		}
	} else if arg == "status" {
		err = loader.GetStatus(&tigCtx)
		if err != nil {
			fmt.Println("Error while getting status: ", err)
			return 1
		}
	} else if arg == "add" {
		err = loader.AddFileTrack(tigCtx, args[2:])
		if err != nil {
			fmt.Println("Error while adding files: ", err)
			return 1
		}
	} else if arg == "rm" {
		err = loader.RmFileTrack(tigCtx, args[2:])
		if err != nil {
			fmt.Println("Error while removing files: ", err)
			return 1
		}
	} else {
		fmt.Println("Unknown command")
	}

	fmt.Println("### Done")
	return 0
}
