package extractor_test

import (
	"testing"

	"github.com/markusheinemann/scfleet/agent/internal/extractor"
)

// buildEngine creates an Engine from an HTML fragment, wrapping it in a full document.
func buildEngine(t *testing.T, body string) *extractor.Engine {
	t.Helper()
	html := "<html><head></head><body>" + body + "</body></html>"
	e, err := extractor.New(html)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return e
}

func extract(t *testing.T, html string, tmpl *extractor.Template) *extractor.Result {
	t.Helper()
	e := buildEngine(t, html)
	r, err := e.Extract(tmpl)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	return r
}

// --- CSS strategy ---

func TestCSS_ExtractsInnerText(t *testing.T) {
	r := extract(t, `<h1>Hello World</h1>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "title", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "h1"},
			}},
		},
	})
	if r.Data["title"] != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", r.Data["title"])
	}
}

func TestCSS_ExtractsAttribute(t *testing.T) {
	r := extract(t, `<a href="/products/widget">Link</a>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "url", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "a", Attribute: "href"},
			}},
		},
	})
	if r.Data["url"] != "/products/widget" {
		t.Errorf("expected '/products/widget', got %q", r.Data["url"])
	}
}

func TestCSS_ExtractsMultipleValues(t *testing.T) {
	r := extract(t, `<ul><li>Alpha</li><li>Beta</li><li>Gamma</li></ul>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "items", Type: "array", ItemType: "string", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "li", Multiple: true},
			}},
		},
	})
	items, ok := r.Data["items"].([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", r.Data["items"])
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
	if items[0] != "Alpha" || items[1] != "Beta" || items[2] != "Gamma" {
		t.Errorf("unexpected items: %v", items)
	}
}

func TestCSS_FirstExtractorWins(t *testing.T) {
	r := extract(t, `<h1>Primary</h1><h2>Secondary</h2>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "heading", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "h1"},
				{Strategy: "css", Selector: "h2"},
			}},
		},
	})
	if r.Data["heading"] != "Primary" {
		t.Errorf("expected 'Primary', got %q", r.Data["heading"])
	}
}

func TestCSS_FallsBackToSecondExtractor(t *testing.T) {
	r := extract(t, `<h2>Fallback</h2>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "heading", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "h1"},     // not found
				{Strategy: "css", Selector: "h2"},     // found
			}},
		},
	})
	if r.Data["heading"] != "Fallback" {
		t.Errorf("expected 'Fallback', got %q", r.Data["heading"])
	}
}

// --- XPath strategy ---

func TestXPath_ExtractsTextNode(t *testing.T) {
	r := extract(t, `<title>XPath Title</title>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "title", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "xpath", Selector: "//title"},
			}},
		},
	})
	if r.Data["title"] != "XPath Title" {
		t.Errorf("expected 'XPath Title', got %q", r.Data["title"])
	}
}

func TestXPath_ExtractsAttribute(t *testing.T) {
	r := extract(t, `<img src="/img/product.jpg" alt="Product">`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "image", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "xpath", Selector: "//img", Attribute: "src"},
			}},
		},
	})
	if r.Data["image"] != "/img/product.jpg" {
		t.Errorf("expected '/img/product.jpg', got %q", r.Data["image"])
	}
}

// --- Meta strategy ---

func TestMeta_ExtractsByProperty(t *testing.T) {
	html := `<meta property="og:title" content="OG Title">`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "og_title", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "meta", Property: "og:title"},
			}},
		},
	})
	if r.Data["og_title"] != "OG Title" {
		t.Errorf("expected 'OG Title', got %q", r.Data["og_title"])
	}
}

func TestMeta_ExtractsByName(t *testing.T) {
	html := `<meta name="description" content="Page description">`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "description", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "meta", Property: "description"},
			}},
		},
	})
	if r.Data["description"] != "Page description" {
		t.Errorf("expected 'Page description', got %q", r.Data["description"])
	}
}

// --- Microdata strategy ---

func TestMicrodata_ExtractsItemprop(t *testing.T) {
	html := `<span itemprop="name">Widget Pro</span>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "name", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "microdata", Itemprop: "name"},
			}},
		},
	})
	if r.Data["name"] != "Widget Pro" {
		t.Errorf("expected 'Widget Pro', got %q", r.Data["name"])
	}
}

func TestMicrodata_ExtractsMetaContent(t *testing.T) {
	html := `<meta itemprop="price" content="29.99">`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "price", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "microdata", Itemprop: "price"},
			}},
		},
	})
	if r.Data["price"] != "29.99" {
		t.Errorf("expected '29.99', got %q", r.Data["price"])
	}
}

// --- JSON-LD / schema_org strategy ---

func TestSchemaOrg_ExtractsProductName(t *testing.T) {
	ldJSON := `{"@context":"https://schema.org","@type":"Product","name":"Widget Pro","price":"29.99"}`
	html := `<script type="application/ld+json">` + ldJSON + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "name", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "schema_org", SchemaType: "Product", Path: "name"},
			}},
		},
	})
	if r.Data["name"] != "Widget Pro" {
		t.Errorf("expected 'Widget Pro', got %q", r.Data["name"])
	}
}

func TestSchemaOrg_WildcardMatchesAnyType(t *testing.T) {
	ldJSON := `{"@context":"https://schema.org","@type":"Hotel","aggregateRating":{"ratingValue":"8.4","reviewCount":4383}}`
	html := `<script type="application/ld+json">` + ldJSON + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "rating", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "schema_org", SchemaType: "*", Path: "aggregateRating.ratingValue"},
			}},
		},
	})
	if r.Data["rating"] != "8.4" {
		t.Errorf("expected '8.4', got %q", r.Data["rating"])
	}
}

func TestSchemaOrg_WildcardSkipsBlocksWithoutPath(t *testing.T) {
	// First block has no aggregateRating; second (Hotel) does.
	block1 := `{"@type":"BreadcrumbList","itemListElement":[]}`
	block2 := `{"@type":"Hotel","aggregateRating":{"ratingValue":"9.1"}}`
	html := `<script type="application/ld+json">` + block1 + `</script>` +
		`<script type="application/ld+json">` + block2 + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "rating", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "schema_org", SchemaType: "*", Path: "aggregateRating.ratingValue"},
			}},
		},
	})
	if r.Data["rating"] != "9.1" {
		t.Errorf("expected '9.1', got %q", r.Data["rating"])
	}
}

func TestSchemaOrg_TopLevelArrayBlock(t *testing.T) {
	// Many sites (e.g. hotel booking) embed JSON-LD as a single top-level array.
	ldArray := `[{"@context":"https://schema.org","@type":"Hotel","aggregateRating":{"ratingValue":"8.4"}},{"@context":"https://schema.org","@type":"BreadcrumbList","itemListElement":[]}]`
	html := `<script type="application/ld+json">` + ldArray + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "rating", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "schema_org", SchemaType: "*", Path: "aggregateRating.ratingValue"},
			}},
		},
	})
	if r.Data["rating"] != "8.4" {
		t.Errorf("expected '8.4', got %q", r.Data["rating"])
	}
}

func TestSchemaOrg_TopLevelArrayWithSpecificType(t *testing.T) {
	ldArray := `[{"@type":"BreadcrumbList"},{"@type":"Hotel","name":"Grand Hotel"}]`
	html := `<script type="application/ld+json">` + ldArray + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "name", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "schema_org", SchemaType: "Hotel", Path: "name"},
			}},
		},
	})
	if r.Data["name"] != "Grand Hotel" {
		t.Errorf("expected 'Grand Hotel', got %q", r.Data["name"])
	}
}

func TestJSONLD_ExtractsByPath(t *testing.T) {
	ldJSON := `{"@type":"Organization","contactPoint":{"telephone":"+1-555-0100"}}`
	html := `<script type="application/ld+json">` + ldJSON + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "phone", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "json_ld", Path: "contactPoint.telephone"},
			}},
		},
	})
	if r.Data["phone"] != "+1-555-0100" {
		t.Errorf("expected '+1-555-0100', got %q", r.Data["phone"])
	}
}

// --- Transform pipeline ---

func TestTransform_TrimsWhitespace(t *testing.T) {
	r := extract(t, `<p>  spaced  </p>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "text", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "p"},
			}},
		},
	})
	// CSS innerText trims automatically, but transform trim should also work
	if r.Data["text"] != "spaced" {
		t.Errorf("expected 'spaced', got %q", r.Data["text"])
	}
}

func TestTransform_RegexCaptureGroup(t *testing.T) {
	trimBool := true
	r := extract(t, `<span>Price: $29.99</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "price", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}, Transform: &extractor.Transform{
				Trim:  &trimBool,
				Regex: `\$(\d+\.\d+)`,
			}},
		},
	})
	if r.Data["price"] != "29.99" {
		t.Errorf("expected '29.99', got %q", r.Data["price"])
	}
}

func TestTransform_PrependAndAppend(t *testing.T) {
	trimBool := true
	r := extract(t, `<span>widget</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "sku", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}, Transform: &extractor.Transform{
				Trim:    &trimBool,
				Prepend: "SKU-",
				Append:  "-v2",
			}},
		},
	})
	if r.Data["sku"] != "SKU-widget-v2" {
		t.Errorf("expected 'SKU-widget-v2', got %q", r.Data["sku"])
	}
}

// --- Type coercion ---

func TestCoerce_StringToNumber(t *testing.T) {
	r := extract(t, `<span>42.5</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "count", Type: "number", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}},
		},
	})
	v, ok := r.Data["count"].(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", r.Data["count"])
	}
	if v != 42.5 {
		t.Errorf("expected 42.5, got %f", v)
	}
}

func TestCoerce_StringToBoolean(t *testing.T) {
	r := extract(t, `<span>true</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "available", Type: "boolean", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}},
		},
	})
	v, ok := r.Data["available"].(bool)
	if !ok {
		t.Fatalf("expected bool, got %T", r.Data["available"])
	}
	if !v {
		t.Error("expected true")
	}
}

// --- Required vs optional field handling ---

func TestRequiredField_MissingReturnsExtractionError(t *testing.T) {
	e := buildEngine(t, `<p>no heading here</p>`)
	tmpl := &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "title", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "h1"},
			}},
		},
	}
	_, err := e.Extract(tmpl)
	if err == nil {
		t.Fatal("expected ExtractionError for missing required field, got nil")
	}
	ee, ok := err.(*extractor.ExtractionError)
	if !ok {
		t.Fatalf("expected *ExtractionError, got %T", err)
	}
	if ee.FieldName != "title" {
		t.Errorf("expected field name 'title', got %q", ee.FieldName)
	}
}

func TestOptionalField_MissingGoesToFieldErrors(t *testing.T) {
	r := extract(t, `<p>no heading here</p>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "rating", Type: "string", Required: false, Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span.rating"},
			}},
		},
	})
	if _, present := r.Data["rating"]; present {
		t.Error("expected rating to be absent from Data")
	}
	if _, present := r.FieldErrors["rating"]; !present {
		t.Error("expected rating to appear in FieldErrors")
	}
}

func TestMultipleFields_MixedRequiredAndOptional(t *testing.T) {
	r := extract(t, `<h1>Title Here</h1>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "title", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "h1"},
			}},
			{Name: "price", Type: "number", Required: false, Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span.price"},
			}},
		},
	})
	if r.Data["title"] != "Title Here" {
		t.Errorf("expected 'Title Here', got %q", r.Data["title"])
	}
	if _, present := r.FieldErrors["price"]; !present {
		t.Error("expected price in FieldErrors")
	}
}
