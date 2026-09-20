package model

import "fmt"

type Summarizer interface {
	Summarize() string
}

type Item struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

func (i Item) Summarize() string {
	return fmt.Sprintf("失物：%s", i.Name)
}
