package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	errors "github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/config"
	"github.com/tomogoma/seedms/logging"
	testingH "github.com/tomogoma/seedms/testing"
)

func TestNewHandler(t *testing.T) {
	tt := []struct {
		name   string
		guard  Guard
		logger logging.Logger
		expErr bool
	}{
		{
			name:   "valid deps",
			guard:  &testingH.GuardMock{},
			logger: &testingH.LoggerMock{},
			expErr: false,
		},
		{
			name:   "nil guard",
			guard:  nil,
			logger: &testingH.LoggerMock{},
			expErr: true,
		},
		{
			name:   "nil logger",
			guard:  &testingH.GuardMock{},
			logger: nil,
			expErr: true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewHandler(tc.guard, tc.logger)
			if tc.expErr {
				if err == nil {
					t.Fatal("Expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if h == nil {
				t.Fatalf("http.NewHandler() yielded a nil handler!")
			}
		})
	}
}

func TestHandler_handleRoute(t *testing.T) {
	tt := []struct {
		name          string
		reqURLSuffix  string
		reqMethod     string
		reqBody       string
		reqWBasicAuth bool
		expStatusCode int
		guard         Guard
	}{
		// values starting and ending with "_" are place holders for variables
		// e.g. _loginType_ is a place holder for "any (valid) login type"

		{
			name:          "status",
			guard:         &testingH.GuardMock{},
			reqURLSuffix:  "/" + config.Version + "/" + config.Name + "/status",
			reqMethod:     http.MethodGet,
			expStatusCode: http.StatusOK,
		},
		{
			name:          "status guard error",
			guard:         &testingH.GuardMock{ExpAPIKValidErr: errors.Newf("guard error")},
			reqURLSuffix:  "/" + config.Version + "/" + config.Name + "/status",
			reqMethod:     http.MethodGet,
			expStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			lg := &testingH.LoggerMock{}
			h := newHandler(t, tc.guard, lg)
			srvr := httptest.NewServer(h)
			defer srvr.Close()

			req, err := http.NewRequest(
				tc.reqMethod,
				srvr.URL+tc.reqURLSuffix,
				bytes.NewReader([]byte(tc.reqBody)),
			)
			if err != nil {
				t.Fatalf("Error setting up: new request: %v", err)
			}
			if tc.reqWBasicAuth {
				req.SetBasicAuth("username", "password")
			}

			cl := &http.Client{}
			resp, err := cl.Do(req)
			if err != nil {
				lg.PrintLogs(t)
				t.Fatalf("Do request error: %v", err)
			}

			if resp.StatusCode != tc.expStatusCode {
				lg.PrintLogs(t)
				t.Errorf("Expected status code %d, got %s",
					tc.expStatusCode, resp.Status)
			}
		})
	}
}

func newHandler(t *testing.T, g Guard, lg logging.Logger) http.Handler {
	h, err := NewHandler(g, lg)
	if err != nil {
		t.Fatalf("http.NewHandler(): %v", err)
	}
	return h
}
