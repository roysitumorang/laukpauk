package models

import (
	"maps"
	"net/url"
	"strconv"
	"strings"
)

type (
	Pagination struct {
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
	}
)

func NewPagination(total, pages, limit, page int64, baseURL string, urlValues url.Values) (*Pagination, error) {
	var (
		response    Pagination
		err         error
		builder     strings.Builder
		u           url.Values
		queryString string
	)
	response.Info.Total = total
	response.Info.Page = page
	response.Info.Pages = pages
	response.Info.Limit = limit
	response.Links.First = baseURL
	response.Links.Current = baseURL
	if len(urlValues) > 0 {
		u = maps.Clone(urlValues)
		if queryString, err = url.QueryUnescape(u.Encode()); err != nil {
			return nil, err
		}
		builder.Reset()
		_, _ = builder.WriteString(baseURL)
		_, _ = builder.WriteString("?")
		_, _ = builder.WriteString(queryString)
		response.Links.Current = builder.String()
		u.Del("page")
		if queryString, err = url.QueryUnescape(u.Encode()); err != nil {
			return nil, err
		}
		builder.Reset()
		_, _ = builder.WriteString(baseURL)
		_, _ = builder.WriteString("?")
		_, _ = builder.WriteString(queryString)
		response.Links.First = builder.String()
	}
	if n := pages - 1; page < n {
		u = maps.Clone(urlValues)
		u.Set("page", strconv.FormatInt(n, 10))
		if queryString, err = url.QueryUnescape(u.Encode()); err != nil {
			return nil, err
		}
		builder.Reset()
		_, _ = builder.WriteString(baseURL)
		_, _ = builder.WriteString("?")
		_, _ = builder.WriteString(queryString)
		response.Links.Last = builder.String()
		u.Set("page", strconv.FormatInt(page+1, 10))
		if queryString, err = url.QueryUnescape(u.Encode()); err != nil {
			return nil, err
		}
		builder.Reset()
		_, _ = builder.WriteString(baseURL)
		_, _ = builder.WriteString("?")
		_, _ = builder.WriteString(queryString)
		response.Links.Next = builder.String()
	}
	if page > 0 {
		u = maps.Clone(urlValues)
		u.Set("page", strconv.FormatInt(page, 10))
		if queryString, err = url.QueryUnescape(u.Encode()); err != nil {
			return nil, err
		}
		builder.Reset()
		_, _ = builder.WriteString(baseURL)
		_, _ = builder.WriteString("?")
		_, _ = builder.WriteString(queryString)
		response.Links.Current = builder.String()
		if page > 1 {
			u.Set("page", strconv.FormatInt(page-1, 10))
			if queryString, err = url.QueryUnescape(u.Encode()); err != nil {
				return nil, err
			}
			builder.Reset()
			_, _ = builder.WriteString(baseURL)
			_, _ = builder.WriteString("?")
			_, _ = builder.WriteString(queryString)
			response.Links.Previous = builder.String()
		}
	}
	return &response, nil
}
