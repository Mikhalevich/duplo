package main

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/Mikhalevich/argparser"
	"github.com/Mikhalevich/duplo/commands"
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
		Host:        "http://duplo",
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

func (p *Params) list() error {
	files, err := commands.List(p.listURL())
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

	files, err := commands.List(p.listURL())
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

func (p *Params) download() error {
	f := func(fileName string) error {
		var s commands.Storer
		if p.view {
			s = commands.NewConsoleStorer()
		} else {
			s = commands.NewFileStorer(fileName)
		}

		file, err := commands.GetFile(p.downloadURL(fileName), s)
		if err != nil {
			return err
		}

		if !p.view {
			fmt.Printf("Downloaded: %s\n", file)
		}
		return nil
	}

	return p.runNumberCommand(f)
}

func (p *Params) upload() error {
	if len(p.arguments) <= 0 {
		return errors.New("No files specified")
	}

	err := commands.Upload(p.uploadURL(), p.arguments)
	if err != nil {
		return err
	}

	fmt.Println("Uploaded...")
	return nil
}

func (p *Params) delete() error {
	f := func(fileName string) error {
		err := commands.PostRequest(p.deleteURL(), map[string]string{"fileName": fileName})
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

	params := make(map[string]string)
	params["title"] = strings.ReplaceAll(p.arguments[0], "/", "")
	params["body"] = p.arguments[1]

	err := commands.PostRequest(p.shareTextURL(), params)
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
		return
	}

	err = p.runCommand()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Done...")
}
