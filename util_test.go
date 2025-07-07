package pocketbase

import (
	"net/url"
	"testing"
)

func TestApplyListOptions(t *testing.T) {
	tests := []struct {
		name string
		opts *ListOptions
		want url.Values
	}{
		{
			name: "nil options",
			opts: nil,
			want: url.Values{},
		},
		{
			name: "empty options",
			opts: &ListOptions{},
			want: url.Values{},
		},
		{
			name: "with page and perPage",
			opts: &ListOptions{Page: 2, PerPage: 10},
			want: url.Values{
				"page":    []string{"2"},
				"perPage": []string{"10"},
			},
		},
		{
			name: "with sort and filter",
			opts: &ListOptions{Sort: "-created", Filter: "status=active"},
			want: url.Values{
				"sort":   []string{"-created"},
				"filter": []string{"status=active"},
			},
		},
		{
			name: "with expand and fields",
			opts: &ListOptions{Expand: "user", Fields: "id,name"},
			want: url.Values{
				"expand": []string{"user"},
				"fields": []string{"id,name"},
			},
		},
		{
			name: "with skipTotal",
			opts: &ListOptions{SkipTotal: true},
			want: url.Values{
				"skipTotal": []string{"1"},
			},
		},
		{
			name: "with QueryParams",
			opts: &ListOptions{
				QueryParams: map[string]string{
					"customParam1": "value1",
					"customParam2": "value2",
				},
			},
			want: url.Values{
				"customParam1": []string{"value1"},
				"customParam2": []string{"value2"},
			},
		},
		{
			name: "all options combined",
			opts: &ListOptions{
				Page:    1,
				PerPage: 5,
				Sort:    "name",
				Filter:  "age>18",
				Expand:  "profile",
				Fields:  "id,email",
				SkipTotal: true,
				QueryParams: map[string]string{
					"extra1": "valA",
					"extra2": "valB",
				},
			},
			want: url.Values{
				"page":        []string{"1"},
				"perPage":     []string{"5"},
				"sort":        []string{"name"},
				"filter":      []string{"age>18"},
				"expand":      []string{"profile"},
				"fields":      []string{"id,email"},
				"skipTotal":   []string{"1"},
				"extra1":      []string{"valA"},
				"extra2":      []string{"valB"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := make(url.Values)
			applyListOptions(q, tt.opts)

			if len(q) != len(tt.want) {
				t.Errorf("applyListOptions() got %v, want %v (length mismatch)", q, tt.want)
				return
			}

			for k, v := range tt.want {
				if got := q.Get(k); got != v[0] {
					t.Errorf("applyListOptions() for key %s got %s, want %s", k, got, v[0])
				}
			}
		})
	}
}
