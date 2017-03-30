package handler_test

import (
	"testing"
	"github.com/limetext/log4go"
	"golang.org/x/net/context"
	// TODO SEEDMS replace all references to github.com/tomogoma/seedms
	// with new path
	"github.com/tomogoma/seedms/handler"
	"github.com/tomogoma/seedms/proto"
	"net/http"
	"reflect"
	"github.com/dgrijalva/jwt-go"
	"github.com/tomogoma/go-commons/errors"
)

type TokenValidatorMock struct {
	ExpToken *jwt.Token
	ExpErr   error
	errors.AuthErrCheck
}

func (t *TokenValidatorMock) Validate(token string, claims jwt.Claims) (*jwt.Token, error) {
	return t.ExpToken, t.ExpErr
}

var srvID = "test_server"
var logger = log4go.Logger{}

func TestNew(t *testing.T) {
	s, err := handler.NewSeed(srvID, &TokenValidatorMock{}, logger)
	if err != nil {
		t.Fatalf("server.NewSeed(): %v", err)
	}
	if s == nil {
		t.Fatal("Got a nil Seed")
	}
}

func TestNew_emptyID(t *testing.T) {
	_, err := handler.NewSeed("", &TokenValidatorMock{}, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_nilTokenValidator(t *testing.T) {
	_, err := handler.NewSeed(srvID, nil, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_nilLogger(t *testing.T) {
	_, err := handler.NewSeed(srvID, &TokenValidatorMock{}, nil)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestServer_Hello(t *testing.T) {
	type HelloTC struct {
		Desc    string
		TknVal  *TokenValidatorMock
		Req     *proto.HelloRequest
		ExpResp *proto.HelloResponse
	}
	tcs := []HelloTC{
		{
			Desc: "Greeting success",
			TknVal: &TokenValidatorMock{
				ExpErr: nil,
			},
			Req: &proto.HelloRequest{
				Token: "some.valid.token",
				Name:  "Test Bot",
			},
			ExpResp: &proto.HelloResponse{
				Code:     http.StatusOK,
				Greeting: "Hello Test Bot",
				Id:       srvID,
			},
		},
		{
			Desc: "Invalid token reported",
			TknVal: &TokenValidatorMock{
				ExpErr: errors.NewAuth("Bad token!"),
			},
			Req: &proto.HelloRequest{
				Token: "some.invalid.token",
				Name:  "Test Bot",
			},
			ExpResp: &proto.HelloResponse{
				Code:   http.StatusUnauthorized,
				Id:     srvID,
				Detail: "Bad token!",
			},
		},
		{
			Desc: "Token validation error",
			TknVal: &TokenValidatorMock{
				ExpErr: errors.New("Internal error"),
			},
			Req: &proto.HelloRequest{
				Token: "some.valid.token",
				Name:  "Test Bot",
			},
			ExpResp: &proto.HelloResponse{
				Code:   http.StatusInternalServerError,
				Id:     srvID,
				Detail: handler.SomethingWickedError,
			},
		},
	}
	for _, tc := range tcs {
		s, err := handler.NewSeed(srvID, tc.TknVal, logger)
		if err != nil {
			t.Fatalf("server.NewSeed(): %v", err)
		}
		resp := new(proto.HelloResponse)
		err = s.Hello(context.TODO(), tc.Req, resp)
		if err != nil {
			t.Fatalf("%s - server.Hello(): %v", tc.Desc, err)
		}
		if !reflect.DeepEqual(tc.ExpResp, resp) {
			t.Errorf("%s - Unexpected response:\nExpect:\t%+v\nGot:\t%+v",
				tc.Desc, tc.ExpResp, resp)
		}
	}
}
