package files

import "os"

func GetFileDetails(filename string) (*File, error) {
	fi, err := os.Lstat(filename)
	if err != nil {
		return nil, err
	}

	hash, _ := Hash(filename)

	file := &File{
		Name:    filename,
		Size:    fi.Size(),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
		IsDir:   fi.IsDir(),
		Hash:    hash,
	}

	return file, nil
}
