package sitemap

import (
	"net/http"
	"net/url"
	"sync"

	"code.google.com/p/go.net/html"
	"github.com/bitantics/amerigo/resource"
)

type page struct {
	Links, Assets map[string]bool
}

type SiteMap struct {
	Pages map[string]page
	Site  *url.URL

	pagesLock sync.RWMutex
	workers   sync.WaitGroup
}

func New(siteURL string) (*SiteMap, error) {
	site, err := url.Parse(siteURL)
	if err != nil {
		return nil, err
	}

	if site.Scheme == "" {
		site, err = url.Parse("http://" + siteURL)
		if err != nil {
			return nil, err
		}
	}

	return &SiteMap{
		Pages: make(map[string]page),
		Site:  site,
	}, nil
}

func (sm *SiteMap) Create() error {
	defer sm.workers.Wait()
	return sm.addPage("")
}

func (sm *SiteMap) addPage(relPath string) error {
	defer sm.workers.Done()

	pageURL, err := sm.Site.Parse(relPath)
	if err != nil {
		return err
	}

	sm.pagesLock.RLock()
	_, isMapped := sm.Pages[pageURL.Path]
	sm.pagesLock.RUnlock()
	if isMapped {
		return nil
	}

	pageResp, err := http.Get(pageURL.String())
	if err != nil {
		return err
	}
	defer pageResp.Body.Close()

	sm.pagesLock.Lock()
	sm.Pages[pageURL.Path] = page{
		Links:  make(map[string]bool),
		Assets: make(map[string]bool),
	}
	sm.pagesLock.Unlock()

	tok := html.NewTokenizer(pageResp.Body)
	for tok.Err() == nil {
		tokType := tok.Next()
		switch tokType {

		case html.StartTagToken, html.SelfClosingTagToken:
			res := resource.FromTagTokenizer(tok)
			if res == nil {
				continue
			}

			sm.addResource(pageURL.Path, res)
			if res.Type == resource.Link && res.IsInternal(sm.Site) {
				sm.workers.Add(1)
				go sm.addPage(res.URL.Path)
			}

		case html.ErrorToken:
			return tok.Err()
		}
	}

	return nil
}

func (sm *SiteMap) addResource(pagePath string, res *resource.Resource) {
	if res == nil {
		return
	}

	sm.pagesLock.RLock()
	page := sm.Pages[pagePath]
	sm.pagesLock.RUnlock()

	var resURL string
	if res.IsInternal(sm.Site) {
		resURL = res.URL.Path
	} else {
		resURL = res.URL.String()
	}

	switch res.Type {
	case resource.Link:
		page.Links[resURL] = true
	case resource.Asset:
		page.Assets[resURL] = true
	}
}
