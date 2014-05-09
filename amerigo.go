/*
Amerigo is an individual website crawler. It is named after Amerigo Vespucci,
the 15th century Italian explorer and cartographer.

It crawls a website to output the pages and their relationships.
It is designed to take a the website graph and output the serialization as
it crawls. This output is then fed to tools which may use the data as they
please.
*/
package main

import (
	"flag"

	"github.com/bitantics/amerigo/crawler"
)

// Currently just prints out the relative URLs to stdout.
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

	c.Start(32)

crawl:
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
		case <-c.Done:
			break crawl
		}
	}
}
