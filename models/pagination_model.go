package models

import (
	"fmt"
	"maps"
	"net/url"
	"strconv"

	"github.com/roysitumorang/laukpauk/pools"
)

type Pagination struct {
	Links struct {
		First    string `json:"first" example:"http://localhost:19000/v1/cities?limit=1&search=Tangerang"`
		Previous string `json:"previous" example:"http://localhost:19000/v1/cities?limit=1&search=Tangerang"`
		Current  string `json:"current" example:"http://localhost:19000/v1/cities?limit=1&page=2&search=Tangerang"`
		Next     string `json:"next" example:"http://localhost:19000/v1/cities?limit=1&page=3&search=Tangerang"`
		Last     string `json:"last" example:"http://localhost:19000/v1/cities?limit=1&page=4&search=Tangerang"`
	} `json:"links"`
	Info struct {
		Limit int64 `json:"limit" example:"1"`
		Page  int64 `json:"page" example:"1"`
		Pages int64 `json:"pages" example:"3"`
		Total int64 `json:"total" example:"3"`
	} `json:"info"`
	Rows []any `json:"rows"`
}

func NewPagination(total, pages, limit, page int64, baseURL string, urlValues url.Values) (*Pagination, error) {
	var response Pagination
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	response.Info.Total = total
	response.Info.Page = page
	response.Info.Pages = pages
	response.Info.Limit = limit
	response.Links.First = baseURL
	response.Links.Current = baseURL
	if len(urlValues) > 0 {
		u := maps.Clone(urlValues)
		queryString, err := url.QueryUnescape(u.Encode())
		if err != nil {
			return nil, err
		}
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%s?%s", baseURL, queryString)
		response.Links.Current = sb.String()
		u.Del("page")
		if queryString, err = url.QueryUnescape(u.Encode()); err != nil {
			return nil, err
		}
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%s?%s", baseURL, queryString)
		response.Links.First = sb.String()
	}
	if n := pages - 1; page < n {
		u := maps.Clone(urlValues)
		u.Set("page", strconv.FormatInt(n, 10))
		queryString, err := url.QueryUnescape(u.Encode())
		if err != nil {
			return nil, err
		}
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%s?%s", baseURL, queryString)
		response.Links.Last = sb.String()
		u.Set("page", strconv.FormatInt(page+1, 10))
		if queryString, err = url.QueryUnescape(u.Encode()); err != nil {
			return nil, err
		}
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%s?%s", baseURL, queryString)
		response.Links.Next = sb.String()
	}
	if page > 0 {
		u := maps.Clone(urlValues)
		u.Set("page", strconv.FormatInt(page, 10))
		queryString, err := url.QueryUnescape(u.Encode())
		if err != nil {
			return nil, err
		}
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%s?%s", baseURL, queryString)
		response.Links.Current = sb.String()
		if page > 1 {
			u.Set("page", strconv.FormatInt(page-1, 10))
			if queryString, err = url.QueryUnescape(u.Encode()); err != nil {
				return nil, err
			}
			sb.Reset()
			_, _ = fmt.Fprintf(sb, "%s?%s", baseURL, queryString)
			response.Links.Previous = sb.String()
		}
	}
	return &response, nil
}

type InfiniteScroll struct {
	Current string  `json:"current"`
	Next    *string `json:"next"`
	Limit   int64   `json:"limit"`
	Rows    []any   `json:"rows"`
}

func NewInfiniteScroll(limit, nextRowNo int64, baseURL string, urlValues url.Values, rows []any) (*InfiniteScroll, error) {
	var response InfiniteScroll
	response.Rows = make([]any, len(rows))
	copy(response.Rows, rows)
	sb := pools.BuilderPool.Get()
	defer func() {
		sb.Reset()
		pools.BuilderPool.Put(sb)
	}()
	response.Limit = limit
	queryString, err := url.QueryUnescape(urlValues.Encode())
	if err != nil {
		return nil, err
	}
	sb.Reset()
	_, _ = fmt.Fprintf(sb, "%s?%s", baseURL, queryString)
	response.Current = sb.String()
	if nextRowNo > 0 {
		u := maps.Clone(urlValues)
		u.Set("next_row_no", strconv.FormatInt(nextRowNo, 10))
		if queryString, err = url.QueryUnescape(u.Encode()); err != nil {
			return nil, err
		}
		sb.Reset()
		_, _ = fmt.Fprintf(sb, "%s?%s", baseURL, queryString)
		response.Next = new(sb.String())
	}
	return &response, nil
}
