package commands

import (
	"encoding/json"
	"net/http"
)

type FileInfo struct {
	Name string `json:"name"`
}

type View struct {
	url   string
	files []FileInfo
}

func NewView(u string) *View {
	return &View{
		url: u,
	}
}

func (v *View) Do() error {
	response, err := http.Get(v.url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	r := json.NewDecoder(response.Body)
	err = r.Decode(&v.files)
	if err != nil {
		return err
	}

	return nil
}

func (v *View) Files() []FileInfo {
	return v.files
}
