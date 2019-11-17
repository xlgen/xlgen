package main

import (
	"log"
	"os"
)

func main() {
	var s sitespec
	var err error
	s.dir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if err = s.loadFromDir(); err != nil {
		log.Fatal(err)
	}
	if err = s.execute(); err != nil {
		log.Fatal(err)
	}
}
