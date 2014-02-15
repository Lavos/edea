package edea

import (
	"github.com/hoisie/web"
	"log"
)

type Server struct {
	server *web.Server
	curator *Curator
}

func NewServer(c *Curator) *Server {
	w := web.NewServer()
	s := &Server{
		server: w,
		curator: c,
	}

	w.Get("/next", s.getTweet)

	return s
}

func (s *Server) Run () {
	log.Print("Edea API Server started.")
	go s.server.Run(":8001")
}

func (s *Server) getTweet (ctx *web.Context) []byte {
	tweet_bytes := s.curator.GetNext()

	return tweet_bytes
}
