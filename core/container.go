package core

type UserConfig struct {
	Args []string `json:"args,omitempty"`
}

type Container struct {
	ID    string   `json:"id"`
	Image *Image   `json:"image,omitempty"`
	Args  []string `json:"args,omitempty"`
}
