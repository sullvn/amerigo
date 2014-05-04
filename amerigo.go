package main

import (
	"net/http"

	"github.com/bitantics/amerigo/page"
)

func main() {
	p, err := http.Get("http://bitantics.com/")
	defer p.Body.Close()
	if err != nil {
		panic(err)
	}

	page.FromText(p.Body)
}
