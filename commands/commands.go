package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type FileInfo struct {
	Name string `json:"name"`
}

func errorMessage(reader io.Reader) string {
	message, _ := ioutil.ReadAll(reader)
	return string(message)
}

func List(url string) ([]FileInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []FileInfo{}, err
	}
	defer resp.Body.Close()

	files := make([]FileInfo, 0, 0)
	r := json.NewDecoder(resp.Body)
	err = r.Decode(&files)
	if err != nil {
		return []FileInfo{}, err
	}

	return files, nil
}

func Upload(url string, body io.Reader, contentType string) error {
	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", contentType)
	request.Close = true

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Unable to upload file: %s", errorMessage(response.Body))
	}

	return nil
}

func GetFile(url string, s Storer) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unable to get file: %s", errorMessage(response.Body))
	}

	return s.Store(response.Body)
}

func PostRequest(urlStr string, params map[string]string) error {
	/*paramList := make([]string, len(params))
	for key, value := range params {
		paramList = append(paramList, fmt.Sprintf("%s=%s", key, value))
	}

	bodyReader := strings.NewReader(strings.Join(paramList, "&"))

	request, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Close = true

	client := http.Client{}
	response, err := client.Do(request)
	*/

	postValues := url.Values{}
	for key, value := range params {
		postValues.Set(key, value)
	}
	response, err := http.PostForm(urlStr, postValues)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Unable to make post request: %s", errorMessage(response.Body))
	}

	return nil
}
