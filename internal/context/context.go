// Package context contains tig system types and constants
package context

const TigMaxFileRead = 1024 * 1024 * 64

// Path relative to the current directory
const TigRootPath = ".tig"

// TigConfigFileName Path relative to TigRootPath
const TigConfigFileName = "config"

// TigConfigFileName Path relative to TigRootPath
const TigTrackFileName = "track"

// Path relative to TigRootPath
const TigIndexFileName = "index"

// Path relative to TigRootPath
const TigCommitFileName = "commit"

// Path relative to TigRootPath
const TigBlobsDirName = "blobs"

// TigCtx
type TigCtx struct {
	Cwd        string
	RootPath   string
	AuthorName string
}
