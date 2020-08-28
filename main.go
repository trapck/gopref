package main

import (
	"log"

	"github.com/trapck/gopref/cfg"
	"github.com/trapck/gopref/mongostore"
	"github.com/trapck/gopref/server"
)

func main() {
	s := &mongostore.Store{}
	if err := s.Init(); err != nil {
		log.Fatalf("could not open mongo db connection %q", err)
	}
	defer s.Close()
	if err := server.NewApp(s).Start(cfg.Port); err != nil {
		log.Fatalf("could not listen on port %d %v", cfg.Port, err)
	}
}
