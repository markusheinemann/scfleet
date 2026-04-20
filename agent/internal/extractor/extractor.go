package extractor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Template mirrors the extraction schema structure from the orchestrator.
type Template struct {
	Version        string  `json:"version"`
	JSWaitSelector string  `json:"js_wait_selector"`
	PageTimeoutS   int     `json:"page_timeout_s"`
	Fields         []Field `json:"fields"`
}

// Field is one item to extract from the page.
type Field struct {
	Name       string      `json:"name"`
	Type       string      `json:"type"`      // string|number|boolean|array
	ItemType   string      `json:"item_type"` // element type when Type==array
	Required   bool        `json:"required"`
	Extractors []Extractor `json:"extractors"`
	Transform  *Transform  `json:"transform"`
}

// Extractor defines one strategy for extracting a value.
type Extractor struct {
	Strategy   string `json:"strategy"`    // css|xpath|schema_org|microdata|meta|json_ld
	Selector   string `json:"selector"`
	Attribute  string `json:"attribute"`
	Multiple   bool   `json:"multiple"`
	Itemprop   string `json:"itemprop"`
	Property   string `json:"property"`
	SchemaType string `json:"schema_type"`
	Path       string `json:"path"`
	BlockIndex int    `json:"block_index"`
}

// Transform is an optional post-processing pipeline applied to the raw value.
type Transform struct {
	Trim    *bool  `json:"trim"`
	Regex   string `json:"regex"`
	Prepend string `json:"prepend"`
	Append  string `json:"append"`
}

// Result is the extraction output for a single job.
type Result struct {
	Data        map[string]any    // field name → coerced value
	FieldErrors map[string]string // field name → error message (non-required misses)
}

// ExtractionError is returned when a required field yields no value.
type ExtractionError struct {
	FieldName string
}

func (e *ExtractionError) Error() string {
	return fmt.Sprintf("required field %q: no extractor yielded a value", e.FieldName)
}

// Engine holds parsed document representations for efficient multi-strategy extraction.
type Engine struct {
	gqDoc      *goquery.Document
	xpathDoc   *html.Node
	jsonLDOnce []string // lazy-parsed JSON-LD blocks
}

// New parses the raw HTML once and returns an Engine ready to extract.
func New(rawHTML string) (*Engine, error) {
	gqDoc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("parse html (goquery): %w", err)
	}

	xpathDoc, err := htmlquery.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("parse html (htmlquery): %w", err)
	}

	return &Engine{gqDoc: gqDoc, xpathDoc: xpathDoc}, nil
}

// Extract runs all fields in the template against the parsed document.
func (e *Engine) Extract(tmpl *Template) (*Result, error) {
	result := &Result{
		Data:        make(map[string]any),
		FieldErrors: make(map[string]string),
	}

	for _, field := range tmpl.Fields {
		raw, err := e.extractField(field)
		if err != nil {
			if field.Required {
				return nil, &ExtractionError{FieldName: field.Name}
			}
			result.FieldErrors[field.Name] = err.Error()
			continue
		}

		if raw == "" {
			if field.Required {
				return nil, &ExtractionError{FieldName: field.Name}
			}
			result.FieldErrors[field.Name] = "no extractor yielded a value"
			continue
		}

		transformed := applyTransform(raw, field.Transform)
		if transformed == "" && field.Required {
			return nil, &ExtractionError{FieldName: field.Name}
		}

		coerced, err := coerce(transformed, field.Type, field.ItemType)
		if err != nil {
			if field.Required {
				return nil, &ExtractionError{FieldName: field.Name}
			}
			result.FieldErrors[field.Name] = err.Error()
			continue
		}

		result.Data[field.Name] = coerced
	}

	return result, nil
}

// extractField tries each extractor in order; returns the first non-empty value.
func (e *Engine) extractField(field Field) (string, error) {
	isArray := field.Type == "array"

	for _, ext := range field.Extractors {
		val, err := e.runExtractor(ext, isArray)
		if err != nil {
			continue
		}

		if val != "" {
			return val, nil
		}
	}

	return "", nil
}

func (e *Engine) runExtractor(ext Extractor, multiple bool) (string, error) {
	switch ext.Strategy {
	case "css":
		return e.runCSS(ext, multiple)
	case "xpath":
		return e.runXPath(ext, multiple)
	case "schema_org":
		return e.runSchemaOrg(ext)
	case "microdata":
		return e.runMicrodata(ext, multiple)
	case "meta":
		return e.runMeta(ext)
	case "json_ld":
		return e.runJSONLD(ext)
	default:
		return "", fmt.Errorf("unknown strategy: %s", ext.Strategy)
	}
}

// encodeArray marshals a string slice to a JSON array string for uniform coerce() handling.
func encodeArray(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	b, _ := json.Marshal(parts)
	return string(b)
}
