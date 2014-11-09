package core

type Image struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id,omitempty"`
}
