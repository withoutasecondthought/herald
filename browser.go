package herald

import (
	"strings"

	"github.com/withoutasecondthought/herald/tokenizer"
)

const productsMapCap = 8

// detectBrowser identifies the browser from product tokens using priority matching.
// Order matters — more specific browsers must be checked before generic ones.
func detectBrowser(tokens []tokenizer.Token, result *Result) {
	products := buildProductMap(tokens)

	if matchSpecificBrowsers(products, result) {
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

func matchSpecificBrowsers(products map[string]string, result *Result) bool {
	if v, ok := findAny(products, "Edg", "EdgA", "EdgIOS"); ok {
		result.Browser = Browser{Name: "Edge", Version: v, Engine: EngineBlink}

		return true
	}

	if v, ok := findAny(products, "OPR", "OPRTouchPhone"); ok {
		result.Browser = Browser{Name: "Opera", Version: v, Engine: EngineBlink}

		return true
	}

	if v, ok := products["YaBrowser"]; ok {
		result.Browser = Browser{Name: "Yandex Browser", Version: v, Engine: EngineBlink}

		return true
	}

	if v, ok := products["SamsungBrowser"]; ok {
		result.Browser = Browser{Name: "Samsung Browser", Version: v, Engine: EngineBlink}

		return true
	}

	if v, ok := products["FBAV"]; ok {
		result.Browser = Browser{Name: "Facebook", Version: v, Engine: EngineWebKit}

		return true
	}

	for name, v := range products {
		if strings.HasPrefix(name, "Instagram") {
			engine := EngineWebKit
			if _, hasChrome := products["Chrome"]; hasChrome {
				engine = EngineBlink
			}

			result.Browser = Browser{Name: "Instagram", Version: v, Engine: engine}

			return true
		}
	}

	if v, ok := products["musical_ly"]; ok {
		result.Browser = Browser{Name: "TikTok", Version: v, Engine: EngineWebKit}

		return true
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
