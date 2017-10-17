package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	errors "github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/config"
	"github.com/tomogoma/seedms/logging"
)

type contextKey string

type Guard interface {
	APIKeyValid(key []byte) (string, error)
}

type handler struct {
	errors.NotImplErrCheck
	errors.AuthErrCheck
	errors.ClErrCheck

	guard  Guard
	logger logging.Logger
}

const (
	internalErrorMessage = "whoops! Something wicked happened"

	keyAPIKey = "x-api-key"

	ctxtKeyBody = contextKey("id")
	ctxKeyLog   = contextKey("log")
)

func NewHandler(g Guard, l logging.Logger) (http.Handler, error) {
	if g == nil {
		return nil, errors.New("Guard was nil")
	}
	if l == nil {
		return nil, errors.New("Logger was nil")
	}

	r := mux.NewRouter().PathPrefix(config.WebRootURL).Subrouter()
	handler{guard: g, logger: l}.handleRoute(r)

	return r, nil
}

func (s handler) handleRoute(r *mux.Router) {

	r.PathPrefix("/status").
		Methods(http.MethodGet).
		HandlerFunc(s.prepLogger(s.guardRoute(s.handleStatus)))
}

func (s handler) prepLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := s.logger.WithField(logging.FieldTransID, uuid.New())
		log.WithFields(map[string]interface{}{
			logging.FieldURL:            r.URL,
			logging.FieldHost:           r.Host,
			logging.FieldMethod:         r.Method,
			logging.FieldRequestHandler: "HTTP",
			logging.FieldHttpReqObj:     r,
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

func (s *handler) readReqBody(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		dataB, err := ioutil.ReadAll(r.Body)
		if err != nil {
			err = errors.NewClientf("Failed to read request body: %v", err)
			s.handleError(w, r, nil, err)
			return
		}
		ctx := context.WithValue(r.Context(), ctxtKeyBody, dataB)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// unmarshalJSONOrRespondError returns true if json is extracted from
// data into req successfully, otherwise, it writes an error response into
// w and returns false.
// The Context in r should contain a logging.Logger with key ctxKeyLog
// for logging in case of error
func (s *handler) unmarshalJSONOrRespondError(w http.ResponseWriter, r *http.Request, data []byte, req interface{}) bool {
	err := json.Unmarshal(data, req)
	if err != nil {
		err = errors.NewClientf("failed to unmarshal JSON request from body: %v", err)
		s.handleError(w, r, nil, err)
		return false
	}
	return true
}

func (s *handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.respondOn(w, r, nil, struct {
		Name          string `json:"name"`
		Version       string `json:"version"`
		Description   string `json:"description"`
		CanonicalName string `json:"canonicalName"`
	}{
		Name:          config.Name,
		Version:       config.Version,
		Description:   config.Description,
		CanonicalName: config.CanonicalWebName,
	}, http.StatusOK, nil)
}

func (s *handler) handleError(w http.ResponseWriter, r *http.Request, reqData interface{}, err error) {
	reqDataB, _ := json.Marshal(reqData)
	log := r.Context().Value(ctxKeyLog).(logging.Logger).
		WithField(logging.FieldRequest, string(reqDataB))
	if s.IsAuthError(err) {
		if s.IsForbiddenError(err) {
			log.Warnf("Forbidden: %v", err)
			http.Error(w, err.Error(), http.StatusForbidden)
		} else {
			log.Warnf("Unauthorized: %v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}
		return
	}
	if s.IsClientError(err) {
		log.Warnf("Bad request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.IsNotImplementedError(err) {
		log.Warnf("Not implemented entity: %v", err)
		http.Error(w, err.Error(), http.StatusNotImplemented)
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

	w.Header().Set("Access-Control-Allow-Origin", "*")
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
	fmt.Fprint(w, "Nothing to see here")
}
