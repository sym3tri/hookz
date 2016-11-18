package server

type Endpoint struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

func (e Endpoint) Sanitize() Endpoint {
	e.Token = ""
	return e
}
