package bookinfo

import (
	"fmt"
	"io"
	"strings"

	"github.com/antchfx/xmlquery"
)

func ParseFB2(input io.Reader) (*Book, error) {
	doc, errParse := xmlquery.Parse(input)
	if errParse != nil {
		return nil, fmt.Errorf("xml parse: %w", errParse)
	}

	book := &Book{
		Title:    xmlquery.FindOne(doc, "//title").InnerText(),
		Language: xmlquery.FindOne(doc, "//lang").InnerText(),
	}

	for _, author := range xmlquery.Find(doc, "//author") {
		firstname := xmlquery.FindOne(author, "//first-name")
		lastname := xmlquery.FindOne(author, "//last-name")
		name := concatXMLTexts(" ", firstname, lastname)
		book.Authors = append(book.Authors, name)
	}

	if annotation := xmlquery.FindOne(doc, "//annotation"); annotation != nil {
		book.Annotation = annotation.InnerText()
	}

	for _, genre := range xmlquery.Find(doc, "//genre") {
		book.Genres = append(book.Genres, genre.InnerText())
	}

	return book, nil
}

func concatXMLTexts(sep string, nodes ...*xmlquery.Node) string {
	var texts []string
	for _, node := range nodes {
		if node != nil {
			texts = append(texts, node.InnerText())
		}
	}
	return strings.Join(texts, sep)
}
