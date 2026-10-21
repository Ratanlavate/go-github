package github

import (
	"net/http"
	"testing"
)

func TestAddOptions(t *testing.T) {
	type test struct {
		s   string
		opt interface{}
		url string
	}
	tests := []test{
		{"/repos", nil, "/repos"},
		{"/repos", &ListOptions{}, "/repos"},
		{"/repos", &ListOptions{Page: 2}, "/repos?page=2"},
		{"/repos", &ListOptions{Page: 2, PerPage: 50}, "/repos?page=2&per_page=50"},
		{"/repos", &ListOptions{Page: 1, PerPage: 0}, "/repos?page=1"},
		{"/repos", &ListOptions{PerPage: 0}, "/repos"},
	}
	for _, tt := range tests {
		u, err := addOptions(tt.s, tt.opt)
		if err != nil {
			t.Errorf("addOptions(%q, %+v) returned err: %v", tt.s, tt.opt, err)
		} else if u != tt.url {
			t.Errorf("addOptions(%q, %+v) returned %q, want %q", tt.s, tt.opt, u, tt.url)
		}
	}
}

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantNext  int
		wantPrev  int
		wantLast  int
		wantFirst int
	}{
		{
			name:     "empty",
			header:   "",
			wantNext: 0,
			wantPrev: 0,
		},
		{
			name:     "next only",
			header:   `<https://api.github.com/repos?page=2&per_page=30>; rel="next"`,
			wantNext: 2,
		},
		{
			name:     "next and last",
			header:   `<https://api.github.com/repos?page=2&per_page=30>; rel="next", <https://api.github.com/repos?page=5&per_page=30>; rel="last"`,
			wantNext: 2,
			wantLast: 5,
		},
		{
			name:     "prev, next, first, last",
			header:   `<https://api.github.com/repos?page=3&per_page=30>; rel="next", <https://api.github.com/repos?page=1&per_page=30>; rel="prev", <https://api.github.com/repos?page=1&per_page=30>; rel="first", <https://api.github.com/repos?page=5&per_page=30>; rel="last"`,
			wantNext: 3,
			wantPrev: 1,
			wantFirst: 1,
			wantLast: 5,
		},
		{
			name:     "per_page omitted in link",
			header:   `<https://api.github.com/repos?page=2>; rel="next", <https://api.github.com/repos?page=5>; rel="last"`,
			wantNext: 2,
			wantLast: 5,
		},
	}

	for _, tt := range tests {
		r := &http.Response{
			Header: make(http.Header),
		}
		if tt.header != "" {
			r.Header.Set("Link", tt.header)
		}
		resp, err := parsePagination(r)
		if err != nil {
			t.Fatalf("%s: parsePagination returned err: %v", tt.name, err)
		}
		if resp.NextPage != tt.wantNext {
			t.Errorf("%s: NextPage = %d, want %d", tt.name, resp.NextPage, tt.wantNext)
		}
		if resp.PrevPage != tt.wantPrev {
			t.Errorf("%s: PrevPage = %d, want %d", tt.name, resp.PrevPage, tt.wantPrev)
		}
		if resp.FirstPage != tt.wantFirst {
			t.Errorf("%s: FirstPage = %d, want %d", tt.name, resp.FirstPage, tt.wantFirst)
		}
		if resp.LastPage != tt.wantLast {
			t.Errorf("%s: LastPage = %d, want %d", tt.name, resp.LastPage, tt.wantLast)
		}
	}
}
