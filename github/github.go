package github

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/google/go-querystring/query"
)

// ListOptions specifies the optional parameters to various lists methods.
// When PerPage is 0, the per_page parameter is omitted and the API default applies.
type ListOptions struct {
	// Page of results to retrieve.
	Page int `url:"page,omitempty"`

	// Number of results per page.
	PerPage int `url:"per_page,omitempty"`
}

// Response is a GitHub API response. This wraps the standard http.Response
// returned from GitHub and provides convenient access to things like
// pagination links.
type Response struct {
	Response *http.Response

	NextPage  int
	PrevPage  int
	FirstPage int
	LastPage  int
}

// addOptions adds the parameters in opt as URL query parameters to s.
// opt must be a struct whose fields may contain "url" tags.
func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

// parsePagination parses the link header.
func parsePagination(r *http.Response) (*Response, error) {
	response := &Response{Response: r}

	if links := r.Header.Get("Link"); links != "" {
		for _, link := range strings.Split(links, ",") {
			segments := strings.Split(strings.TrimSpace(link), ";")
			if len(segments) < 2 {
				continue
			}

			// link must be like <...>
			if !strings.HasPrefix(segments[0], "<") || !strings.HasSuffix(segments[0], ">") {
				continue
			}

			urlStr := segments[0][1 : len(segments[0])-1]
			linkURL, err := url.Parse(urlStr)
			if err != nil {
				continue
			}

			for _, segment := range segments[1:] {
				segment = strings.TrimSpace(segment)
				if strings.HasPrefix(segment, "rel=") {
					rel := strings.Trim(segment[4:], "\"")
					page := linkURL.Query().Get("page")
					if page == "" {
						continue
					}
					var pageNum int
					_, _ = fmt.Sscanf(page, "%d", &pageNum)
					if pageNum == 0 {
						continue
					}
					switch rel {
					case "next":
						response.NextPage = pageNum
					case "prev":
						response.PrevPage = pageNum
					case "first":
						response.FirstPage = pageNum
					case "last":
						response.LastPage = pageNum
					}
				}
			}
		}
	}

	return response, nil
}

// newResponse creates a new Response for the provided http.Response.
func newResponse(r *http.Response) *Response {
	response, err := parsePagination(r)
	if err != nil {
		response = &Response{Response: r}
	}
	return response
}
