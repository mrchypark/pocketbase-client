package pocketbase

import (
	"fmt"
	"net/url"
	"strconv"
)

// applyListOptions applies ListOptions to url.Values.
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

// buildQueryString converts ListOptions to URL query string.
func buildQueryString(opts *ListOptions) string {
	if opts == nil {
		return ""
	}
	q := url.Values{}
	applyListOptions(q, opts)
	return q.Encode()
}

// buildPathWithQuery adds query string to the base path.
func buildPathWithQuery(basePath string, queryString string) string {
	if queryString == "" {
		return basePath
	}
	return basePath + "?" + queryString
}

// wrapError provides consistent error wrapping.
func wrapError(operation, entity string, err error) error {
	return fmt.Errorf("pocketbase: %s %s: %w", operation, entity, err)
}
