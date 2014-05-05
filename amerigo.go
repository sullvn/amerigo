package main

import (
	"flag"

	"github.com/bitantics/amerigo/sitemap"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		return
	}
	site := flag.Arg(0)

	sm, err := sitemap.New(site)
	if err != nil {
		panic(err)
	}

	sm.Create()
	for key, _ := range sm.Pages {
		println(key)
	}
}
