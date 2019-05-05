package http

import "github.com/gorilla/mux"

type SubHandler interface {
	HandleRoute(*API, *mux.Router)
}
