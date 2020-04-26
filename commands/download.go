package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

type FileWriter struct {
	name        string
	once        sync.Once
	file        *os.File
	openFileErr error
}

func NewFileWriter(n string) *FileWriter {
	return &FileWriter{
		name: n,
	}
}

func (fw *FileWriter) makeUniqueName() error {
	baseName := fw.name

	counter := 1
	for {
		f, err := os.Open(fw.name)
		if os.IsNotExist(err) {
			break
		}
		defer f.Close()

		if err != nil {
			return err
		}

		fw.name = fmt.Sprintf("%s_%d", baseName, counter)
		counter++
	}

	return nil
}

func (fw *FileWriter) Write(p []byte) (int, error) {
	fw.once.Do(func() {
		fw.openFileErr = fw.makeUniqueName()
		if fw.openFileErr != nil {
			return
		}

		fw.file, fw.openFileErr = os.Create(fw.name)
		if fw.openFileErr != nil {
			return
		}
	})

	if fw.openFileErr != nil {
		return 0, fw.openFileErr
	}

	return fw.file.Write(p)
}

func (fw *FileWriter) Close() error {
	return fw.file.Close()
}

func (fw *FileWriter) FileName() string {
	return fw.name
}

type ConsoleWriter struct {
}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}

func (cw *ConsoleWriter) Write(p []byte) (int, error) {
	return fmt.Println(string(p))
}

func (cw *ConsoleWriter) Close() error {
	return nil
}

type Download struct {
	url    string
	writer io.Writer
}

func NewDownload(u string, w io.Writer) *Download {
	return &Download{
		url:    u,
		writer: w,
	}
}

func errorMessage(reader io.Reader) string {
	message, _ := ioutil.ReadAll(reader)
	return string(message)
}

func (d *Download) Do() error {
	response, err := http.Get(d.url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Unable to get file: %s", errorMessage(response.Body))
	}

	_, err = io.Copy(d.writer, response.Body)
	return err
}
