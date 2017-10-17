package rpc_test

import (
	"testing"

	"github.com/tomogoma/seedms/handler/rpc"
	"github.com/tomogoma/seedms/logging"
	testH "github.com/tomogoma/seedms/testing"
	"context"
	"github.com/tomogoma/seedms/api"
	"github.com/tomogoma/go-typed-errors"
)

func TestNewHandler(t *testing.T) {
	tt := []struct {
		name   string
		guard  rpc.Guard
		logger logging.Logger
		expErr bool
	}{
		{
			name:   "valid deps",
			guard:  &testH.GuardMock{},
			logger: &testH.LoggerMock{},
			expErr: false,
		},
		{
			name:   "nil guard",
			guard:  nil,
			logger: &testH.LoggerMock{},
			expErr: true,
		},
		{
			name:   "nil logger",
			guard:  &testH.GuardMock{},
			logger: nil,
			expErr: true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sh, err := rpc.NewStatusHandler(tc.guard, tc.logger)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if sh == nil {
				t.Fatalf("Got nil *rpc.StatusHandler")
			}
		})
	}
}


func TestStatusHandler_Check(t *testing.T) {
	tt := []struct {
		name string
		guard *testH.GuardMock
		req *api.Request
		expErr bool
	}{
		{
			name: "valid",
			guard: &testH.GuardMock{},
			req: &api.Request{},
			expErr: false,
		},
		{
			name: "forbidden",
			guard: &testH.GuardMock{ExpAPIKValidErr: typederrs.NewForbidden("guard")},
			req: &api.Request{},
			expErr: true,
		},
		{
			name: "unauthorized",
			guard: &testH.GuardMock{ExpAPIKValidErr: typederrs.NewUnauthorized("guard")},
			req: &api.Request{},
			expErr: true,
		},
		{
			name: "internal error",
			guard: &testH.GuardMock{ExpAPIKValidErr: typederrs.Newf("guard")},
			req: &api.Request{},
			expErr: true,
		},
	}
	for _, tc := range tt {
		sh := newStatusHandler(t, tc.guard, &testH.LoggerMock{})
		resp := new(api.Response)
		err := sh.Check(context.TODO(), tc.req, resp)
		if tc.expErr {
			if err ==nil {
				t.Fatalf("Expected an error, got nil")
			}
			return
		}
		if err != nil {
			t.Fatalf("Got error: %v", err)
		}
	}
}

func newStatusHandler(t *testing.T, g rpc.Guard, lg logging.Logger) *rpc.StatusHandler {
	sh, err := rpc.NewStatusHandler(g, lg)
	if err != nil {
		t.Fatalf("Error setting up: new status handler: %v", err)
	}
	return sh
}