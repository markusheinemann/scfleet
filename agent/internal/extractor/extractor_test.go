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

func TestExtractionError_ErrorMessage(t *testing.T) {
	err := &extractor.ExtractionError{FieldName: "price"}
	want := `required field "price": no extractor yielded a value`
	if err.Error() != want {
		t.Errorf("expected %q, got %q", want, err.Error())
	}
}

func TestRequiredField_CoerceError_ReturnsExtractionError(t *testing.T) {
	e := buildEngine(t, `<span>not-a-number</span>`)
	tmpl := &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "price", Type: "number", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}},
		},
	}
	_, err := e.Extract(tmpl)
	if err == nil {
		t.Fatal("expected ExtractionError for required field with coerce failure")
	}
	ee, ok := err.(*extractor.ExtractionError)
	if !ok {
		t.Fatalf("expected *ExtractionError, got %T", err)
	}
	if ee.FieldName != "price" {
		t.Errorf("expected field name 'price', got %q", ee.FieldName)
	}
}

func TestUnknownStrategy_TreatedAsMissing(t *testing.T) {
	r := extract(t, `<p>text</p>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "val", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "unknown_strategy"},
			}},
		},
	})
	if _, ok := r.Data["val"]; ok {
		t.Error("expected val absent for unknown strategy")
	}
	if _, ok := r.FieldErrors["val"]; !ok {
		t.Error("expected val in FieldErrors for unknown strategy")
	}
}

func TestEmptyFields_ReturnsEmptyResult(t *testing.T) {
	r := extract(t, `<h1>Title</h1>`, &extractor.Template{
		Version: "1",
		Fields:  []extractor.Field{},
	})
	if len(r.Data) != 0 {
		t.Errorf("expected empty Data, got %v", r.Data)
	}
	if len(r.FieldErrors) != 0 {
		t.Errorf("expected empty FieldErrors, got %v", r.FieldErrors)
	}
}

func TestRequiredField_TransformEmptyResult_ReturnsExtractionError(t *testing.T) {
	trimTrue := true
	e := buildEngine(t, `<span>hello world</span>`)
	tmpl := &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "num", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}, Transform: &extractor.Transform{
				Trim:  &trimTrue,
				Regex: `\d+`, // no digits → no match → empty string
			}},
		},
	}
	_, err := e.Extract(tmpl)
	if err == nil {
		t.Fatal("expected ExtractionError when required field transform returns empty")
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

// --- XPath edge cases ---

func TestXPath_InvalidSelector_TreatedAsMissing(t *testing.T) {
	r := extract(t, `<p>text</p>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "val", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "xpath", Selector: "///[invalid"},
			}},
		},
	})
	if _, ok := r.Data["val"]; ok {
		t.Error("expected val absent for invalid XPath selector")
	}
}

func TestXPath_NoMatch_TreatedAsMissing(t *testing.T) {
	r := extract(t, `<p>text</p>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "val", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "xpath", Selector: "//h1"},
			}},
		},
	})
	if _, ok := r.Data["val"]; ok {
		t.Error("expected val absent when XPath has no match")
	}
}

func TestXPath_MultipleValues(t *testing.T) {
	r := extract(t, `<ul><li>A</li><li>B</li><li>C</li></ul>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "items", Type: "array", ItemType: "string", Extractors: []extractor.Extractor{
				{Strategy: "xpath", Selector: "//li", Multiple: true},
			}},
		},
	})
	items, ok := r.Data["items"].([]any)
	if !ok || len(items) != 3 {
		t.Fatalf("expected 3 items, got %v", r.Data["items"])
	}
	if items[0] != "A" || items[2] != "C" {
		t.Errorf("unexpected values: %v", items)
	}
}

func TestXPath_MultipleValues_FiltersEmpty(t *testing.T) {
	r := extract(t, `<ul><li>A</li><li></li><li>B</li></ul>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "items", Type: "array", ItemType: "string", Extractors: []extractor.Extractor{
				{Strategy: "xpath", Selector: "//li", Multiple: true},
			}},
		},
	})
	items, ok := r.Data["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected 2 items (empty filtered), got %v", r.Data["items"])
	}
}

// --- CSS edge cases ---

func TestCSS_NoMatch_TreatedAsMissing(t *testing.T) {
	r := extract(t, `<p>text</p>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "val", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "h1"},
			}},
		},
	})
	if _, ok := r.Data["val"]; ok {
		t.Error("expected val absent when CSS selector has no match")
	}
}

func TestCSS_MultipleMode_AllEmpty_TreatedAsMissing(t *testing.T) {
	r := extract(t, `<ul><li></li><li></li></ul>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "items", Type: "array", ItemType: "string", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "li", Multiple: true},
			}},
		},
	})
	if _, ok := r.Data["items"]; ok {
		t.Error("expected items absent when all matched elements are empty")
	}
}

func TestCSS_HtmlAttribute(t *testing.T) {
	r := extract(t, `<div><span>inner</span></div>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "inner", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "div", Attribute: "html"},
			}},
		},
	})
	if r.Data["inner"] != "<span>inner</span>" {
		t.Errorf("expected '<span>inner</span>', got %q", r.Data["inner"])
	}
}

// --- Meta edge cases ---

func TestMeta_CaseInsensitiveProperty(t *testing.T) {
	r := extract(t, `<meta property="OG:TITLE" content="Upper Case">`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "title", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "meta", Property: "og:title"},
			}},
		},
	})
	if r.Data["title"] != "Upper Case" {
		t.Errorf("expected 'Upper Case', got %q", r.Data["title"])
	}
}

func TestMeta_MultipleTagsFindsSecondByName(t *testing.T) {
	// First meta tag has a different name, forcing the loop to continue (return true) before matching.
	html := `<meta name="author" content="Someone"><meta name="description" content="Page desc">`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "desc", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "meta", Property: "description"},
			}},
		},
	})
	if r.Data["desc"] != "Page desc" {
		t.Errorf("expected 'Page desc', got %q", r.Data["desc"])
	}
}

func TestMeta_NotFound_TreatedAsMissing(t *testing.T) {
	r := extract(t, `<p>no meta</p>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "title", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "meta", Property: "og:title"},
			}},
		},
	})
	if _, ok := r.Data["title"]; ok {
		t.Error("expected title absent when meta not found")
	}
}

// --- Microdata edge cases ---

func TestMicrodata_LinkTag_ExtractsHref(t *testing.T) {
	r := extract(t, `<link itemprop="url" href="https://example.com/p">`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "url", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "microdata", Itemprop: "url"},
			}},
		},
	})
	if r.Data["url"] != "https://example.com/p" {
		t.Errorf("expected 'https://example.com/p', got %q", r.Data["url"])
	}
}

func TestMicrodata_CustomAttribute(t *testing.T) {
	r := extract(t, `<div itemprop="rating" data-value="4.5">4.5 stars</div>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "rating", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "microdata", Itemprop: "rating", Attribute: "data-value"},
			}},
		},
	})
	if r.Data["rating"] != "4.5" {
		t.Errorf("expected '4.5', got %q", r.Data["rating"])
	}
}

func TestMicrodata_MultipleItems(t *testing.T) {
	r := extract(t, `<span itemprop="tag">Go</span><span itemprop="tag">Testing</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "tags", Type: "array", ItemType: "string", Extractors: []extractor.Extractor{
				{Strategy: "microdata", Itemprop: "tag", Multiple: true},
			}},
		},
	})
	items, ok := r.Data["tags"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected 2 items, got %v", r.Data["tags"])
	}
}

func TestMicrodata_NotFound_TreatedAsMissing(t *testing.T) {
	r := extract(t, `<p>no microdata</p>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "val", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "microdata", Itemprop: "price"},
			}},
		},
	})
	if _, ok := r.Data["val"]; ok {
		t.Error("expected val absent when microdata not found")
	}
}

// --- JSON-LD edge cases ---

func TestJSONLD_BlockIndexOutOfBounds_TreatedAsMissing(t *testing.T) {
	html := `<script type="application/ld+json">{"name":"test"}</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "val", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "json_ld", BlockIndex: 5, Path: "name"},
			}},
		},
	})
	if _, ok := r.Data["val"]; ok {
		t.Error("expected val absent when block index out of bounds")
	}
}

func TestJSONLD_NoBlocks_TreatedAsMissing(t *testing.T) {
	r := extract(t, `<p>no json-ld here</p>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "val", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "json_ld", Path: "name"},
			}},
		},
	})
	if _, ok := r.Data["val"]; ok {
		t.Error("expected val absent when no JSON-LD blocks present")
	}
}

func TestJSONLD_CachingReturnsSameBlocks(t *testing.T) {
	// Two fields both reading from JSON-LD blocks exercising the cached path.
	html := `<script type="application/ld+json">{"a":"1"}</script>` +
		`<script type="application/ld+json">{"b":"2"}</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "a", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "json_ld", BlockIndex: 0, Path: "a"},
			}},
			{Name: "b", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "json_ld", BlockIndex: 1, Path: "b"},
			}},
		},
	})
	if r.Data["a"] != "1" {
		t.Errorf("expected '1', got %q", r.Data["a"])
	}
	if r.Data["b"] != "2" {
		t.Errorf("expected '2', got %q", r.Data["b"])
	}
}

func TestSchemaOrg_GraphExtraction(t *testing.T) {
	ldJSON := `{"@context":"https://schema.org","@graph":[{"@type":"Product","name":"Graph Widget"},{"@type":"BreadcrumbList"}]}`
	html := `<script type="application/ld+json">` + ldJSON + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "name", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "schema_org", SchemaType: "Product", Path: "name"},
			}},
		},
	})
	if r.Data["name"] != "Graph Widget" {
		t.Errorf("expected 'Graph Widget', got %q", r.Data["name"])
	}
}

func TestSchemaOrg_TypeAsArray(t *testing.T) {
	ldJSON := `{"@context":"https://schema.org","@type":["Product","Thing"],"name":"Multi-type Widget"}`
	html := `<script type="application/ld+json">` + ldJSON + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "name", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "schema_org", SchemaType: "Product", Path: "name"},
			}},
		},
	})
	if r.Data["name"] != "Multi-type Widget" {
		t.Errorf("expected 'Multi-type Widget', got %q", r.Data["name"])
	}
}

func TestSchemaOrg_ArrayBlockNoMatch_FallsThroughToNextBlock(t *testing.T) {
	// First block is a top-level array with no matching type; second block has the match.
	block1 := `[{"@type":"BreadcrumbList"},{"@type":"WebPage"}]`
	block2 := `{"@type":"Product","name":"Found in second block"}`
	html := `<script type="application/ld+json">` + block1 + `</script>` +
		`<script type="application/ld+json">` + block2 + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "name", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "schema_org", SchemaType: "Product", Path: "name"},
			}},
		},
	})
	if r.Data["name"] != "Found in second block" {
		t.Errorf("expected 'Found in second block', got %q", r.Data["name"])
	}
}

func TestSchemaOrg_GraphBlockNoMatch_FallsThroughToNextBlock(t *testing.T) {
	// First block has a @graph but no matching type; second block has the match.
	block1 := `{"@graph":[{"@type":"BreadcrumbList"},{"@type":"WebPage"}]}`
	block2 := `{"@type":"Product","name":"Found after graph"}`
	html := `<script type="application/ld+json">` + block1 + `</script>` +
		`<script type="application/ld+json">` + block2 + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "name", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "schema_org", SchemaType: "Product", Path: "name"},
			}},
		},
	})
	if r.Data["name"] != "Found after graph" {
		t.Errorf("expected 'Found after graph', got %q", r.Data["name"])
	}
}

func TestSchemaOrg_NumericType_NotMatched(t *testing.T) {
	// @type is a JSON number — not a string or array, so matchesSchemaType returns false.
	ldJSON := `{"@type":42,"name":"Numeric type"}`
	html := `<script type="application/ld+json">` + ldJSON + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "name", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "schema_org", SchemaType: "Product", Path: "name"},
			}},
		},
	})
	if _, ok := r.Data["name"]; ok {
		t.Error("expected name absent when @type is a JSON number")
	}
}

func TestSchemaOrg_MissingType_NotMatched(t *testing.T) {
	ldJSON := `{"@context":"https://schema.org","name":"No Type Object"}`
	html := `<script type="application/ld+json">` + ldJSON + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "name", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "schema_org", SchemaType: "Product", Path: "name"},
			}},
		},
	})
	if _, ok := r.Data["name"]; ok {
		t.Error("expected name absent when @type is missing and non-wildcard SchemaType given")
	}
}

// --- Transform edge cases ---

func TestTransform_TrimDisabled_PreservesWhitespace(t *testing.T) {
	trimFalse := false
	ldJSON := `{"text":"  padded  "}`
	html := `<script type="application/ld+json">` + ldJSON + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "text", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "json_ld", Path: "text"},
			}, Transform: &extractor.Transform{Trim: &trimFalse}},
		},
	})
	if r.Data["text"] != "  padded  " {
		t.Errorf("expected '  padded  ' with trim disabled, got %q", r.Data["text"])
	}
}

func TestTransform_InvalidRegex_ReturnsEmpty(t *testing.T) {
	trimTrue := true
	r := extract(t, `<span>some text</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "val", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}, Transform: &extractor.Transform{Trim: &trimTrue, Regex: `[invalid`}},
		},
	})
	// applyTransform returns "" on compile error; coerce("","string") stores "" in Data
	if r.Data["val"] != "" {
		t.Errorf("expected empty string when regex is invalid, got %q", r.Data["val"])
	}
}

func TestTransform_RegexNoMatch_ReturnsEmpty(t *testing.T) {
	trimTrue := true
	r := extract(t, `<span>hello world</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "val", Type: "string", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}, Transform: &extractor.Transform{Trim: &trimTrue, Regex: `\d+`}},
		},
	})
	// applyTransform returns "" when regex has no match; coerce stores "" in Data
	if r.Data["val"] != "" {
		t.Errorf("expected empty string when regex has no match, got %q", r.Data["val"])
	}
}

func TestTransform_RegexNoCaptureGroup_ReturnsFullMatch(t *testing.T) {
	trimTrue := true
	r := extract(t, `<span>abc 123 def</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "num", Type: "string", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}, Transform: &extractor.Transform{Trim: &trimTrue, Regex: `\d+`}},
		},
	})
	if r.Data["num"] != "123" {
		t.Errorf("expected '123', got %q", r.Data["num"])
	}
}

// --- Coerce edge cases ---

func TestCoerce_NumberWithCommas(t *testing.T) {
	r := extract(t, `<span>1,234.56</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "price", Type: "number", Required: true, Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}},
		},
	})
	v, ok := r.Data["price"].(float64)
	if !ok {
		t.Fatalf("expected float64, got %T", r.Data["price"])
	}
	if v != 1234.56 {
		t.Errorf("expected 1234.56, got %f", v)
	}
}

func TestCoerce_InvalidNumber_GoesToFieldErrors(t *testing.T) {
	r := extract(t, `<span>not-a-number</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "count", Type: "number", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}},
		},
	})
	if _, ok := r.Data["count"]; ok {
		t.Error("expected count absent for invalid number")
	}
	if _, ok := r.FieldErrors["count"]; !ok {
		t.Error("expected count in FieldErrors for invalid number")
	}
}

func TestCoerce_UnknownType_GoesToFieldErrors(t *testing.T) {
	r := extract(t, `<span>42</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "val", Type: "unknown_type", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}},
		},
	})
	if _, ok := r.Data["val"]; ok {
		t.Error("expected val absent for unknown type")
	}
	if _, ok := r.FieldErrors["val"]; !ok {
		t.Error("expected val in FieldErrors for unknown type")
	}
}

func TestCoerce_BooleanTrueVariants(t *testing.T) {
	for _, input := range []string{"true", "1", "yes", "TRUE", "Yes"} {
		r := extract(t, `<span>`+input+`</span>`, &extractor.Template{
			Version: "1",
			Fields: []extractor.Field{
				{Name: "v", Type: "boolean", Required: true, Extractors: []extractor.Extractor{
					{Strategy: "css", Selector: "span"},
				}},
			},
		})
		if v, ok := r.Data["v"].(bool); !ok || !v {
			t.Errorf("input %q: expected true bool, got %v (%T)", input, r.Data["v"], r.Data["v"])
		}
	}
}

func TestCoerce_BooleanFalseVariants(t *testing.T) {
	for _, input := range []string{"false", "0", "no", "FALSE", "No"} {
		r := extract(t, `<span>`+input+`</span>`, &extractor.Template{
			Version: "1",
			Fields: []extractor.Field{
				{Name: "v", Type: "boolean", Required: true, Extractors: []extractor.Extractor{
					{Strategy: "css", Selector: "span"},
				}},
			},
		})
		if v, ok := r.Data["v"].(bool); !ok || v {
			t.Errorf("input %q: expected false bool, got %v (%T)", input, r.Data["v"], r.Data["v"])
		}
	}
}

func TestCoerce_InvalidBoolean_GoesToFieldErrors(t *testing.T) {
	r := extract(t, `<span>maybe</span>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "flag", Type: "boolean", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "span"},
			}},
		},
	})
	if _, ok := r.Data["flag"]; ok {
		t.Error("expected flag absent for invalid boolean")
	}
	if _, ok := r.FieldErrors["flag"]; !ok {
		t.Error("expected flag in FieldErrors for invalid boolean")
	}
}

func TestCoerce_ArrayWithNumberItems(t *testing.T) {
	r := extract(t, `<ul><li>1.5</li><li>2.0</li><li>3.7</li></ul>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "prices", Type: "array", ItemType: "number", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "li", Multiple: true},
			}},
		},
	})
	items, ok := r.Data["prices"].([]any)
	if !ok || len(items) != 3 {
		t.Fatalf("expected 3 items, got %v", r.Data["prices"])
	}
	if items[0] != 1.5 {
		t.Errorf("expected 1.5, got %v", items[0])
	}
}

func TestCoerce_ArrayType_NotJSONArray_GoesToFieldErrors(t *testing.T) {
	// json_ld always returns a plain string (ignores multiple mode).
	// coerce("plain text", "array", "string") → json.Unmarshal error.
	ldJSON := `{"tags":"not-a-json-array"}`
	html := `<script type="application/ld+json">` + ldJSON + `</script>`
	r := extract(t, html, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "items", Type: "array", ItemType: "string", Extractors: []extractor.Extractor{
				{Strategy: "json_ld", Path: "tags"},
			}},
		},
	})
	if _, ok := r.Data["items"]; ok {
		t.Error("expected items absent when raw value is not a JSON array")
	}
	if _, ok := r.FieldErrors["items"]; !ok {
		t.Error("expected items in FieldErrors when raw value is not a JSON array")
	}
}

func TestCoerce_ArrayWithInvalidItems_SkipsBadItems(t *testing.T) {
	r := extract(t, `<ul><li>1.0</li><li>not-a-num</li><li>3.0</li></ul>`, &extractor.Template{
		Version: "1",
		Fields: []extractor.Field{
			{Name: "nums", Type: "array", ItemType: "number", Extractors: []extractor.Extractor{
				{Strategy: "css", Selector: "li", Multiple: true},
			}},
		},
	})
	items, ok := r.Data["nums"].([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", r.Data["nums"])
	}
	if len(items) != 2 {
		t.Errorf("expected 2 valid items (bad one skipped), got %d", len(items))
	}
}
