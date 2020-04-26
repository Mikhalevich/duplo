package commands

import (
	"fmt"
	"net/http"
	"net/url"
)

type Doer interface {
	Do() error
}

func postRequest(urlStr string, params map[string]string) error {
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
