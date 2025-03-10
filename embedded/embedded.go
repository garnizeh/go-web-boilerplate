package embedded

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed static/*
var static embed.FS

//go:embed mails/*
var mails embed.FS

// Static returns a file system with the static files used by
// the application, mainly css and js files. To exclude a file
// to be loaded into the embed, rename it to start with an
// underscore and the go embed tool will skip it.
func Static() fs.FS {
	staticFS, err := fs.Sub(static, "static")
	if err != nil {
		panic(fmt.Sprintf("failed to create static file system: %v", err))
	}

	return staticFS
}

// Mails returns a file system with the mail template files used by
// the application. To exclude a file to be loaded into the embed,
// rename it to start with an underscore and the go embed tool will
// skip it.
func Mails() fs.FS {
	smailsFS, err := fs.Sub(mails, "mails")
	if err != nil {
		panic(fmt.Sprintf("failed to create mails file system: %v", err))
	}

	return smailsFS
}
