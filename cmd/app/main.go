package main

import (
	"io/fs"

	"github.com/garnizeh/go-web-boilerplate/embeded"
)

func main() {
	if err := fs.WalkDir(
		embeded.Static(),
		".",
		func(path string, d fs.DirEntry, err error) error {
			println(d.Name())
			return nil
		},
	); err != nil {
		println(err)
	}
}
