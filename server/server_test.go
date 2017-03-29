package server_test

import (
	"testing"
	"github.com/tomogoma/go-commons/auth/token"
	"github.com/tomogoma/seedms/server"
	"github.com/limetext/log4go"
	"golang.org/x/net/context"
	// TODO SEEDMS replace all references to github.com/tomogoma/seedms
	// with new path
	"github.com/tomogoma/seedms/server/proto"
	"net/http"
	"errors"
	"reflect"
)

type TokenValidatorMock struct {
	ExpToken *token.Token
	ExpErr   error
	ExpClErr bool
}

func (t *TokenValidatorMock) Validate(token string) (*token.Token, error) {
	return t.ExpToken, t.ExpErr
}

func (t *TokenValidatorMock) IsClientError(error) bool {
	return t.ExpClErr
}

var srvID = "test_server"
var logger = log4go.Logger{}

func TestNew(t *testing.T) {
	s, err := server.New(srvID, &TokenValidatorMock{}, logger)
	if err != nil {
		t.Fatalf("server.New(): %v", err)
	}
	if s == nil {
		t.Fatal("Got a nil Server")
	}
}

func TestNew_emptyID(t *testing.T) {
	_, err := server.New("", &TokenValidatorMock{}, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_nilTokenValidator(t *testing.T) {
	_, err := server.New(srvID, nil, logger)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestNew_nilLogger(t *testing.T) {
	_, err := server.New(srvID, &TokenValidatorMock{}, nil)
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
				ExpClErr: false,
			},
			Req: &proto.HelloRequest{
				Token: "some.valid.token",
				Name: "Test Bot",
			},
			ExpResp: &proto.HelloResponse{
				Code: http.StatusOK,
				Greeting: "Hello Test Bot",
				Id: srvID,
			},
		},
		{
			Desc: "Invalid token reported",
			TknVal: &TokenValidatorMock{
				ExpErr: errors.New("Bad token!"),
				ExpClErr: true,
			},
			Req: &proto.HelloRequest{
				Token: "some.invalid.token",
				Name: "Test Bot",
			},
			ExpResp: &proto.HelloResponse{
				Code: http.StatusUnauthorized,
				Id: srvID,
				Detail: "Bad token!",
			},
		},
		{
			Desc: "Token validation error",
			TknVal: &TokenValidatorMock{
				ExpErr: errors.New("Internal error"),
				ExpClErr: false,
			},
			Req: &proto.HelloRequest{
				Token: "some.valid.token",
				Name: "Test Bot",
			},
			ExpResp: &proto.HelloResponse{
				Code: http.StatusInternalServerError,
				Id: srvID,
				Detail: server.SomethingWickedError,
			},
		},
	}
	for _, tc := range tcs {
		s, err := server.New(srvID, tc.TknVal, logger)
		if err != nil {
			t.Fatalf("server.New(): %v", err)
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
