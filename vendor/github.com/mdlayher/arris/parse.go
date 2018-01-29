package arris

import (
	"io"
	"log"
	"strings"

	"golang.org/x/net/html"
)

// Parse parses Arris modem status from an input HTML document.
func Parse(r io.Reader) (*Status, error) {
	node, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	s := new(Status)

	var walk func(*html.Node) (*Status, error)
	walk = func(n *html.Node) (*Status, error) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)

			// Only looking for tbody elements.
			if c.Data != "tbody" {
				continue
			}

			// Grab raw table data and use it to parse status.
			rows, err := walkTBody(c)
			if err != nil {
				log.Fatalf("failed to tbody: %v", err)
			}

			if err := s.parse(rows); err != nil {
				return nil, err
			}
		}

		return s, nil
	}

	return walk(node)
}

// walkTBody walks through a tbody element to parse elements from its
// tr element children.
func walkTBody(tbody *html.Node) ([][]string, error) {
	var rows [][]string

	var walk func(*html.Node) ([][]string, error)
	walk = func(n *html.Node) ([][]string, error) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)

			// Only looking for tr elements.
			if c.Data != "tr" {
				continue
			}

			// Parse a single row of table data and append it to the existing
			// rows, but only if it isn't empty.
			r, err := parseTableRow(c)
			if err != nil {
				log.Fatalf("failed to parse: %v", err)
			}
			if len(r) == 0 {
				continue
			}

			rows = append(rows, r)
		}

		return rows, nil
	}

	return walk(tbody)
}

// parseTableRow parses a single row of data, stripping all of the unnecessary
// elements.
func parseTableRow(tr *html.Node) ([]string, error) {
	var out []string

	var walk func(*html.Node) ([]string, error)
	walk = func(n *html.Node) ([]string, error) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)

			// This is a very silly hack to remove all the unnecessary data in
			// the rows.
			d := strings.TrimSpace(c.Data)
			if d == "" || d == "b" || d == "font" || d == "td" {
				continue
			}

			out = append(out, d)
		}

		return out, nil
	}

	return walk(tr)
}
