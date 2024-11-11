package loader

/*
- .tig
	- .config: file metadata about author or user options
	- .tree: file containing data about the file system
	- .blobs: directory of files snapshots, filename = content hash
		- 09ba49bc09b40ab4c: a snapshot of a file
		- 029fea2ef09dfa09f
		- ...
*/

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"tig/internal/context"
)

// ClearSystem removes tig root folder '.tig', and all subsequent files
func ClearSystem(ctx *context.TigCtx) error {
	var err error

	if err = os.RemoveAll(ctx.RootPath); err != nil {
		return err
	}
	*ctx = context.TigCtx{}

	return nil
}

// InitSysten read values from system or global tig config
// Nothing is read inside .tig directory
func InitSystem(ctx *context.TigCtx) error {
	ctx.RootPath = path.Join(ctx.Cwd, context.TigRootPath)

	return nil
}

// CreateTigDir create directories and files needed by tig
func CreateTigDir(ctx *context.TigCtx) error {
	var err error

	if err = os.Mkdir(ctx.RootPath, 0o755); err != nil {
		if os.IsExist(err) {
			fmt.Println("tig already initialized")
			return nil
		}
		return err
	}

	if err = os.Mkdir(path.Join(ctx.RootPath, context.TigBlobsDirName), 0o775); err != nil {
		return err
	}

	fileConfig, err := os.Create(path.Join(ctx.RootPath, context.TigConfigFileName))
	defer fileConfig.Close()
	if err != nil {
		return err
	}

	fileTrack, err := os.Create(path.Join(ctx.RootPath, context.TigTrackFileName))
	defer fileTrack.Close()
	if err != nil {
		return err
	}

	fileTree, err := os.Create(path.Join(ctx.RootPath, context.TigTreeFileName))
	defer fileTree.Close()
	if err != nil {
		return err
	}

	return nil
}

// ReadFileWithLimit is the same as ReadFile, but read no more then 'limit' bytes
// cf : https://cs.opensource.google/go/go/+/refs/tags/go1.23.3:src/os/file.go;l=783
func ReadFileWithLimit(filePath string, limit int) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var size int
	if info, err := f.Stat(); err == nil {
		size64 := info.Size()
		if int64(int(size64)) == size64 {
			size = int(size64)
		}
	}
	size++ // one byte for final read at EOF

	if size > limit {
		return nil, errors.New("Can't read file bigger than " + strconv.Itoa(limit) + " bytes")
	}
	if size < 512 {
		size = 512
	}

	data := make([]byte, 0, size)
	for {
		n, err := f.Read(data[len(data):cap(data)])
		data = data[:len(data)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return data, err
		}

		if len(data) >= cap(data) {
			d := append(data[:cap(data)], 0)
			data = d[:len(data)]
		}
	}
}

func ParseFiletreeFile(ctx *context.TigCtx) error {
	var err error

	filePath := path.Join(ctx.RootPath, context.TigConfigFileName)
	_, err = ReadFileWithLimit(filePath, context.TigMaxFileRead)
	if err != nil {
		fmt.Println("Can't read config file ", filePath)
		return err
	}

	return nil
}

func GetFilesToProcessTrack(ctx context.TigCtx, files []string, mode string) error {
	var err error

	if len(files) == 0 {
		return errors.New("No file to process")
	}

	filesTracked, err := GetTrackedFileFromFile(ctx)
	if err != nil {
		return err
	}
	filesToProcess := make(map[string]bool)
	for _, file := range filesTracked {
		filesToProcess[path.Clean(file)] = true
	}
	for _, file := range files {
		file = path.Clean(file)
		if mode == "add" {
			_, err := os.Stat(file)
			if errors.Is(err, os.ErrNotExist) {
				return errors.New("File " + file + " does not exist")
			} else {
				filesToProcess[file] = true
			}
		} else {
			if !filesToProcess[file] {
				return errors.New("tig don't know about " + file)
			}
			delete(filesToProcess, file)
		}
	}

	var fileList []string
	for k := range filesToProcess {
		fileList = append(fileList, k)
	}
	if err := WriteTrackedFile(ctx, fileList); err != nil {
		return err
	}

	return nil
}

func AddFileTrack(ctx context.TigCtx, files []string) error {
	return GetFilesToProcessTrack(ctx, files, "add")
}

func RmFileTrack(ctx context.TigCtx, files []string) error {
	return GetFilesToProcessTrack(ctx, files, "rm")
}

func WriteTrackedFile(ctx context.TigCtx, filesToAdd []string) error {
	var err error
	file, err := os.Create(path.Join(ctx.RootPath, context.TigTrackFileName))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(strings.Join(filesToAdd, "\n"))
	if err != nil {
		return err
	}
	return nil
}

func GetStatus(ctx *context.TigCtx) error {
	cwdFileList, err := GetCwdTree(*ctx)
	if err != nil {
		return err
	}
	trackFileList, err := GetTrackedFileFromFile(*ctx)
	if err != nil {
		return err
	}

	trackFiles := make(map[string]*context.TigTrackStatus)
	untrackFiles := make(map[string]*context.TigTrackStatus)

	for _, v := range trackFileList {
		trackFiles[v] = &context.TigTrackStatus{FilePath: v, IsTrack: true, Exists: false}
	}
	for _, v := range cwdFileList {
		tmpFile, ok := trackFiles[v]
		if !ok {
			untrackFiles[v] = &context.TigTrackStatus{FilePath: v, IsTrack: false, Exists: true}
		} else {
			tmpFile.Exists = true
		}
	}
	fmt.Println("Track files:")
	for k, v := range trackFiles {
		var fileState string
		if v.Exists {
			fileState = "tracked"
		} else {
			fileState = "deleted"
		}
		fmt.Println("\t", fileState, "\t", k)
	}
	fmt.Println("\nUntrack files:")
	for k := range untrackFiles {
		fmt.Println("\t" + k)
	}

	return nil
}

func GetTrackedFileFromFile(ctx context.TigCtx) ([]string, error) {
	var fileList []string

	fileBytes, err := ReadFileWithLimit(path.Join(ctx.RootPath, context.TigTrackFileName), context.TigMaxFileRead)
	if err != nil {
		return nil, err
	}
	for _, v := range bytes.Split(fileBytes, []byte("\n")) {
		if len(v) > 0 {
			fileList = append(fileList, string(v))
		}
	}

	return fileList, nil
}

func GetCwdTree(ctx context.TigCtx) ([]string, error) {
	var fileList []string

	dirToWalk := []string{"."}
	for {
		lenDirToWalk := len(dirToWalk)
		if lenDirToWalk == 0 {
			break
		}
		currentDir := dirToWalk[lenDirToWalk-1]
		dirToWalk = dirToWalk[:lenDirToWalk-1]
		dirEntries, err := os.ReadDir(currentDir)
		if err != nil {
			return nil, err
		}
		for _, v := range dirEntries {
			tmpFileName := v.Name()
			// Skip tig system files
			if tmpFileName == ".tig" || tmpFileName == ".tigignore" || tmpFileName == ".git" {
				continue
			}
			if v.IsDir() {
				dirToWalk = append(dirToWalk, path.Join(currentDir, tmpFileName))
			} else {
				fileList = append(fileList, path.Join(currentDir, tmpFileName))
			}

		}
	}
	return fileList, nil
}
