package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/N4darae/anti-mage/server"
)

//go:embed all:web
var embedded embed.FS

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "loopback address to serve on")
	webDir := flag.String("web", "", "serve the page from this directory instead of the copy compiled in")
	flag.Parse()

	os.Unsetenv("ZONEINFO")

	logger := log.New(os.Stderr, "", log.LstdFlags)

	var pages fs.FS
	if *webDir != "" {
		pages = os.DirFS(*webDir)
		logger.Printf("serving the page from %s", *webDir)
	} else {
		sub, err := fs.Sub(embedded, "web")
		if err != nil {
			logger.Fatalf("the page was not compiled in: %v", err)
		}
		pages = sub
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := server.New(server.Options{Addr: *addr, Web: pages, Log: logger})
	if err := s.ListenAndServe(ctx); err != nil {
		logger.Fatalf("%v", err)
	}
}
