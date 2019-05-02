package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/pkg/config"
	"github.com/tomogoma/seedms/pkg/logging"
)

type contextKey string

type Guard interface {
	APIKeyValid(key []byte) (string, error)
}

type handler struct {
	errors.ErrToHTTP

	guard   Guard
	logger  logging.Logger
	docsDir string
}

const (
	keyAPIKey = "x-api-key"

	ctxKeyLog = contextKey("log")
)

func NewHandler(g Guard, l logging.Logger, baseURL, docsDir string, allowedOrigins []string) (http.Handler, error) {
	if g == nil {
		return nil, errors.New("Guard was nil")
	}
	if l == nil {
		return nil, errors.New("Logger was nil")
	}
	if len(docsDir) == 0 {
		docsDir = config.DefaultDocsDir()
	}

	r := mux.NewRouter().PathPrefix(baseURL).Subrouter()
	handler{guard: g, logger: l, docsDir: docsDir}.handleRoute(r)

	corsOpts := []handlers.CORSOption{
		handlers.AllowedHeaders([]string{
			"X-Requested-With", "Accept", "Content-Type", "Content-Length",
			"Accept-Encoding", "X-CSRF-Token", "Authorization", "X-api-key",
		}),
		handlers.AllowedOrigins(allowedOrigins),
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
	}
	return handlers.CORS(corsOpts...)(r), nil
}

func (s handler) handleRoute(r *mux.Router) {
	s.handleStatus(r)
	s.handleDocs(r)
	s.handleNotFound(r)
}

/**
 * @api {get} /status Status
 * @apiName Status
 * @apiVersion 0.1.0
 * @apiGroup Service
 *
 * @apiHeader x-api-key the api key
 *
 * @apiSuccess (200) {String} name Micro-service name.
 * @apiSuccess (200)  {String} version http://semver.org version.
 * @apiSuccess (200)  {String} description Short description of the micro-service.
 * @apiSuccess (200)  {String} canonicalName Canonical name of the micro-service.
 *
 */
func (s *handler) handleStatus(r *mux.Router) {
	fmt.Println("Checking status")
	r.Methods(http.MethodGet).
		PathPrefix("/status").
		HandlerFunc(
			s.apiGuardChain(func(w http.ResponseWriter, r *http.Request) {
				s.respondJsonOn(w, r, nil, struct {
					Name          string `json:"name"`
					Version       string `json:"version"`
					Description   string `json:"description"`
					CanonicalName string `json:"canonicalName"`
				}{
					Name:          config.Name,
					Version:       config.VersionFull,
					Description:   config.Description,
					CanonicalName: config.CanonicalWebName(),
				}, http.StatusOK, nil, s)
			}),
		)
}

/**
 * @api {get} /docs Docs
 * @apiName Docs
 * @apiVersion 0.1.0
 * @apiGroup Service
 *
 * @apiSuccess (200) {html} docs Docs page to be viewed on browser.
 *
 */
func (s *handler) handleDocs(r *mux.Router) {
	r.PathPrefix("/" + config.DocsPath).
		Handler(http.FileServer(http.Dir(s.docsDir)))
}

func (s handler) handleNotFound(r *mux.Router) {
	r.NotFoundHandler = http.HandlerFunc(
		s.prepLogger(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Nothing to see here", http.StatusNotFound)
		}),
	)
}

func (s *handler) apiGuardChain(next http.HandlerFunc) http.HandlerFunc {
	return s.prepLogger(s.guardRoute(next))
}

func (s handler) prepLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log := s.logger.WithHTTPRequest(r).
			WithField(logging.FieldTransID, uuid.New())

		log.WithFields(map[string]interface{}{
			logging.FieldURLPath:    r.URL.Path,
			logging.FieldHTTPMethod: r.Method,
		}).Info("new request")

		ctx := context.WithValue(r.Context(), ctxKeyLog, log)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *handler) guardRoute(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		APIKey := r.Header.Get(keyAPIKey)
		clUsrID, err := s.guard.APIKeyValid([]byte(APIKey))
		log := r.Context().Value(ctxKeyLog).(logging.Logger).
			WithField(logging.FieldClientAppUserID, clUsrID)
		ctx := context.WithValue(r.Context(), ctxKeyLog, log)
		if err != nil {
			handleError(w, r.WithContext(ctx), nil, err, s)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// respondJsonOn marshals respData to json and writes it and the code as the
// http header to w. If err is not nil, handleError is called instead of the
// documented write to w.
func (s *handler) respondJsonOn(w http.ResponseWriter, r *http.Request, reqData interface{},
	respData interface{}, code int, err error, errSrc errors.ToHTTPResponser) int {

	if err != nil {
		handleError(w, r, reqData, err, errSrc)
		return 0
	}

	respBytes, err := json.Marshal(respData)
	if err != nil {
		handleError(w, r, reqData, err, errSrc)
		return 0
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	i, err := w.Write(respBytes)
	if err != nil {
		log := r.Context().Value(ctxKeyLog).(logging.Logger)
		log.Errorf("unable write data to response stream: %v", err)
		return i
	}

	return i
}

// handleError writes an error to w using errSrc's logic and logs the error
// using the logger acquired by the prepLogger middleware on r. reqData is
// included in the log data.
func handleError(w http.ResponseWriter, r *http.Request, reqData interface{}, err error, errSrc errors.ToHTTPResponser) {
	reqDataB, _ := json.Marshal(reqData)
	log := r.Context().Value(ctxKeyLog).(logging.Logger).
		WithField(logging.FieldRequest, string(reqDataB))

	if code, ok := errSrc.ToHTTPResponse(err, w); ok {
		log.WithField(logging.FieldResponseCode, code).Warn(err)
		return
	}

	log.WithField(logging.FieldResponseCode, http.StatusInternalServerError).
		Error(err)
	http.Error(w, "Something wicked happened, please try again later",
		http.StatusInternalServerError)
}
