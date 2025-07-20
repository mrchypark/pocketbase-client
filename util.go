package pocketbase

import (
	"net/url"
	"strconv"
)

// ApplyListOptions applies the given ListOptions to query parameters.
func ApplyListOptions(q url.Values, opts *ListOptions) {
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
