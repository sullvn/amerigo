/*
Package crawler implements an internal website crawler. In other words, it
limits itself to a single domain.

The crawler will emit page.Pages, which include URIs to themselves and their
connected resources. Resources may be assets or links to other pages.

It doesn't store anything beyond the set of visited pages.
*/
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

// Crawler contains the site URL it is running on and the set of visited pages.
// To obtain results, the caller should `select` over Crawler.Pages,
// Crawler.Errors, and Crawler.Done. When Crawler.Done is emitted, it has
// finished all work and closed all channels.
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

// New will create a new Crawler which works on a given domain. Domains may
// be given with the scheme, http://www.website.com, or not, www.website.com.
func New(siteURL string) (*Crawler, error) {
	site, err := url.Parse(siteURL)
	if err != nil {
		return nil, err
	}

	// Normalize URL to have scheme. Otherwise url.URL will consider schemeless
	// domains to be relative URLs with no domain.
	if site.Scheme == "" {
		site, err = url.Parse("http://" + siteURL)
		if err != nil {
			return nil, err
		}
	}

	return &Crawler{Site: site}, nil
}

// Start begins the web crawling process. Initializes datastructures and starts
// download workers.
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

	// Intermittently close idle TCP connections. A response to Go not
	// closing any sockets, causing file descriptor exhaustion errors.
	// TODO (bitantics): investigate source of this to see if there is a way
	// to not do this manually.
	go (func() {
		reqCount := 0
		for _ = range c.countRequests {
			if reqCount++; reqCount%workers == 0 {
				http.DefaultTransport.(*http.Transport).CloseIdleConnections()
			}
		}
	})()

	// When all download workers are idle, we know we are done.
	go (func() {
		c.activeWorkers.Wait()
		close(c.pendingRequests)
		close(c.countRequests)
		close(c.Pages)
		close(c.Errors)
		close(c.Done)
	})()
}

// requestWorker collects paths, visits them, and returns the results.
func (c *Crawler) requestWorker() {
	for reqPath := range c.pendingRequests {
		pg, err := c.visitPage(reqPath)

		if err != nil {
			c.Errors <- err
		} else if pg != nil {
			c.Pages <- pg
		}

		c.activeWorkers.Done()
	}
}

// markVisited marks an internal URL as been visited by the crawler.
func (c *Crawler) markVisited(path string) {
	c.visitedLock.Lock()
	defer c.visitedLock.Unlock()

	c.visited[path] = true
}

// hasVisited checks if a URL has been visited.
func (c *Crawler) hasVisited(path string) bool {
	c.visitedLock.RLock()
	defer c.visitedLock.RUnlock()

	_, visited := c.visited[path]
	return visited
}

// scheduleVisit queues a URL to be visited by the download workers.
func (c *Crawler) scheduleVisit(path string) {
	if c.hasVisited(path) {
		return
	}

	c.activeWorkers.Add(1)
	// Wait for an available worker in the background.
	// NOTE (bitantics): This not only sort of abuses goroutines as a queue,
	// it won't prevent duplicates.
	go (func() {
		c.pendingRequests <- path
	})()
}

// visitPage takes a URL, then returns a page.Page containing all the links and
// assets at that URL. It will also schedule visits to any URLs it comes
// across.
func (c *Crawler) visitPage(relPath string) (*page.Page, error) {
	// Turn relative URL to absolute.
	pageURL, err := c.Site.Parse(relPath)
	if err != nil {
		return nil, err
	}

	// Avoid HTTP call if we have already been here.
	if c.hasVisited(pageURL.Path) {
		return nil, nil
	}

	pageResp, err := http.Get(pageURL.String())
	if err != nil {
		return nil, err
	}
	defer pageResp.Body.Close()

	// Visited status may have changed since HTTP call was made. Check again.
	if c.hasVisited(pageURL.Path) {
		return nil, nil
	}
	c.markVisited(pageURL.Path)

	rs := resource.NewSet()

	// Walk through HTML page, looking for link and asset URIs.
	tok := html.NewTokenizer(pageResp.Body)
	for tok.Err() != io.EOF {
		tokType := tok.Next()
		switch tokType {

		// URIs are only in some tags.
		case html.StartTagToken, html.SelfClosingTagToken:
			res := resource.FromTagTokenizer(tok)
			if res == nil {
				continue
			}

			rs.Add(res, c.Site)
			// Schedule visit if URI is an internal link
			if res.Type == resource.Link && res.IsInternal(c.Site) {
				c.scheduleVisit(res.URL.Path)
			}

		case html.ErrorToken:
			if err := tok.Err(); err != io.EOF && err != nil {
				return nil, err
			}
		}
	}

	// Generate lists of unique links and assets
	return page.NewFromResourceSet(pageURL.Path, rs), nil
}
