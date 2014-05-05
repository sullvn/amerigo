package crawler

import (
	"io"
	"net/http"
	"net/url"
	"sync"

	"code.google.com/p/go.net/html"
	"github.com/bitantics/amerigo/page"
	"github.com/bitantics/amerigo/resource"
)

type Crawler struct {
	Site   *url.URL
	Pages  chan *page.Page
	Errors chan error

	visited     map[string]bool
	visitedLock sync.RWMutex
	workers     sync.WaitGroup
}

func New(siteURL string) (*Crawler, error) {
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

	return &Crawler{Site: site}, nil
}

func (c *Crawler) Start() chan bool {
	c.visited = make(map[string]bool)

	c.Pages = make(chan *page.Page)
	c.Errors = make(chan error)
	done := make(chan bool)

	go (func() {
		c.workers.Add(1)
		c.visitPage("")

		c.workers.Wait()
		close(c.Pages)
		close(c.Errors)
		close(done)
	})()

	return done
}

func (c *Crawler) markVisited(path string) {
	c.visitedLock.Lock()
	defer c.visitedLock.Unlock()

	c.visited[path] = true
}

func (c *Crawler) hasVisited(path string) bool {
	c.visitedLock.RLock()
	defer c.visitedLock.RUnlock()

	_, visited := c.visited[path]
	return visited
}

func (c *Crawler) visitPage(relPath string) {
	defer c.workers.Done()

	pageURL, err := c.Site.Parse(relPath)
	if err != nil {
		c.Errors <- err
		return
	}

	if c.hasVisited(pageURL.Path) {
		return
	}

	pageResp, err := http.Get(pageURL.String())
	if err != nil {
		c.Errors <- err
		return
	}
	defer pageResp.Body.Close()

	if c.hasVisited(pageURL.Path) {
		return
	}
	c.markVisited(pageURL.Path)

	rs := resource.NewSet()

	tok := html.NewTokenizer(pageResp.Body)
	for tok.Err() != io.EOF {
		tokType := tok.Next()
		switch tokType {

		case html.StartTagToken, html.SelfClosingTagToken:
			res := resource.FromTagTokenizer(tok)
			if res == nil {
				continue
			}

			rs.Add(res, c.Site)
			if res.Type == resource.Link && res.IsInternal(c.Site) {
				c.workers.Add(1)
				go c.visitPage(res.URL.Path)
			}

		case html.ErrorToken:
			if err := tok.Err(); err != io.EOF && err != nil {
				c.Errors <- err
				return
			}
		}
	}

	c.Pages <- page.NewFromResourceSet(pageURL.Path, rs)
}
