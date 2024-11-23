// Package context contains tig system types and constants
package context

import "tig/internal/fs"

const TigMaxFileRead = 1024 * 1024 * 64

// Path relative to the current directory
const TigRootPath = ".tig"

// TigConfigFileName Path relative to TigRootPath
const TigConfigFileName = "config"

// TigConfigFileName Path relative to TigRootPath
const TigTreeFileName = "tree"

// TigCtx
type TigCtx struct {
	Cwd        string
	RootPath   string
	AuthorName string
	FS         *fs.TigFS
}
