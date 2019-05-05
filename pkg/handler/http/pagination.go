package http

import (
	"bitbucket.org/rfhkenya/qms-server/pkg/keys"
	"net/http"
	"strconv"
)

type Pagination struct {
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

func ReadPagination(r *http.Request) Pagination {

	req := Pagination{}

	q := r.URL.Query()

	var err error

	qOffset := q.Get(keys.Offset)
	if req.Offset, err = strconv.Atoi(qOffset); err != nil {
		req.Offset = 0
	}

	qCount := q.Get(keys.Count)
	if req.Count, err = strconv.Atoi(qCount); err != nil {
		req.Count = 10
	}

	return req
}
