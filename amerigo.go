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
	"encoding/json"
	"flag"
	"fmt"

	"github.com/bitantics/amerigo/crawler"
	"github.com/bitantics/amerigo/page"
)

// Currently just prints out the relative URLs to stdout
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

	fmt.Println("[")
	defer fmt.Println("\n]")

	firstPage := true

crawl:
	for {
		select {
		case pg := <-c.Pages:
			if pg != nil {
				if firstPage {
					firstPage = false
				} else {
					fmt.Println(",")
				}

				printPage(pg)
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

func printPage(p *page.Page) error {
	pg, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}
	fmt.Print(string(pg))
	return nil
}
