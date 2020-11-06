package saving

import (
	"os"
)

// FolderPath is the path of the folder where the replays are located.
const FolderPath = "." + string(os.PathSeparator) + "replays-storage"

// RetrieveReplays returns a string array which will
// contain all the zipped replay names.
func RetrieveReplays() []string {
	folder, err := os.Open(FolderPath)
	defer folder.Close()

	if err != nil {
		return nil
	}

	files, err := folder.Readdirnames(-1)

	return files
}

// DeleteReplay tries to delete a zipped file.
// A bool will be returned determining the file was deleted or not.
func DeleteReplay(replayID string) bool {
	path := FolderPath + string(os.PathSeparator) + replayID
	err := os.Remove(path)

	if err != nil {
		return false
	}

	return true
}

// CreateLocalDatabase creates the initial folder
// which contains all the zipped replays.
// An error will be returned if an I/O Exception occurred.
func CreateLocalDatabase() error {
	folderNotExists := fileNotExists(FolderPath)

	if folderNotExists {
		return os.Mkdir(FolderPath, 0755)
	}

	return nil
}

func fileNotExists(filePath string) bool {
	_, err := os.Stat(filePath)

	return os.IsNotExist(err)
}
