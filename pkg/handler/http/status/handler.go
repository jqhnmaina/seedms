package status

import (
	"github.com/gorilla/mux"
	"github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/pkg/config"
	httpAPI "github.com/tomogoma/seedms/pkg/handler/http"
	"net/http"
)

type Handler struct {
	errors.ErrToHTTP
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleRoute(api *httpAPI.API, r *mux.Router) {
	h.handleIndex(api, r)
}

func (h *Handler) handleIndex(api *httpAPI.API, r *mux.Router) {
	r.Methods(http.MethodGet).
		HandlerFunc(
			api.RouteChain(func(w http.ResponseWriter, r *http.Request) {
				httpAPI.RespondJsonOn(w, r, nil, struct {
					Name          string `json:"name"`
					Version       string `json:"version"`
					Description   string `json:"description"`
					CanonicalName string `json:"canonicalName"`
				}{
					Name:          config.Name,
					Version:       config.VersionFull,
					Description:   config.Description,
					CanonicalName: config.CanonicalWebName(),
				}, http.StatusOK, nil, h)
			}),
		)
}
