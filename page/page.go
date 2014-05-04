package page

import (
	"io"

	"code.google.com/p/go.net/html"
	"github.com/bitantics/amerigo/resource"
)

type Page struct {
	Links, Assets []*resource.Resource
}

func FromText(pageText io.Reader) (*Page, error) {
	page := new(Page)
	tok := html.NewTokenizer(pageText)

	for tok.Err() == nil {
		tokType := tok.Next()
		switch tokType {
		case html.StartTagToken, html.SelfClosingTagToken:
			res := resource.FromTagTokenizer(tok)
			if res != nil {
				println(res.URL.String())
			}

		case html.ErrorToken:
			return nil, tok.Err()
		}
	}

	return page, nil
}
