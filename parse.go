package herald

import (
	"strings"

	"github.com/withoutasecondthought/herald/db"
	"github.com/withoutasecondthought/herald/tokenizer"
)

//nolint:gochecknoglobals // read-only lookup table for HTTP client signatures
var httpClientSignatures = []string{
	"dart/", "curl/", "wget/", "python-requests/", "python-urllib/",
	"go-http-client/", "java/", "libwww-perl/", "httpie/",
	"node-fetch/", "axios/", "embarcadero uri client/", "ees update",
	"winhttprequest", "aiohttp/", "go-resty/", "python-httpx/",
	"httpx/", "siege/", "req/",
}

// Parse parses a User-Agent string using the Parser's in-memory database.
func (p *Parser) Parse(ua string) *Result {
	return p.run(ua, nil)
}

// ParseWithHints parses a UA string and enriches the result with Client Hints.
func (p *Parser) ParseWithHints(ua string, hints ClientHints) *Result {
	return p.run(ua, &hints)
}

// DetectType is a fast path that only determines the client type
// without full parsing of browser/OS/device details.
func (p *Parser) DetectType(ua string) ClientType {
	if len(ua) == 0 {
		return ClientTypeUnknown
	}

	if _, found := p.db.BotTrie.Search(ua); found {
		return ClientTypeBot
	}

	tokens := tokenizer.Tokenize(ua)

	return classifyClient(tokens, ua)
}

// run executes the full parsing pipeline.
func (p *Parser) run(raw string, hints *ClientHints) *Result {
	result := &Result{Raw: raw}

	if len(raw) == 0 {
		result.ClientType = ClientTypeUnknown

		return result
	}

	tokens := tokenizer.Tokenize(raw)

	if detectBot(tokens, result, p.db) {
		return result
	}

	clientType := classifyClient(tokens, raw)

	result.ClientType = clientType

	p.runStages(clientType, tokens, result)

	if hints != nil {
		applyHints(hints, result)
	}

	return result
}

func (p *Parser) runStages(clientType ClientType, tokens []tokenizer.Token, result *Result) {
	switch clientType {
	case ClientTypeBrowser, ClientTypeIAB:
		detectBrowser(tokens, result)
		detectOS(tokens, result, p.db)
		detectDevice(tokens, result, p.db)

		if detectIAB(tokens, result, p.db) {
			result.ClientType = ClientTypeIAB
		}

	case ClientTypeNativeApp:
		detectNative(tokens, result)
		detectOS(tokens, result, p.db)
		detectDevice(tokens, result, p.db)

	case ClientTypeHttpClient:
		detectNative(tokens, result)

	default:
		result.ClientType = ClientTypeUnknown
	}
}

// classifyClient determines the client type from tokens before detailed parsing.
func classifyClient(tokens []tokenizer.Token, raw string) ClientType {
	if ct, ok := classifyByTokens(tokens); ok {
		return ct
	}

	if isHTTPClientUA(raw) {
		return ClientTypeHttpClient
	}

	if hasMozillaPrefix(tokens) {
		return ClientTypeBrowser
	}

	if len(tokens) == 1 && tokens[0].Kind == tokenizer.KindProduct {
		return ClientTypeHttpClient
	}

	if len(tokens) == 0 {
		return ClientTypeUnknown
	}

	return ClientTypeBrowser
}

func classifyByTokens(tokens []tokenizer.Token) (ClientType, bool) {
	for _, tok := range tokens {
		if tok.Kind == tokenizer.KindFBBlock {
			return ClientTypeIAB, true
		}

		if tok.Kind == tokenizer.KindProduct {
			switch tok.Name {
			case "FBAV", "Instagram", "musical_ly", "Barcelona", "MetaIAB":
				return ClientTypeIAB, true
			case ProductCFNetwork, "okhttp", "Dalvik":
				return ClientTypeNativeApp, true
			}
		}
	}

	return 0, false
}

func isHTTPClientUA(raw string) bool {
	rawLower := strings.ToLower(raw)

	for _, sig := range httpClientSignatures {
		if strings.Contains(rawLower, sig) {
			return true
		}
	}

	return false
}

func hasMozillaPrefix(tokens []tokenizer.Token) bool {
	for _, tok := range tokens {
		if tok.Kind == tokenizer.KindProduct && tok.Name == ProductMozilla {
			return true
		}
	}

	return false
}

// detectNative extracts native app information from tokens.
func detectNative(tokens []tokenizer.Token, result *Result) {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct {
			continue
		}

		if matchNativeProduct(tok, tokens, result) {
			return
		}
	}

	// Fallback: first product token
	for _, tok := range tokens {
		if tok.Kind == tokenizer.KindProduct {
			result.Native = NativeApp{Name: tok.Name, Version: tok.Version}

			return
		}
	}
}

func matchNativeProduct(tok tokenizer.Token, allTokens []tokenizer.Token, result *Result) bool {
	switch tok.Name {
	case ProductCFNetwork:
		result.Native = NativeApp{Runtime: ProductCFNetwork, Version: tok.Version}
		fillCFNetworkAppName(allTokens, result)

		return true
	case "okhttp":
		result.Native = NativeApp{Name: "OkHttp", Version: tok.Version, Runtime: "OkHttp"}

		return true
	case "Dart":
		runtime := findDartRuntime(allTokens)

		result.Native = NativeApp{Name: "Dart", Version: tok.Version, Runtime: runtime}

		return true
	case "curl":
		result.Native = NativeApp{Name: "curl", Version: tok.Version, Runtime: "curl"}

		return true
	case "Wget":
		result.Native = NativeApp{Name: "Wget", Version: tok.Version, Runtime: "Wget"}

		return true
	case "python-requests":
		result.Native = NativeApp{Name: "python-requests", Version: tok.Version, Runtime: "Python"}

		return true
	default:
		return false
	}
}

func fillCFNetworkAppName(tokens []tokenizer.Token, result *Result) {
	for _, t := range tokens {
		if t.Kind == tokenizer.KindProduct && t.Name != ProductCFNetwork && t.Name != ProductDarwin {
			result.Native.Name = t.Name
			result.Native.Version = t.Version

			break
		}
	}
}

func findDartRuntime(tokens []tokenizer.Token) string {
	for _, t := range tokens {
		if t.Kind == tokenizer.KindComment {
			for _, attr := range t.Attrs {
				if strings.TrimSpace(attr) == "dart:io" {
					return "dart:io"
				}
			}
		}
	}

	return ""
}

// Package-level convenience functions using the default parser.

// Parse parses a UA string using the default parser. Init must be called first.
func Parse(ua string) *Result {
	p := getDefault()
	if p == nil {
		return &Result{Raw: ua, ClientType: ClientTypeUnknown}
	}

	return p.Parse(ua)
}

// ParseWithHints parses a UA string with Client Hints. Init must be called first.
func ParseWithHints(ua string, hints ClientHints) *Result {
	p := getDefault()
	if p == nil {
		return &Result{Raw: ua, ClientType: ClientTypeUnknown}
	}

	return p.ParseWithHints(ua, hints)
}

// DetectType detects the client type. Init must be called first.
func DetectType(ua string) ClientType {
	p := getDefault()
	if p == nil {
		return ClientTypeUnknown
	}

	return p.DetectType(ua)
}

// LookupAppleModel looks up an Apple model identifier (e.g., "iPhone14,2") → human name.
func (p *Parser) LookupAppleModel(modelID string) (string, bool) {
	name, ok := p.db.AppleModels[modelID]

	return name, ok
}

// LookupAndroidModel looks up an Android model (e.g., "SM-G991B") → brand + model.
func (p *Parser) LookupAndroidModel(modelID string) (db.AndroidDevice, bool) {
	dev, ok := p.db.AndroidDB[modelID]

	return dev, ok
}
