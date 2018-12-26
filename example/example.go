package main

import (
	"github.com/jakobvarmose/packer"
	"log"
	"net/http"
)

func main() {
	root, err := packer.Load()
	if err != nil {
		log.Println(err)
		return
	}

	err = http.ListenAndServe(":8080", http.FileServer(root))
	if err != nil {
		log.Println(err)
		return
	}
}
