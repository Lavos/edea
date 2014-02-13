package edea

import (
	"github.com/hoisie/web"
	"github.com/cznic/kv"
	"log"
)

type Server struct {
	server *web.Server
	tweetDB *kv.DB
	flagDB *kv.DB
}

func NewServer(tweetDB *kv.DB, flagDB *kv.DB) *Server {
	w := web.NewServer()
	s := &Server{
		server: w,
		tweetDB: tweetDB,
		flagDB: flagDB,
	}

	w.Get("/next", s.getTweet)

	return s
}

func (s *Server) Run () {
	log.Print("Edea API Server started.")
	go s.server.Run(":8001")
}

func (s *Server) getTweet (ctx *web.Context) []byte {
	key, value, _ := s.tweetDB.First()
	s.tweetDB.Delete(key)

	return value
}
