// Package schedule reads the NSU group directory and official calendar export.
package schedule

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// ErrGroupNotFound indicates that a requested group is absent from the IIR directory.
var ErrGroupNotFound = errors.New("group not found in IIR")

// Lesson contains a subject, local start and end times, and a classroom.
type Lesson struct {
	Subject   string `json:"subject"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Room      string `json:"room"`
}

// Client retrieves group listings and calendar exports from the NSU website.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient creates a schedule client for baseURL.
// A nil HTTP client selects the default client with a 15-second request timeout.
func NewClient(baseURL string, client *http.Client) *Client {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: client}
}

// get downloads a source path using ctx and requires an HTTP 200 response.
// It limits the response body to 4 MiB and always closes the response stream.
func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NSU %s: HTTP %d", path, res.StatusCode)
	}

	const limit = 4 << 20
	body, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil {
		return nil, err
	}

	if len(body) > limit {
		return nil, errors.New("NSU response exceeds 4 MiB")
	}
	return body, nil
}

// Groups fetches the IIR directory and returns sorted, unique group identifiers.
func (c *Client) Groups(ctx context.Context) ([]string, error) {
	body, err := c.get(ctx, "/faculty/iir")
	if err != nil {
		return nil, err
	}
	return parseGroups(bytes.NewReader(body))
}

// parseGroups walks the HTML tree and extracts identifiers from /group/ links.
// An empty directory is treated as a source format error.
func parseGroups(r io.Reader) ([]string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var walk func(*html.Node)
	// walk visits each element and its descendants to collect group links.
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key != "href" || !strings.HasPrefix(attr.Val, "/group/") {
					continue
				}
				group := strings.TrimPrefix(attr.Val, "/group/")
				if group != "" && !strings.ContainsAny(group, "/?#") {
					seen[group] = true
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	if len(seen) == 0 {
		return nil, errors.New("NSU IIR page contains no group links")
	}
	groups := make([]string, 0, len(seen))
	for group := range seen {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	return groups, nil
}

// Lessons verifies IIR membership, downloads the group calendar, and selects
// lessons starting on day in the Novosibirsk timezone.
func (c *Client) Lessons(ctx context.Context, group string, day time.Time) ([]Lesson, error) {
	groups, err := c.Groups(ctx)
	if err != nil {
		return nil, err
	}
	if i := sort.SearchStrings(groups, group); i == len(groups) || groups[i] != group {
		return nil, ErrGroupNotFound
	}
	body, err := c.get(ctx, "/ics/group/"+url.PathEscape(group))
	if err != nil {
		return nil, err
	}
	return parseCalendar(bytes.NewReader(body), day)
}
