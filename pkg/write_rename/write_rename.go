package write_rename

import (
	"io"
	"os"
)

func WriteRename(filePath string, content io.Reader) error {
	var err error

	var f *os.File
	f, err = os.CreateTemp("", "")
	if err != nil {
		return err
	}
	tempFileName := f.Name()
	defer func(filename string) {
		// ignore error when removing.
		// because if WriteRename succeed, filename file will not exist.
		_ = os.Remove(filename)
	}(tempFileName)

	_, err = io.Copy(f, content)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return os.Rename(tempFileName, filePath)

}
