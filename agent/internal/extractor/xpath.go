package extractor

import (
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func (e *Engine) runXPath(ext Extractor, multiple bool) (string, error) {
	nodes, err := htmlquery.QueryAll(e.xpathDoc, ext.Selector)
	if err != nil {
		return "", err
	}

	if len(nodes) == 0 {
		return "", nil
	}

	extract := func(n *html.Node) string {
		attr := ext.Attribute
		if attr == "" || attr == "innerText" {
			return htmlquery.InnerText(n)
		}

		return htmlquery.SelectAttr(n, attr)
	}

	if multiple {
		parts := make([]string, 0, len(nodes))
		for _, n := range nodes {
			if v := extract(n); v != "" {
				parts = append(parts, v)
			}
		}

		return encodeArray(parts), nil
	}

	return extract(nodes[0]), nil
}
