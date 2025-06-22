package types

type Import struct {
	Name   string `json:"name"`
	Alias  string `json:"alias,omitempty"`
	Path   string `json:"path,omitempty"`
	Ignore bool   `json:"ignore,omitempty"`
}
