package herald

import (
	"strings"

	"github.com/withoutasecondthought/herald/tokenizer"
)

const productsMapCap = 8

// tokenBrowsers get their engine from engineByChrome: these browsers are Blink
// on Android/desktop but WebKit in their iOS builds, which reuse the same token.
//
//nolint:gochecknoglobals // read-only lookup table
var tokenBrowsers = []struct{ token, name string }{
	{"YaBrowser", "Yandex Browser"},
	{"SamsungBrowser", "Samsung Browser"},
	{"UCBrowser", "UC Browser"},
	{"HuaweiBrowser", "Huawei Browser"},
	{"MiuiBrowser", "Mi Browser"},
	{"Whale", "Whale"},
	{"MQQBrowser", "QQ Browser"},
	{"OPX", "Opera GX"},
	{"DuckDuckGo", "DuckDuckGo"},
	{"GSA", "Google App"},
}

// detectBrowser identifies the browser from product tokens using priority matching.
// Order matters — more specific browsers must be checked before generic ones.
func detectBrowser(tokens []tokenizer.Token, result *Result) {
	products := buildProductMap(tokens)

	if matchFBBrowser(tokens, products, result) {
		return
	}

	if matchSpecificBrowsers(tokens, products, result) {
		return
	}

	if matchGenericBrowsers(products, result) {
		return
	}

	matchIE(tokens, result)
}

func buildProductMap(tokens []tokenizer.Token) map[string]string {
	products := make(map[string]string, productsMapCap)

	for _, tok := range tokens {
		if tok.Kind == tokenizer.KindProduct {
			products[tok.Name] = tok.Version
		}
	}

	return products
}

func matchFBBrowser(tokens []tokenizer.Token, products map[string]string, result *Result) bool {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindFBBlock {
			continue
		}

		fb := parseFBAttrs(tok.Attrs)

		app, ok := fb["FBAN"]
		if !ok {
			app, ok = fb["FB_IAB"]
		}

		if !ok || app == "" {
			continue
		}

		result.Browser = Browser{Name: fbAppName(app), Version: fb["FBAV"], Engine: engineByChrome(products)}

		return true
	}

	return false
}

func matchSpecificBrowsers(tokens []tokenizer.Token, products map[string]string, result *Result) bool {
	if v, ok := findAny(products, "Edg", "EdgA", "EdgiOS"); ok {
		result.Browser = Browser{Name: "Edge", Version: v, Engine: engineByChrome(products)}

		return true
	}

	if v, ok := findAny(products, "OPR", "OPRTouchPhone"); ok {
		result.Browser = Browser{Name: "Opera", Version: v, Engine: engineByChrome(products)}

		return true
	}

	if v, ok := products["OPiOS"]; ok {
		result.Browser = Browser{Name: "Opera Mini", Version: v, Engine: EngineWebKit}

		return true
	}

	if matchTokenBrowsers(products, result) {
		return true
	}

	if matchMiBrowser(products, result) {
		return true
	}

	if matchInstagram(products, result) {
		return true
	}

	if matchTikTokBrowser(tokens, products, result) {
		return true
	}

	if matchTwitterBrowser(tokens, products, result) {
		return true
	}

	return matchOperaMini(tokens, result)
}

func matchTokenBrowsers(products map[string]string, result *Result) bool {
	for _, e := range tokenBrowsers {
		if v, ok := products[e.token]; ok {
			result.Browser = Browser{Name: e.name, Version: v, Engine: engineByChrome(products)}

			return true
		}
	}

	return false
}

// matchMiBrowser handles the nested-slash token "XiaoMi/MiuiBrowser/14.44.1-gn",
// which tokenizes as product "XiaoMi" with version "MiuiBrowser/14.44.1-gn".
func matchMiBrowser(products map[string]string, result *Result) bool {
	v, ok := products["XiaoMi"]
	if !ok {
		return false
	}

	ver, ok := strings.CutPrefix(v, "MiuiBrowser/")
	if !ok {
		return false
	}

	result.Browser = Browser{Name: "Mi Browser", Version: ver, Engine: EngineBlink}

	return true
}

func matchTwitterBrowser(tokens []tokenizer.Token, products map[string]string, result *Result) bool {
	for i, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct {
			continue
		}

		if tok.Name == tokenTwitterAndroid {
			result.Browser = Browser{Name: appTwitter, Version: tok.Version, Engine: engineByChrome(products)}

			return true
		}

		if tok.Name == appTwitter {
			if dev, ok := twitterForTarget(tokens, i); ok {
				result.Browser = Browser{Name: appTwitter, Version: dev.Version, Engine: engineByChrome(products)}

				return true
			}
		}
	}

	return false
}

func matchInstagram(products map[string]string, result *Result) bool {
	for name, v := range products {
		if strings.HasPrefix(name, "Instagram") {
			result.Browser = Browser{Name: "Instagram", Version: v, Engine: engineByChrome(products)}

			return true
		}
	}

	return false
}

func matchTikTokBrowser(tokens []tokenizer.Token, products map[string]string, result *Result) bool {
	app, version, ok := tikTokApp(tokens)
	if !ok {
		return false
	}

	result.Browser = Browser{Name: app, Version: version, Engine: engineByChrome(products)}

	return true
}

// matchOperaMini handles Opera Mini extreme-savings mode on Android,
// where the identifier lives inside the comment: "(Android; Opera Mini/33.0.2254/174.101; U; en)".
func matchOperaMini(tokens []tokenizer.Token, result *Result) bool {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindComment {
			continue
		}

		for _, attr := range tok.Attrs {
			if v, ok := strings.CutPrefix(strings.TrimSpace(attr), "Opera Mini/"); ok {
				version, _, _ := strings.Cut(v, "/")

				result.Browser = Browser{Name: "Opera Mini", Version: version, Engine: EnginePresto}

				return true
			}
		}
	}

	return false
}

func matchGenericBrowsers(products map[string]string, result *Result) bool {
	if v, ok := products["CriOS"]; ok {
		result.Browser = Browser{Name: "Chrome", Version: v, Engine: EngineWebKit}

		return true
	}

	if v, ok := products["Chrome"]; ok {
		result.Browser = Browser{Name: "Chrome", Version: v, Engine: EngineBlink}

		return true
	}

	if v, ok := products["FxiOS"]; ok {
		result.Browser = Browser{Name: "Firefox", Version: v, Engine: EngineWebKit}

		return true
	}

	if v, ok := products["Firefox"]; ok {
		result.Browser = Browser{Name: "Firefox", Version: v, Engine: EngineGecko}

		return true
	}

	if v, ok := products["Version"]; ok {
		_, hasChrome := products["Chrome"]
		_, hasSafari := products["Safari"]

		if !hasChrome && hasSafari {
			result.Browser = Browser{Name: "Safari", Version: v, Engine: EngineWebKit}

			return true
		}
	}

	return false
}

func engineByChrome(products map[string]string) string {
	if _, ok := products["Chrome"]; ok {
		return EngineBlink
	}

	return EngineWebKit
}

func matchIE(tokens []tokenizer.Token, result *Result) {
	for _, tok := range tokens {
		if tok.Kind == tokenizer.KindProduct && tok.Name == "Trident" {
			result.Browser = Browser{
				Name:    "Internet Explorer",
				Version: tridentToIE(tok.Version),
				Engine:  EngineTrident,
			}

			return
		}

		if tok.Kind == tokenizer.KindComment {
			for _, attr := range tok.Attrs {
				if v, ok := strings.CutPrefix(attr, "MSIE "); ok {
					result.Browser = Browser{
						Name:    "Internet Explorer",
						Version: v,
						Engine:  EngineTrident,
					}

					return
				}
			}
		}
	}
}

func findAny(m map[string]string, keys ...string) (string, bool) {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v, true
		}
	}

	return "", false
}

func tridentToIE(tridentVersion string) string {
	switch tridentVersion {
	case "7.0":
		return "11"
	case "6.0":
		return "10"
	case "5.0":
		return "9"
	case "4.0":
		return "8"
	default:
		return ""
	}
}
