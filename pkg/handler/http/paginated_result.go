package http

type PaginatedResult struct {
	Request    Pagination  `json:"request"`
	Count      int         `json:"count"`
	TotalCount int         `json:"total_count"`
	Data       interface{} `json:"data"`
}

func NewPaginatedResult(req Pagination, count, totalCount int, data interface{}) PaginatedResult {
	return PaginatedResult{
		Request:    req,
		Count:      count,
		TotalCount: totalCount,
		Data:       data,
	}
}
