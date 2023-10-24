package bookinfo

import "time"

type Book struct {
	Title      string
	WrittenAt  time.Time
	Authors    []string
	Genres     []string
	Language   string
	Annotation string
}
