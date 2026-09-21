package dto

import (
	"encoding/json"
	"io"
)

type UploadMeta struct {
	Name   string   `json:"name" binding:"required"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Mime   string   `json:"mime"`
	Grant  []string `json:"grant"`
}

type UploadInput struct {
	OwnerLogin string
	Name       string
	Mime       string
	Public     bool
	Grant      []string
	JSONData   json.RawMessage
	File       io.Reader
}

type ListInput struct {
	RequesterLogin string
	OwnerLogin     string // "" — "показать свои"
	Key            string
	Value          string
	Limit          int
}
