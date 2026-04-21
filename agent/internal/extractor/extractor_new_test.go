package extractor

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"
)

func TestNewFromReaders_GoqueryError(t *testing.T) {
	_, err := newFromReaders(iotest.ErrReader(errors.New("read error")), strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error from failing goquery reader, got nil")
	}
	if !strings.Contains(err.Error(), "goquery") {
		t.Errorf("expected error to mention goquery, got: %v", err)
	}
}

func TestNewFromReaders_HtmlqueryError(t *testing.T) {
	_, err := newFromReaders(strings.NewReader("<html></html>"), iotest.ErrReader(errors.New("read error")))
	if err == nil {
		t.Fatal("expected error from failing htmlquery reader, got nil")
	}
	if !strings.Contains(err.Error(), "htmlquery") {
		t.Errorf("expected error to mention htmlquery, got: %v", err)
	}
}
