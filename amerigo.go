package main

import (
	"flag"

	"github.com/bitantics/amerigo/crawler"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		return
	}
	site := flag.Arg(0)

	c, err := crawler.New(site)
	if err != nil {
		panic(err)
	}

	done := c.Start()

Crawl:
	for {
		select {
		case page := <-c.Pages:
			if page != nil {
				println(page.Path)
			}
		case err = <-c.Errors:
			if err != nil {
				panic(err)
			}
		case <-done:
			break Crawl
		}
	}
}
