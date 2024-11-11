// Package context contains types and constants of the tig system
package context

const TigMaxFileRead = 1024 * 1024 * 32

// Path relative to the current directory
const TigRootPath = ".tig"

// TigConfigFileName Path relative to TigRootPath
const TigConfigFileName = "config"

// TigConfigFileName Path relative to TigRootPath
const TigTrackFileName = "track"

// Path relative to TigRootPath
const TigTreeFileName = "tree"

// Path relative to TigRootPath
const TigBlobsDirName = "blobs"

type TigTrackStatus struct {
	IsTrack  bool
	Exists   bool
	FilePath string
}

// TigCtx
type TigCtx struct {
	Cwd        string
	RootPath   string
	AuthorName string
}
