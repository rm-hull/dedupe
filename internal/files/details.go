package files

import "os"

func GetFileDetails(filename string) (*File, error) {
	fileInfo, err := os.Lstat(filename)
	if err != nil {
		return nil, err
	}

	hash, _ := Hash(filename)

	file := &File{
		Name:    filename,
		Size:    fileInfo.Size(),
		Mode:    fileInfo.Mode(),
		ModTime: fileInfo.ModTime(),
		Hash:    hash,
	}

	return file, nil
}
