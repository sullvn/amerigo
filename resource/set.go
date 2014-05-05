package resource

import "net/url"

type ResourceSet struct {
	Links, Assets map[string]bool
}

func NewSet() *ResourceSet {
	return &ResourceSet{
		Links:  make(map[string]bool),
		Assets: make(map[string]bool),
	}
}

func (rs *ResourceSet) Add(res *Resource, site *url.URL) {
	if res == nil {
		return
	}
	var resURL string
	if res.IsInternal(site) {
		resURL = res.URL.Path
	} else {
		resURL = res.URL.String()
	}

	switch res.Type {
	case Link:
		rs.Links[resURL] = true
	case Asset:
		rs.Assets[resURL] = true
	}
}
