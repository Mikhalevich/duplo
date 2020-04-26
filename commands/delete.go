package commands

type Delete struct {
	url  string
	name string
}

func NewDelete(u string, n string) *Delete {
	return &Delete{
		url:  u,
		name: n,
	}
}

func (d *Delete) Do() error {
	return postRequest(d.url, map[string]string{"fileName": d.name})
}
