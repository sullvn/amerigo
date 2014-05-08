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
	Done   chan bool

	visited         map[string]bool
	visitedLock     sync.RWMutex
	activeWorkers   sync.WaitGroup
	pendingRequests chan string
	countRequests   chan bool
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

func (c *Crawler) Start(workers int) {
	c.visited = make(map[string]bool)

	c.Pages = make(chan *page.Page)
	c.Errors = make(chan error)
	c.Done = make(chan bool)

	c.pendingRequests = make(chan string)
	c.countRequests = make(chan bool)

	for i := workers; i > 0; i-- {
		go c.requestWorker()
	}

	c.scheduleVisit("")

	go (func() {
		reqCount := 0
		for _ = range c.countRequests {
			if reqCount++; reqCount%workers == 0 {
				http.DefaultTransport.(*http.Transport).CloseIdleConnections()
			}
		}
	})()

	go (func() {
		c.activeWorkers.Wait()
		close(c.pendingRequests)
		close(c.countRequests)
		close(c.Pages)
		close(c.Errors)
		close(c.Done)
	})()
}

func (c *Crawler) requestWorker() {
	requests := 0
	for reqPath := range c.pendingRequests {
		pg, err := c.visitPage(reqPath)

		if err != nil {
			c.Errors <- err
		} else if pg != nil {
			c.Pages <- pg
		}

		c.activeWorkers.Done()
		requests++
	}
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

func (c *Crawler) scheduleVisit(path string) {
	if c.hasVisited(path) {
		return
	}

	c.activeWorkers.Add(1)
	go (func() {
		c.pendingRequests <- path
	})()
}

func (c *Crawler) visitPage(relPath string) (*page.Page, error) {
	pageURL, err := c.Site.Parse(relPath)
	if err != nil {
		return nil, err
	}

	if c.hasVisited(pageURL.Path) {
		return nil, nil
	}

	pageResp, err := http.Get(pageURL.String())
	if err != nil {
		return nil, err
	}
	defer pageResp.Body.Close()

	if c.hasVisited(pageURL.Path) {
		return nil, nil
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
				c.scheduleVisit(res.URL.Path)
			}

		case html.ErrorToken:
			if err := tok.Err(); err != io.EOF && err != nil {
				return nil, err
			}
		}
	}

	return page.NewFromResourceSet(pageURL.Path, rs), nil
}
