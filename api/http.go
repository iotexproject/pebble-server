package api

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/iotexproject/pebble-server/db"
)

type errResp struct {
	Error string `json:"error,omitempty"`
}

func newErrResp(err error) *errResp {
	return &errResp{Error: err.Error()}
}

type httpServer struct {
	engine *gin.Engine
	db     *db.DB
}

func (s *httpServer) query(c *gin.Context) {

}

func (s *httpServer) receive(c *gin.Context) {

}

func Run(db *db.DB, address string) error {
	s := &httpServer{
		engine: gin.Default(),
		db:     db,
	}

	s.engine.GET("/device", s.query)
	s.engine.POST("/device", s.receive)

	err := s.engine.Run(address)
	return errors.Wrap(err, "failed to start http server")
}
