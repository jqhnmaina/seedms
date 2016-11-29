package server

import (
	"errors"
	"golang.org/x/net/context"
	"github.com/tomogoma/seedms/server/proto"
	"net/http"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/go-commons/server/helper"
)

type Logger interface {
	Info(interface{}, ...interface{})
	Warn(interface{}, ...interface{}) error
	Error(interface{}, ...interface{}) error
}

type TokenValidator interface {
	Validate(token string) (*token.Token, error)
}

type Model interface {
}

type Server struct {
	token TokenValidator
	model Model
	log   Logger
	tIDCh chan int
	id    string
}

const (
	SomethingWickedError = "Something wicked happened"
	InvalidTokenError = "Token was invalid"
)

var ErrorNilTokenValidator = errors.New("TokenValidator was nil")
var ErrorNilRiderModel = errors.New("Model was nil");
var ErrorNilLogger = errors.New("Logger was nil");

func New(id string, m Model, tv TokenValidator, lg Logger) (*Server, error) {
	if m == nil {
		return nil, ErrorNilRiderModel
	}
	if tv == nil {
		return nil, ErrorNilTokenValidator
	}
	if lg == nil {
		return nil, ErrorNilLogger
	}
	tIDCh := make(chan int)
	go helper.TransactionSerializer(tIDCh)
	return &Server{id: id, model:m, token: tv, log:lg, tIDCh: tIDCh}, nil
}

func (s *Server) Hello(c context.Context, req *seed.HelloRequest, resp *seed.HelloResponse) error {
	tID := <-s.tIDCh
	s.log.Info("%d - Hello request", tID)
	_, err := s.token.Validate(req.Token);
	if err != nil {
		s.packageTokenError(resp)
		s.log.Info("%d - Token was invalid: %s", tID, err)
		return nil
	}
	greeting := "Hello " + req.Name
	s.packageGreeting(http.StatusOK, greeting, resp)
	return nil
}

func (s *Server) packageTokenError(resp *seed.HelloResponse) {
	s.packageError(http.StatusUnauthorized, InvalidTokenError, resp)
}

func (s *Server) packageInternalError(resp *seed.HelloResponse) {
	s.packageError(http.StatusInternalServerError, SomethingWickedError, resp)
}

func (s *Server) packageError(code int32, errStr string, resp *seed.HelloResponse) {
	resp.Id = s.id
	resp.Code = code
	resp.Detail = errStr
}

func (s *Server) packageGreeting(code int32, greeting string, resp *seed.HelloResponse) {
	resp.Id = s.id
	resp.Code = code
	resp.Greeting = greeting
}
