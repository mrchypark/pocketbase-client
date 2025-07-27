package pocketbase

import (
	"fmt"
	"net/url"
	"strconv"
)

// applyListOptions는 ListOptions를 url.Values에 적용합니다.
func applyListOptions(q url.Values, opts *ListOptions) {
	if opts == nil {
		return
	}
	if opts.Page > 0 {
		q.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.PerPage > 0 {
		q.Set("perPage", strconv.Itoa(opts.PerPage))
	}
	if opts.Sort != "" {
		q.Set("sort", opts.Sort)
	}
	if opts.Filter != "" {
		q.Set("filter", opts.Filter)
	}
	if opts.Expand != "" {
		q.Set("expand", opts.Expand)
	}
	if opts.Fields != "" {
		q.Set("fields", opts.Fields)
	}
	if opts.SkipTotal {
		q.Set("skipTotal", "1")
	}
	for k, v := range opts.QueryParams {
		q.Set(k, v)
	}
}

// buildQueryString은 ListOptions를 URL 쿼리 스트링으로 변환합니다.
func buildQueryString(opts *ListOptions) string {
	if opts == nil {
		return ""
	}
	q := url.Values{}
	applyListOptions(q, opts)
	return q.Encode()
}

// buildPathWithQuery는 기본 경로에 쿼리 스트링을 추가합니다.
func buildPathWithQuery(basePath string, queryString string) string {
	if queryString == "" {
		return basePath
	}
	return basePath + "?" + queryString
}

// wrapError는 일관된 에러 래핑을 제공합니다.
func wrapError(operation, entity string, err error) error {
	return fmt.Errorf("pocketbase: %s %s: %w", operation, entity, err)
}
