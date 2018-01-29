package arris

import (
	"net/http"
	"net/url"
)

// A Client is an HTTP client that can retrieve statistics from the web
// interface of an Arris modem.
type Client struct {
	addr string
	c    *http.Client
}

// New creates a new Client pointed at the specified address.  If a nil
// http.Client is specified, a default client will be used.
func New(addr string, c *http.Client) (*Client, error) {
	if c == nil {
		c = &http.Client{}
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	// Always gather information from the status page.
	u.Path = "/cgi-bin/status_cgi"

	return &Client{
		addr: u.String(),
		c:    c,
	}, nil
}

// Status fetches a Status structure from the modem.
func (c *Client) Status() (*Status, error) {
	res, err := c.c.Get(c.addr)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return Parse(res.Body)
}
