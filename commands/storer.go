package commands

import (
	"fmt"
	"io"
	"os"
)

type Storer interface {
	Store(src io.Reader) (string, error)
}

type FileStorer struct {
	fileName string
}

func NewFileStorer(name string) *FileStorer {
	return &FileStorer{
		fileName: name,
	}
}

func (fs *FileStorer) makeUniqueName() error {
	baseName := fs.fileName

	counter := 0
	for {
		f, err := os.Open(fs.fileName)
		if os.IsNotExist(err) {
			break
		}
		defer f.Close()

		if err != nil {
			return err
		}

		fs.fileName = fmt.Sprintf("%s_%d", baseName, counter)
		counter++
	}

	return nil
}

func (fs *FileStorer) Store(src io.Reader) (string, error) {
	err := fs.makeUniqueName()
	if err != nil {
		return "", err
	}

	f, err := os.Create(fs.fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, src)
	if err != nil {
		return "", err
	}

	return fs.fileName, nil
}

type ConsoleStorer struct {
}

func NewConsoleStorer() *ConsoleStorer {
	return &ConsoleStorer{}
}

func (cs *ConsoleStorer) Store(src io.Reader) (string, error) {
	_, err := io.Copy(os.Stdout, src)
	if err != nil {
		return "", err
	}
	fmt.Println()

	return "", nil
}
