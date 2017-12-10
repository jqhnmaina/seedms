package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	errors "github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/logging"
	"github.com/tomogoma/seedms/config"
	"net/url"
)

type contextKey string

type Guard interface {
	APIKeyValid(key []byte) (string, error)
}

type handler struct {
	errors.AllErrCheck

	guard  Guard
	logger logging.Logger
}

const (
	internalErrorMessage = "whoops! Something wicked happened"

	keyAPIKey = "x-api-key"
	keyOffset          = "offset"
	keyCount           = "count"

	ctxtKeyBody = contextKey("id")
	ctxKeyLog   = contextKey("log")

	defaultOffset = "0"
	defaultCount  = "10"
)

func NewHandler(g Guard, l logging.Logger, baseURL string, allowedOrigins []string) (http.Handler, error) {
	if g == nil {
		return nil, errors.New("Guard was nil")
	}
	if l == nil {
		return nil, errors.New("Logger was nil")
	}

	r := mux.NewRouter().PathPrefix(baseURL).Subrouter()
	handler{guard: g, logger: l}.handleRoute(r)

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

// TODO move middleware to own file
func (s handler) handleRoute(r *mux.Router) {

	r.PathPrefix("/status").
		Methods(http.MethodGet).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleStatus)))

	r.PathPrefix("/" + config.DocsPath).
		Handler(http.FileServer(http.Dir(config.DefaultDocsDir())))

	r.NotFoundHandler = http.HandlerFunc(s.prepLogger(s.notFoundHandler))
}

func (s *handler) midwareChain(next http.HandlerFunc) http.HandlerFunc {
	return s.prepLogger(s.guardRoute(next))
}

func (s handler) prepLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log := s.logger.WithHTTPRequest(r).
			WithField(logging.FieldTransID, uuid.New())

		log.WithFields(map[string]interface{}{
			logging.FieldURLPath:        r.URL.Path,
			logging.FieldHTTPMethod:         r.Method,
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
			s.handleError(w, r.WithContext(ctx), nil, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// unmarshalBodyOrRespondError returns true if json is extracted from
// request body, otherwise, it writes an error response into w and returns false.
// The Context in r should contain a logging.Logger with key ctxKeyLog
// for logging in case of error
func (s *handler) unmarshalBodyOrRespondError(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err = errors.NewClientf("unable to read request body: %v", err)
		s.handleError(w, r, nil, err)
		return false
	}
	if err = json.Unmarshal(data, req); err != nil {
		err = errors.NewClientf("unable to unmarshal request body: %v", err)
		s.handleError(w, r, string(data), err)
		return false
	}
	return true
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
func (s *handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.respondOn(w, r, nil, struct {
		Name          string `json:"name"`
		Version       string `json:"version"`
		Description   string `json:"description"`
		CanonicalName string `json:"canonicalName"`
	}{
		Name:          config.Name,
		Version:       config.VersionFull,
		Description:   config.Description,
		CanonicalName: config.CanonicalWebName(),
	}, http.StatusOK, nil)
}

func (s *handler) handleError(w http.ResponseWriter, r *http.Request, reqData interface{}, err error) {
	reqDataB, _ := json.Marshal(reqData)
	log := r.Context().Value(ctxKeyLog).(logging.Logger).
		WithField(logging.FieldRequest, string(reqDataB))
	if s.IsUnauthorizedError(err) {
		log.Warnf("Unauthorized: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if s.IsForbiddenError(err) || s.IsAuthError(err) {
		log.Warnf("Forbidden: %v", err)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if s.IsClientError(err) {
		log.Warnf("Bad request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.IsNotFoundError(err) {
		log.Warnf("Not found: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.IsNotImplementedError(err) {
		log.Warnf("Not implemented entity: %v", err)
		resp := struct {
			Message string      `json:"message"`
			ReqData interface{} `json:"reqData"`
		}{
			Message: "Not yet implemented",
			ReqData: reqData,
		}
		s.respondOn(w, r, reqData, resp, http.StatusNotImplemented, nil)
		return
	}
	log.Errorf("Internal error: %v", err)
	http.Error(w, internalErrorMessage, http.StatusInternalServerError)
}

func (s *handler) respondOn(w http.ResponseWriter, r *http.Request, reqData interface{}, respData interface{}, code int, err error) int {

	if err != nil {
		s.handleError(w, r, reqData, err)
		return 0
	}

	respBytes, err := json.Marshal(respData)
	if err != nil {
		s.handleError(w, r, reqData, err)
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

func (s handler) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Nothing to see here", http.StatusNotFound)
}

func getOffset(q url.Values) string {
	offset := q.Get(keyOffset)
	if offset == "" {
		offset = defaultOffset
	}
	return offset
}

func getCount(q url.Values) string {
	count := q.Get(keyCount)
	if count == "" {
		count = defaultCount
	}
	return count
}
