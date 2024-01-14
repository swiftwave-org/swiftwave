package dockerconfiggenerator

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Parses build arguments from dockerfile.
// Accepts a dockerfile as a string and returns a map of build arguments.
func ParseBuildArgsFromDockerfile(dockerfile string) map[string]Variable {
	variables := map[string]Variable{}

	// Extract ARG names and default values (if any)
	argPattern := `ARG\s+(\w+)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|(\S+)))?`
	re := regexp.MustCompile(argPattern)
	matches := re.FindAllStringSubmatch(dockerfile, -1)

	// Extract ARG names and default values (if any)
	for _, match := range matches {
		argName := match[1]
		defaultValue := match[2]
		if defaultValue == "" {
			defaultValue = match[3]
		}
		variables[argName] = Variable{
			Type:        argName,
			Default:     defaultValue,
			Description: argName,
		}
	}

	return variables
}

// Extract tar file to a folder.
func ExtractTar(tarFilePath string, destFolder string) error {
	reader, err := os.Open(tarFilePath)
	if err != nil {
		return err
	}

	// Create destination folder
	if _, err := os.Stat(destFolder); os.IsNotExist(err) {
		err = os.MkdirAll(destFolder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Create tar reader
	tr := tar.NewReader(reader)

	for {
		header, err := tr.Next()

		switch {
		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// clean header
		header.Name = strings.ReplaceAll(header.Name, "../", "")

		// the target location where the dir/file should be created
		target := filepath.Join(destFolder, header.Name)
		// check the file type
		switch header.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := createFileWithDirectories(target, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}

/*
Creates directories for the given file path if they do not exist.
It returns an error if it fails to create the directories.
*/
func createDirectoriesIfNotExist(filePath string) error {
	dir := filepath.Dir(filePath)
	return os.MkdirAll(dir, 0755) // 0755 sets the permissions for the new directories
}

/*
Creates a file with the given file path and file mode.
If the directories in the file path do not exist, it creates them.
Returns a pointer to the created file and any error encountered during the process.
*/
func createFileWithDirectories(filePath string, fileMode os.FileMode) (*os.File, error) {
	if err := createDirectoriesIfNotExist(filePath); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, fileMode)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Delete a directory
func deleteDirectory(dir string) {
	os.RemoveAll(dir)
}

// Check if a file exists in folder
func existsInFolder(destFolder string, file string) bool {
	path := filepath.Join(destFolder, file)
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// Check if any file exists in folder with the provided extensions
func hasFileWithExtension(destFolderPath string, extension string) bool {
	// Get all files and directories in rootPath
	entries, err := os.ReadDir(destFolderPath)
	if err != nil {
		return false
	}

	// Check if any file has the provided extension
	for _, entry := range entries {
		if entry.IsDir() {
			// Recursively check subdirectories
			if hasFileWithExtension(filepath.Join(destFolderPath, entry.Name()), extension) {
				return true
			}
		} else if strings.HasSuffix(entry.Name(), extension) {
			return true
		}
	}

	return false
}
