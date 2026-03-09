package tokenizer_test

import (
	"testing"

	"github.com/withoutasecondthought/herald/tokenizer"
)

func TestTokenize_Empty(t *testing.T) {
	t.Parallel()

	tokens := tokenizer.Tokenize("")
	if len(tokens) != 0 {
		t.Errorf("expected 0 tokens, got %d", len(tokens))
	}
}

func TestTokenize_SimpleProduct(t *testing.T) {
	t.Parallel()

	tokens := tokenizer.Tokenize("curl/7.68.0")
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}

	tok := tokens[0]

	if tok.Kind != tokenizer.KindProduct {
		t.Errorf("expected KindProduct, got %d", tok.Kind)
	}

	if tok.Name != "curl" {
		t.Errorf("expected name 'curl', got %q", tok.Name)
	}

	if tok.Version != "7.68.0" {
		t.Errorf("expected version '7.68.0', got %q", tok.Version)
	}
}

func TestTokenize_ProductWithoutVersion(t *testing.T) {
	t.Parallel()

	tokens := tokenizer.Tokenize("Safari")
	if len(tokens) != 1 {
		t.Fatalf("expected 1 token, got %d", len(tokens))
	}

	if tokens[0].Name != "Safari" {
		t.Errorf("expected name 'Safari', got %q", tokens[0].Name)
	}

	if tokens[0].Version != "" {
		t.Errorf("expected empty version, got %q", tokens[0].Version)
	}
}

func TestTokenize_ChromeDesktop(t *testing.T) {
	t.Parallel()

	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/120.0.0.0 Safari/537.36"
	tokens := tokenizer.Tokenize(ua)

	minExpected := 4
	if len(tokens) < minExpected {
		t.Fatalf("expected at least %d tokens, got %d", minExpected, len(tokens))
	}

	if tokens[0].Kind != tokenizer.KindProduct ||
		tokens[0].Name != "Mozilla" ||
		tokens[0].Version != "5.0" {
		t.Errorf("first token: got %+v", tokens[0])
	}

	if tokens[1].Kind != tokenizer.KindComment {
		t.Errorf("second token should be comment, got kind=%d", tokens[1].Kind)
	}

	if len(tokens[1].Attrs) < 1 || tokens[1].Attrs[0] != "Windows NT 10.0" {
		t.Errorf("expected 'Windows NT 10.0' in attrs, got %v", tokens[1].Attrs)
	}
}

func TestTokenize_NestedComments(t *testing.T) {
	t.Parallel()

	ua := "Mozilla/5.0 (KHTML, like Gecko (compatible))"
	tokens := tokenizer.Tokenize(ua)

	commentFound := false

	for _, tok := range tokens {
		if tok.Kind == tokenizer.KindComment {
			commentFound = true

			if tok.Raw != "(KHTML, like Gecko (compatible))" {
				t.Errorf("expected full nested comment, got %q", tok.Raw)
			}
		}
	}

	if !commentFound {
		t.Error("no comment token found")
	}
}

func TestTokenize_FBBlock(t *testing.T) {
	t.Parallel()

	ua := "Mozilla/5.0 [FBAN/FBIOS;FBAV/450.0.0;FBDV/iPhone15,4]"
	tokens := tokenizer.Tokenize(ua)

	fbFound := false

	for _, tok := range tokens {
		if tok.Kind == tokenizer.KindFBBlock {
			fbFound = true

			minAttrs := 3
			if len(tok.Attrs) < minAttrs {
				t.Errorf("expected at least %d FB attrs, got %d: %v",
					minAttrs, len(tok.Attrs), tok.Attrs)
			}
		}
	}

	if !fbFound {
		t.Error("no FB block token found")
	}
}

func TestTokenize_CFNetwork(t *testing.T) {
	t.Parallel()

	ua := "ut-1/3 CFNetwork/3860.400.51 Darwin/25.3.0"
	tokens := tokenizer.Tokenize(ua)

	expectedCount := 3
	if len(tokens) != expectedCount {
		t.Fatalf("expected %d tokens, got %d", expectedCount, len(tokens))
	}

	if tokens[0].Name != "ut-1" || tokens[0].Version != "3" {
		t.Errorf("first token: %+v", tokens[0])
	}

	if tokens[1].Name != "CFNetwork" || tokens[1].Version != "3860.400.51" {
		t.Errorf("second token: %+v", tokens[1])
	}

	if tokens[2].Name != "Darwin" || tokens[2].Version != "25.3.0" {
		t.Errorf("third token: %+v", tokens[2])
	}
}

func TestTokenize_Dart(t *testing.T) {
	t.Parallel()

	ua := "Dart/3.10 (dart:io)"
	tokens := tokenizer.Tokenize(ua)

	expectedCount := 2
	if len(tokens) != expectedCount {
		t.Fatalf("expected %d tokens, got %d", expectedCount, len(tokens))
	}

	if tokens[0].Name != "Dart" || tokens[0].Version != "3.10" {
		t.Errorf("product: %+v", tokens[0])
	}

	if tokens[1].Kind != tokenizer.KindComment {
		t.Errorf("expected comment, got kind=%d", tokens[1].Kind)
	}

	if len(tokens[1].Attrs) != 1 || tokens[1].Attrs[0] != "dart:io" {
		t.Errorf("comment attrs: %v", tokens[1].Attrs)
	}
}
