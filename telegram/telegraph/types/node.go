package types

import (
	"encoding/json"
)

type Node struct {
	Tag      string
	Attrs    map[string]string
	Children []Node
}

func (entity Node) MarshalJSON() ([]byte, error) {
	if entity.Tag == "text" {
		return json.Marshal(entity.Attrs["value"])
	} else {
		return json.Marshal(struct {
			Tag      string            `json:"tag"`
			Attrs    map[string]string `json:"attrs,omitempty"`
			Children []Node            `json:"children,omitempty"`
		}{
			Tag:      entity.Tag,
			Attrs:    entity.Attrs,
			Children: entity.Children,
		})
	}
}
