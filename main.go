package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/Mikhalevich/argparser"
	"github.com/Mikhalevich/duplo/commands"
	"github.com/Mikhalevich/iowatcher"
	"github.com/Mikhalevich/pbw"
)

type Params struct {
	Host        string `json:"host"`
	Storage     string `json:"storage"`
	command     string
	isPermanent bool
	view        bool
	arguments   []string
}

func NewParams() *Params {
	return &Params{
		Host:        "http://duplo.viberlab.com",
		Storage:     "common",
		isPermanent: false,
		view:        false,
	}
}

func join(base string, elem ...string) string {
	u, err := url.Parse(base)
	if err != nil {
		fmt.Println(err)
	}
	for _, e := range elem {
		u.Path = path.Join(u.Path, e)
	}

	resURL := u.String()
	if resURL[:len(resURL)-1] != "/" {
		resURL = fmt.Sprintf("%s/", resURL)
	}
	return resURL

}

func (p *Params) makeBaseURL() string {
	u := join(p.Host, p.Storage)
	if p.isPermanent {
		u = join(u, "permanent")
	}
	return u
}

func (p *Params) listURL() string {
	u := join(p.Host, "api", p.Storage)
	if p.isPermanent {
		u = join(u, "permanent")
	}
	return u
}

func (p *Params) uploadURL() string {
	return join(p.makeBaseURL(), "upload")
}

func (p *Params) downloadURL(fileName string) string {
	return join(p.makeBaseURL(), fileName)
}

func (p *Params) deleteURL() string {
	return join(p.makeBaseURL(), "remove")
}

func (p *Params) shareTextURL() string {
	return join(p.makeBaseURL(), "shareText")
}

func (p *Params) files() ([]commands.FileInfo, error) {
	view := commands.NewView(p.listURL())
	err := view.Do()
	if err != nil {
		return []commands.FileInfo{}, err
	}

	return view.Files(), nil
}

func (p *Params) list() error {
	files, err := p.files()
	if err != nil {
		return err
	}

	for i, f := range files {
		fmt.Printf("%d => %s\n", i+1, f.Name)
	}

	return nil
}

func (p *Params) runNumberCommand(f func(fileName string) error) error {
	if len(p.arguments) <= 0 {
		return errors.New("No files specified")
	}

	files, err := p.files()
	if err != nil {
		return err
	}

	processed := make(map[int]bool)

	for _, numStr := range p.arguments {
		number, err := strconv.Atoi(numStr)
		if err != nil {
			fmt.Printf("Invalid number %s\n", numStr)
			continue
		}

		if (number <= 0) || (number > len(files)) {
			fmt.Printf("No file with index %d\n", number)
			continue
		}

		if processed[number] {
			fmt.Printf("Already processed %d\n", number)
			continue
		}

		err = f(files[number-1].Name)
		if err != nil {
			return err
		}

		processed[number] = true
	}

	return nil
}

func (p *Params) downloadContentLength(fileName string) (int64, error) {
	response, err := http.Head(p.downloadURL(fileName))
	if err != nil {
		return 0, errors.New("Unable to determine content length")
	}
	defer response.Body.Close()
	return response.ContentLength, nil
}

func (p *Params) download() error {
	f := func(fileName string) error {
		var w io.WriteCloser
		if p.view {
			w = commands.NewConsoleWriter()
		} else {
			w = commands.NewFileWriter(fileName)
		}

		var ww io.Writer = w

		cl, err := p.downloadContentLength(fileName)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(cl)
		if cl > 0 {
			ww := iowatcher.NewWriteWatcher(w)
			notifier := make(chan int64)
			go func() {
				for v := range ww.Notifier() {
					notifier <- int64(v)
				}
				close(notifier)
			}()
			pbw.ShowWithMax(notifier, cl)
		}

		err = commands.NewDownload(p.downloadURL(fileName), ww).Do()
		if err != nil {
			return err
		}

		if !p.view {
			fmt.Printf("Downloaded: %s\n", w.(*commands.FileWriter).FileName())
		}
		return nil
	}

	return p.runNumberCommand(f)
}

func (p *Params) upload() error {
	if len(p.arguments) <= 0 {
		return errors.New("No files specified")
	}

	mr, contentType, size, err := commands.MakeMultipartReader(p.arguments)
	if err != nil {
		return err
	}

	rw := iowatcher.NewReadWatcher(mr)
	notifier := make(chan int64)
	go func() {
		for v := range rw.Notifier() {
			notifier <- int64(v)
		}
		close(notifier)
	}()
	pbw.ShowWithMax(notifier, size)

	err = commands.NewUpload(p.uploadURL(), rw, contentType).Do()
	if err != nil {
		return err
	}

	fmt.Println("Uploaded...")
	return nil
}

func (p *Params) delete() error {
	f := func(fileName string) error {
		err := commands.NewDelete(p.deleteURL(), fileName).Do()
		if err != nil {
			return err
		}
		fmt.Printf("Deleted: %s\n", fileName)
		return nil
	}

	return p.runNumberCommand(f)
}

func (p *Params) text() error {
	if len(p.arguments) < 2 {
		return errors.New("No <title> or <body> parametes provided")
	}

	err := commands.NewText(p.shareTextURL(), strings.ReplaceAll(p.arguments[0], "/", ""), p.arguments[1]).Do()
	if err != nil {
		return err
	}

	fmt.Println("Text uploaded...")
	return nil
}

func (p *Params) runCommand() error {
	switch p.command {
	case "list":
		return p.list()

	case "get":
		return p.download()

	case "push":
		return p.upload()

	case "del":
		return p.delete()

	case "text":
		return p.text()
	}

	return fmt.Errorf("Unknown commnad %s", p.command)
}

func normalizeURL(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	if u.Scheme == "" {
		baseURL = "http://" + baseURL
	}

	return baseURL, nil
}

func loadParams() (*Params, error) {
	parser := argparser.NewParser()
	host := parser.String("h", "", "host")
	storage := parser.String("s", "", "storage name to upload")
	isPermanent := parser.Bool("p", false, "user permanent storage")
	view := parser.Bool("v", false, "view file contents. Option is used for get command only")

	commands := map[string]string{"list": "get file list on current storage. Use -s parameter to specify storage name(common by default)",
		"get":  "Download files by index(see list command) from current storage. (duplo get 1 2 3, duplo -v get 4 5)",
		"push": "Upload files to current storage. (duplo push [file names])",
		"del":  "Delete files by index(see list command) from current storage. (duplo del 1 2)",
		"text": "Upload text message to current storage. (duplo text [description] [content])"}
	parser.AddCommands(commands)

	basicParams := NewParams()
	params, err, gen := parser.Parse(basicParams)

	if gen {
		return nil, errors.New("Config should be autogenerated")
	}

	p := params.(*Params)
	p.command = parser.Command()
	p.arguments = parser.Arguments()

	if *host != "" {
		u, err := normalizeURL(*host)
		if err != nil {
			return nil, err
		}
		p.Host = u
	}

	if *storage != "" {
		p.Storage = *storage
	}
	p.isPermanent = *isPermanent
	p.view = *view

	if p.Host == "" {
		return nil, errors.New("Host is invalid")
	}

	if p.Storage == "" {
		return nil, errors.New("Storage name is invalid")
	}

	return p, err
}

func main() {
	p, err := loadParams()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}

	err = p.runCommand()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}

	fmt.Println("Done...")
}
