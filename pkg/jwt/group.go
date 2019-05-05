package jwt

import "time"

type Group struct {
	ID          string
	Name        string
	AccessLevel float32
	CreateDate  time.Time
	UpdateDate  time.Time
}

func (g Group) HasValue() bool {
	return g.ID != ""
}
