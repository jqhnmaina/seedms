package handler

import (
	"errors"
	"golang.org/x/net/context"
	// TODO SEEDMS replace all references to github.com/tomogoma/seedms
	// with new path
	"github.com/tomogoma/seedms/handler/proto"
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
	IsClientError(error) bool
}

type Seed struct {
	token TokenValidator
	log   Logger
	tIDCh chan int
	id    string
}

const (
	SomethingWickedError = "Something wicked happened"
)

var ErrorNilTokenValidator = errors.New("TokenValidator was nil")
var ErrorNilLogger = errors.New("Logger was nil")
var ErrorEmptyID = errors.New("ID was empty")

func NewSeed(ID string, tv TokenValidator, lg Logger) (*Seed, error) {
	if tv == nil {
		return nil, ErrorNilTokenValidator
	}
	if lg == nil {
		return nil, ErrorNilLogger
	}
	if ID == "" {
		return nil, ErrorEmptyID
	}
	tIDCh := make(chan int)
	go helper.TransactionSerializer(tIDCh)
	return &Seed{id: ID, token: tv, log: lg, tIDCh: tIDCh}, nil
}

func (s *Seed) Hello(c context.Context, req *proto.HelloRequest, resp *proto.HelloResponse) error {
	resp.Id = s.id
	tID := <-s.tIDCh
	s.log.Info("%d - Hello request", tID)
	_, err := s.token.Validate(req.Token);
	if err != nil {
		if s.token.IsClientError(err) {
			resp.Code = http.StatusUnauthorized
			resp.Detail = err.Error()
			return nil
		}
		resp.Code = http.StatusInternalServerError
		resp.Detail = SomethingWickedError
		s.log.Error("%d - Failed to validate user token: %s", tID, err)
		return nil
	}
	resp.Code = http.StatusOK
	resp.Greeting = "Hello " + req.Name
	return nil
}
