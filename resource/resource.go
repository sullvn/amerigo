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
var attrTypes = map[string]Type{
	"code":     Asset,
	"data":     Asset,
	"download": Asset,
	"href":     Link,
	"manifest": Asset,
	"poster":   Asset,
	"src":      Asset,
}

type Resource struct {
	Type Type
	URL  *url.URL
}

func attrRes(z *html.Tokenizer, attr string) (*Resource, error) {
	defer println("")
	moreAttr := true
	for moreAttr {
		key, val, ma := z.TagAttr()
		moreAttr = ma

		print(" ", string(key))
		if attr != string(key) {
			continue
		}

		resURL, err := url.Parse(string(val))
		if err != nil {
			return nil, err
		}

		res := &Resource{
			Type: attrTypes[attr],
			URL:  resURL,
		}
		return res, nil
	}
	return nil, nil
}

func FromTagTokenizer(z *html.Tokenizer) *Resource {
	tag, hasAttr := z.TagName()
	println(string(tag))

	attr, relevantTag := tagAttrs[string(tag)]
	if !relevantTag || !hasAttr {
		return nil
	}

	if res, err := attrRes(z, attr); err == nil {
		return res
	}
	return nil
}
