package models

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

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (u User) Summarize() string {
	return fmt.Sprintf("用户：%s", u.Name)
}
