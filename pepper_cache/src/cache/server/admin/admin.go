package admin

import (
	"fmt"
	"net/http"

	"cache/server/core"
	"cache/server/setting"
)

type Server struct {
	*core.Server
	AdminPort int
}

func (s *Server) HttpListen() {
	http.Handle("/cluster", s.clusterHandler())
	http.Handle("/stat", s.statusHandler())
	http.ListenAndServe(fmt.Sprintf(":%d", s.AdminPort), nil)
}

func NewAdminServer(s *core.Server) *Server {
	return &Server{s, setting.AdminPort}
}
