package model

// Note is the shared HTTP response shape. Versions reject stale edits.
type Note struct {
	ID      int    `json:"id"`
	Text    string `json:"text"`
	Version int    `json:"version"`
}
