package web

import (
	"encoding/json"
	"io"
)

type captureViews struct{}

func (captureViews) Load() error { return nil }
func (captureViews) Render(w io.Writer, name string, data any, _ ...string) error {
	return json.NewEncoder(w).Encode(struct {
		Name string
		Data any
	}{name, data})
}
