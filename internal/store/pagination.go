package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int      `json:"limit" validate:"min=1,max=100"`
	Offset int      `json:"offset" validate:"min=0"`
	Sort   string   `json:"sort" validate:"omitempty,oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
}

func (p PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	qs := r.URL.Query()
	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(qs.Get("limit"))
		if err != nil {
			return p, nil
		}
		p.Limit = l
	}

	offset := qs.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(qs.Get("offset"))
		if err != nil {
			return p, nil
		}
		p.Offset = o
	}

	sort := qs.Get("sort")
	if sort != "" {
		p.Sort = sort
	}

	tags := qs.Get("tags")
	if tags != "" {
		p.Tags = strings.Split(tags, ",")
	} else {
		p.Tags = []string{}
	}

	search := qs.Get("search")
	if search != "" {
		p.Search = search
	}
	//
	//since := qs.Get("since")
	//if since != "" {
	//	p.Since = parseTime(since)
	//}
	//
	//until := qs.Get("until")
	//if until != "" {
	//	p.Until = parseTime(until)
	//}

	return p, nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}
	return t.Format(time.DateTime)
}
