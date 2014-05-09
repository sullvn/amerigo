// Package page deals with representing HTML pages and their connections
package page

import "github.com/bitantics/amerigo/resource"

// Page represents a HTML page connections by containing
// * Relative path to itself
// * Paths for any links and assets. Internal paths should be relative, while
// external paths are absolute.
type Page struct {
	Path          string
	Links, Assets []string
}

// NewFromResourceSet generates a Page from a resource.ResourceSet. It
// is useful for generating Pages with only unique paths.
func NewFromResourceSet(path string, rs *resource.ResourceSet) *Page {
	page := &Page{
		Path:   path,
		Links:  make([]string, 0, len(rs.Links)),
		Assets: make([]string, 0, len(rs.Assets)),
	}

	for link, _ := range rs.Links {
		page.Links = append(page.Links, link)
	}
	for asset, _ := range rs.Assets {
		page.Assets = append(page.Assets, asset)
	}

	return page
}
