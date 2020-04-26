package commands

type Text struct {
	url string
	title string
	content string
}

func NewText(u string, t string, c string) *Text {
	return &Text{
		url: u,
		title: t,
		content: c,
	}
}

func (t *Text) Do() error {
	params := make(map[string]string)
	params["title"] = t.title
	params["body"] = t.content

	return postRequest(t.url, params)
}