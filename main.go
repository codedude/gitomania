package main

/*
Step 1 (init and track): DONE
- init: create a directory .tig with data X
- status: show tracked and untracked files X
- add: track file X
- rm: untrack file X

Step 2 (commit):
- Add modified/created files to the commit X
- Remove staged files X
- Commit changes
- List commit

Step 3 (revert, head):
- Revert to a specific commit
- Delete a commit
- Reset head

Refactor time! Add tests!

Step 4 (branch):
- Create a branch
- Switch branch

*/

import (
	"errors"
	"fmt"
	"os"
	"tig/internal/tgcommit"
	"tig/internal/tgcontext"
	"tig/internal/tgstatus"
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
		tigCtx = tgcontext.TigCtx{AuthorName: "codedude"} // TODO: Config will be done later
	)

	if len(args) < 2 {
		fmt.Println("No command")
		return 0
	}
	var command string = args[1]

	err = tigCtx.LoadPaths()
	if err != nil {
		fmt.Println("Error during tig initialization: ", err)
		return 1
	}

	if command == "init" {
		err = tigCtx.Init()
		if err != nil {
			if errors.Is(err, tgcontext.ErrAlreadyInit) {
				fmt.Println(err)
				return 0
			}
			fmt.Println("Error in command init: ", err)
			return 1
		}
		return 0
	}

	err = tigCtx.LoadConfig()
	if err != nil {
		if errors.Is(err, tgcontext.ErrNotInit) {
			fmt.Println(err)
			return 1
		}
		fmt.Println("Error during tig configuration loading: ", err)
		return 1
	}

	err = tigCtx.LoadFS()
	if err != nil {
		fmt.Println("Error during tig initialization: ", err)
		return 1
	}
	_, err = tgcommit.LoadCommits(tigCtx)
	if err != nil {
		fmt.Printf("Error during tree initialization: %s\n", err)
		return 1
	}

	if command == "status" {
		err = tgstatus.GetStatus(&tigCtx)
	} else if command == "add" {
		err = tgstatus.AddFile(tigCtx, args[2:])
	} else if command == "rm" {
		err = tgstatus.RemoveFile(tigCtx, args[2:])
	} else if command == "commit" {
		if len(args) < 3 {
			fmt.Println("tig commit require a message argument")
			return 1
		}
		err = tgcommit.Commit(tigCtx, args[2])
	} else if command == "reset" {
		// DEV ONLY
		err = tigCtx.Delete()
		if err != nil {
			fmt.Printf("Error in command %s: %s\n", command, err)
			return 1
		}
		err = tigCtx.Init()
		if err != nil {
			fmt.Printf("Error in command %s: %s\n", command, err)
			return 1
		}
	} else {
		err = errors.New("Unknown command")
	}
	if err != nil {
		fmt.Println("Error in command ", command, ": ", err)
		return 1
	}

	return 0
}
