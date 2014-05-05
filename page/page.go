package page

import "github.com/bitantics/amerigo/resource"

type Page struct {
	Path          string
	Links, Assets []string
}

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
