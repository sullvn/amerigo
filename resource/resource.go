package resource

import (
	"net/url"

	"code.google.com/p/go.net/html"
)

type Type int

const (
	Link Type = iota
	Asset
)

var tagAttrs = map[string]string{
	"a":      "href",
	"applet": "code",
	"area":   "href",
	"audio":  "src",
	"base":   "href",
	"embed":  "src",
	"html":   "manifest",
	"link":   "href",
	"iframe": "src",
	"img":    "src",
	"input":  "src",
	"object": "data",
	"script": "src",
	"source": "src",
	"track":  "src",
	"video":  "src",
}

type Resource struct {
	Type Type
	URL  *url.URL
}

func (r *Resource) IsInternal(site *url.URL) bool {
	if !r.URL.IsAbs() {
		return true
	}

	return r.URL.Host == site.Host
}

func attrRes(z *html.Tokenizer, attr string) (*Resource, error) {
	moreAttr := true
	for moreAttr {
		key, val, ma := z.TagAttr()
		moreAttr = ma

		if attr != string(key) {
			continue
		}

		resURL, err := url.Parse(string(val))
		if err != nil {
			return nil, err
		}

		res := &Resource{
			Type: Asset,
			URL:  resURL,
		}
		return res, nil
	}
	return nil, nil
}

func FromTagTokenizer(z *html.Tokenizer) *Resource {
	tag, hasAttr := z.TagName()

	attr, relevantTag := tagAttrs[string(tag)]
	if !relevantTag || !hasAttr {
		return nil
	}

	if res, err := attrRes(z, attr); res != nil && err == nil {
		if string(tag) == "a" {
			res.Type = Link
		}
		return res
	}
	return nil
}
