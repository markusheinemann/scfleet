package extractor

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tidwall/gjson"
)

// jsonLDBlocks lazily parses all <script type="application/ld+json"> text content.
func (e *Engine) jsonLDBlocks() []string {
	if e.jsonLDOnce != nil {
		return e.jsonLDOnce
	}

	e.gqDoc.Find(`script[type="application/ld+json"]`).Each(func(_ int, s *goquery.Selection) {
		if text := strings.TrimSpace(s.Text()); text != "" {
			e.jsonLDOnce = append(e.jsonLDOnce, text)
		}
	})

	if e.jsonLDOnce == nil {
		e.jsonLDOnce = []string{}
	}

	return e.jsonLDOnce
}

func (e *Engine) runJSONLD(ext Extractor) (string, error) {
	blocks := e.jsonLDBlocks()
	if ext.BlockIndex >= len(blocks) {
		return "", nil
	}

	return gjson.Get(blocks[ext.BlockIndex], ext.Path).String(), nil
}

func (e *Engine) runSchemaOrg(ext Extractor) (string, error) {
	blocks := e.jsonLDBlocks()
	schemaTypeLower := strings.ToLower(ext.SchemaType)
	wildcard := schemaTypeLower == "*" || schemaTypeLower == ""

	for _, block := range blocks {
		result := gjson.Parse(block)

		// Top-level array: treat each element as an independent block.
		if result.IsArray() {
			var found string
			result.ForEach(func(_, item gjson.Result) bool {
				found = e.matchSchemaItem(item, ext.Path, schemaTypeLower, wildcard)
				return found == ""
			})
			if found != "" {
				return found, nil
			}
			continue
		}

		if graph := result.Get("@graph"); graph.Exists() {
			var found string
			graph.ForEach(func(_, item gjson.Result) bool {
				found = e.matchSchemaItem(item, ext.Path, schemaTypeLower, wildcard)
				return found == ""
			})
			if found != "" {
				return found, nil
			}
			continue
		}

		if wildcard || matchesSchemaType(result, schemaTypeLower) {
			if val := result.Get(ext.Path); val.Exists() {
				return val.String(), nil
			}
		}
	}

	return "", nil
}

func (e *Engine) matchSchemaItem(item gjson.Result, path, schemaTypeLower string, wildcard bool) string {
	if wildcard || matchesSchemaType(item, schemaTypeLower) {
		if v := item.Get(path); v.Exists() {
			return v.String()
		}
	}
	return ""
}

func matchesSchemaType(result gjson.Result, typeLower string) bool {
	t := result.Get("@type")
	if !t.Exists() {
		return false
	}

	if t.Type == gjson.String {
		return strings.ToLower(t.String()) == typeLower
	}

	if t.IsArray() {
		for _, v := range t.Array() {
			if strings.ToLower(v.String()) == typeLower {
				return true
			}
		}
	}

	return false
}
