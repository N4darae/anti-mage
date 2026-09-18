package server

import (
	"context"
	"io"
	"log"
	"testing"
	"testing/fstest"
)

func newTestServer(t *testing.T, files fstest.MapFS) *Server {
	t.Helper()
	return New(Options{Addr: "127.0.0.1:0", Web: files, Log: log.New(io.Discard, "", 0)})
}

func TestNonLoopbackAddressIsRefused(t *testing.T) {

	for _, addr := range []string{"0.0.0.0:8787", "198.51.100.10:8787", "[::]:8787", ":8787", ":0", "[::1%25lo]:0"} {
		s := New(Options{Addr: addr, Log: log.New(io.Discard, "", 0)})
		if err := s.ListenAndServe(context.Background()); err == nil {
			t.Errorf("address %q was accepted; this server examines the browser on this machine only", addr)
		}
	}
}
