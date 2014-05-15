package resource

import "net/url"

// ResourceSet is a bag of unique Resources
type ResourceSet struct {
	Links, Assets map[string]bool
}

// NewSet creates and initializes a new ResourceSet
func NewSet() *ResourceSet {
	return &ResourceSet{
		Links:  make(map[string]bool),
		Assets: make(map[string]bool),
	}
}

// Add takes a resource and site URL, then adds the normalized resource to the
// ResourceSet. A normalized resource has a relative path if it is internal.
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
