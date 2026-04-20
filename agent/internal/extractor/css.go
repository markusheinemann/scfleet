package extractor

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (e *Engine) runCSS(ext Extractor, multiple bool) (string, error) {
	sel := e.gqDoc.Find(ext.Selector)
	if sel.Length() == 0 {
		return "", nil
	}

	attr := ext.Attribute
	if attr == "" {
		attr = "innerText"
	}

	if multiple {
		var parts []string
		sel.Each(func(_ int, s *goquery.Selection) {
			if v := extractAttr(s, attr); v != "" {
				parts = append(parts, v)
			}
		})
		return encodeArray(parts), nil
	}

	return extractAttr(sel.First(), attr), nil
}

func extractAttr(s *goquery.Selection, attr string) string {
	switch attr {
	case "innerText":
		return strings.TrimSpace(s.Text())
	case "html":
		h, _ := s.Html()
		return h
	default:
		v, _ := s.Attr(attr)
		return v
	}
}

func (e *Engine) runMeta(ext Extractor) (string, error) {
	propLower := strings.ToLower(ext.Property)

	var found string

	e.gqDoc.Find("meta").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if prop, _ := s.Attr("property"); strings.ToLower(prop) == propLower {
			found, _ = s.Attr("content")
			return false
		}

		return true
	})

	if found != "" {
		return found, nil
	}

	e.gqDoc.Find("meta").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if name, _ := s.Attr("name"); strings.ToLower(name) == propLower {
			found, _ = s.Attr("content")
			return false
		}

		return true
	})

	return found, nil
}

func (e *Engine) runMicrodata(ext Extractor, multiple bool) (string, error) {
	sel := e.gqDoc.Find(fmt.Sprintf("[itemprop=%q]", ext.Itemprop))
	if sel.Length() == 0 {
		return "", nil
	}

	extract := func(s *goquery.Selection) string {
		tag := goquery.NodeName(s)
		switch tag {
		case "meta":
			v, _ := s.Attr("content")
			return v
		case "link":
			v, _ := s.Attr("href")
			return v
		default:
			if ext.Attribute != "" {
				v, _ := s.Attr(ext.Attribute)
				return v
			}

			return strings.TrimSpace(s.Text())
		}
	}

	if multiple {
		var parts []string
		sel.Each(func(_ int, s *goquery.Selection) {
			if v := extract(s); v != "" {
				parts = append(parts, v)
			}
		})

		return encodeArray(parts), nil
	}

	return extract(sel.First()), nil
}
