package http

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/seedms/pkg/jwt"
	"github.com/tomogoma/seedms/pkg/keys"
	"github.com/tomogoma/seedms/pkg/logging"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Guard interface {
	APIKeyValid(key []byte) (string, error)
}

type API struct {
	errors.ErrToHTTP

	guard     Guard
	Logger    logging.Logger
	jwtHelper JWTHelper
	docsDir   string
}

var ToHttpResponser errors.ErrToHTTP

func NewHandler(g Guard, baseURL string, lg logging.Logger, jwtH JWTHelper, docsDir string, allowedOrigins []string, subRs ...SubRoute) (http.Handler, error) {

	if lg == nil {
		return nil, errors.New("nil Logger")
	}
	if jwtH == nil {
		return nil, errors.New("nil JWTValidater")
	}

	r := mux.NewRouter().
		PathPrefix(baseURL).
		Subrouter()

	h := &API{Logger: lg, guard: g, docsDir: docsDir, jwtHelper: jwtH}

	for _, subR := range subRs {
		r := r.PathPrefix(subR.Path).
			Subrouter()
		subR.Handler.HandleRoute(h, r)
	}

	h.handleNotFound(r)

	corsOpts := []handlers.CORSOption{
		handlers.AllowedHeaders([]string{
			"X-Requested-With", "Accept", "Content-Type", "Content-Length",
			"Accept-Encoding", "X-CSRF-Token", "Authorization", "X-api-key",
		}),
		handlers.AllowedOrigins(allowedOrigins),
		handlers.AllowedMethods([]string{http.MethodPost, http.MethodGet,
			http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodOptions}),
	}
	return handlers.CORS(corsOpts...)(r), nil
}

func (a *API) RouteChain(next http.HandlerFunc) http.HandlerFunc {
	return a.PrepLogger(next)
}

func (a *API) RouteChainWIthJWTGuard(next http.HandlerFunc) http.HandlerFunc {
	return a.PrepLogger(a.GuardJWT(next))
}

func (a *API) RouteChainWithAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return a.PrepLogger(a.GuardAPIKey(next))
}

func (a *API) GuardAPIKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		APIKey := r.Header.Get(keyAPIKey)
		clUsrID, err := a.guard.APIKeyValid([]byte(APIKey))
		log := r.Context().Value(ctxKeyLog).(logging.Logger).
			WithField(logging.FieldClientAppUserID, clUsrID)
		ctx := context.WithValue(r.Context(), ctxKeyLog, log)
		if err != nil {
			HandleError(w, r.WithContext(ctx), nil, err, a)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (a *API) PrepLogger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log := a.Logger.WithHTTPRequest(r).
			WithField(keys.TransactionID, uuid.New())

		log.WithFields(map[string]interface{}{
			keys.URLPath:    r.URL.Path,
			keys.HTTPMethod: r.Method,
		}).Info("new request")

		ctx := context.WithValue(r.Context(), ctxKeyLog, log)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (a *API) GuardJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := ReadToken(r)
		if err != nil {
			HandleError(w, r, token, err, a)
			return
		}

		claims, err := a.jwtHelper.Valid(token)
		if err != nil {
			HandleError(w, r, token, err, a.jwtHelper)
			return
		}

		ctx := r.Context()

		log := ctx.Value(ctxKeyLog).(logging.Logger)
		log = log.WithField(keys.ByUser, Marshal(log, claims))

		ctx = context.WithValue(ctx, ctxKeyLog, log)

		ctx = context.WithValue(ctx, ctxKeyClaims, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (a *API) handleNotFound(r *mux.Router) {
	r.NotFoundHandler = http.HandlerFunc(
		a.PrepLogger(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Nothing to see here", http.StatusNotFound)
		}),
	)
}

// respondJsonOn marshals respData to json and writes it and the code as the
// http header to w. If err is not nil, HandleError is called instead of the
// documented write to w.
func RespondJsonOn(w http.ResponseWriter, r *http.Request, reqData interface{},
	respData interface{}, code int, err error, errSrc errors.ToHTTPResponser) int {

	if err != nil {
		HandleError(w, r, reqData, err, errSrc)
		return 0
	}

	if respData == nil {
		w.WriteHeader(code)
		return 0
	}

	respBytes, err := json.Marshal(respData)
	if err != nil {
		log := r.Context().Value(ctxKeyLog).(logging.Logger)
		log.Infof("intended response: %+v", respData)
		err = errors.Newf("failed to marshal JSON response: %v", err)
		HandleError(w, r, reqData, err, errSrc)
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

// HandleError writes an error to w using errSrc's logic and logs the error
// using the logger acquired by the PrepLogger middleware on r. reqData is
// included in the log data.
func HandleError(w http.ResponseWriter, r *http.Request, reqData interface{}, err error, errSrc errors.ToHTTPResponser) {
	reqDataB, _ := json.Marshal(reqData)
	log := r.Context().Value(ctxKeyLog).(logging.Logger).
		WithField(keys.Request, string(reqDataB))

	if code, ok := errSrc.ToHTTPResponse(err, w); ok {
		log.WithField(keys.ResponseCode, code).Warn(err)
		return
	}

	log.WithField(keys.ResponseCode, http.StatusInternalServerError).Error(err)
	http.Error(w, "Something wicked happened, please try again later", http.StatusInternalServerError)
}

func ReadToken(r *http.Request) (string, error) {

	authHeaders := r.Header[HeaderAuthorization]
	bearerPrefixLen := len(HeaderAuthorizationBearerPrefix)
	for _, authHeader := range authHeaders {
		if len(authHeader) <= bearerPrefixLen {
			continue
		}

		if strings.HasPrefix(strings.ToLower(authHeader), HeaderAuthorizationBearerPrefix) {
			return authHeader[bearerPrefixLen:], nil
		}
	}
	return "", errors.NewUnauthorizedf("No %s token was found among the %s headers",
		HeaderAuthorizationBearerPrefix, HeaderAuthorization)
}

func ReadJsonBody(r *http.Request, destination interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &destination); err != nil {
		return errors.NewClientf("invalid JSON in request body: %v", err)
	}

	return nil
}

func ReadID(r *http.Request) (uint, error) {
	return ReadPathID(keys.ID, r)
}

func ReadPathID(key string, r *http.Request) (uint, error) {

	valStr := mux.Vars(r)[key]

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, errors.NewClientf("%s (%s) does not exist", key, valStr)
	}
	if val < 0 {
		return 0, errors.NewClientf("%s (%s) does not exist", key, valStr)
	}

	return uint(val), nil
}

func ReadJWTClaims(r *http.Request) (*jwt.Claim, error) {
	claims, exist := r.Context().Value(ctxKeyClaims).(*jwt.Claim)
	if !exist {
		return nil, errors.Newf("claims were not embeded in request")
	}
	return claims, nil
}

// Marshal attempts to convert string to JSON using json Marshal
// and falls back to printing the struct if this fails.
func Marshal(lg logging.Logger, val interface{}) string {
	valJson, err := json.Marshal(val)
	if err != nil {
		lg.Warnf("failed to marshal %+v, falling back to printing struct values: %v", val, err)
		return fmt.Sprintf("%+v", val)
	}
	return string(valJson)
}
