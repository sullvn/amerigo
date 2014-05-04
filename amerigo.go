package main

import "github.com/bitantics/amerigo/sitemap"

func main() {
	sm, err := sitemap.New("http://bitantics.com")
	if err != nil {
		panic(err)
	}

	sm.Create()
	for key, _ := range sm.Pages {
		println(key)
	}
}
