package herald

import (
	"strconv"
	"strings"

	"github.com/withoutasecondthought/herald/db"
	"github.com/withoutasecondthought/herald/tokenizer"
)

// Instagram positional attribute indices.
// Instagram/Barcelona iOS positional attribute indices.
const (
	igAttrModel      = 0
	igAttrOS         = 1
	igAttrLocale     = 2
	igAttrScale      = 4
	igAttrResolution = 5

	igMinModel      = 1
	igMinOS         = 2
	igMinLocale     = 3
	igMinScale      = 5
	igMinResolution = 6

	fbMapCap     = 8
	scalePrefix  = "scale="
	scalePrefLen = 6
)

// Instagram/Barcelona Android positional attribute indices.
// Format: (apiLevel/ver; dpi; WxH; brand; model; codename; chipset; locale; buildId)
const (
	igAndroidAttrResolution = 2
	igAndroidAttrBrand      = 3
	igAndroidAttrModel      = 4
	igAndroidAttrLocale     = 7

	igAndroidMinResolution = 3
	igAndroidMinModel      = 5
	igAndroidMinLocale     = 8
)

// detectIAB parses in-app browser info from Facebook, Instagram, Barcelona (Threads), and TikTok UAs.
// Returns true if IAB was detected.
func detectIAB(tokens []tokenizer.Token, result *Result, database *db.Database) bool {
	if parseFacebookIAB(tokens, result, database) {
		return true
	}

	if parseInstagramIAB(tokens, result, database) {
		return true
	}

	if parseBarcelonaIAB(tokens, result, database) {
		return true
	}

	if parseTikTokIAB(tokens, result) {
		return true
	}

	return parseMetaIABGeneric(tokens, result)
}

func parseFacebookIAB(tokens []tokenizer.Token, result *Result, database *db.Database) bool {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindFBBlock {
			continue
		}

		fb := parseFBAttrs(tok.Attrs)

		if _, ok := fb["FBAN"]; !ok {
			continue
		}

		result.IAB = IABInfo{
			App:        fb["FBAN"],
			AppVersion: fb["FBAV"],
			Locale:     fb["FBLC"],
		}

		scale, err := strconv.ParseFloat(fb["FBSS"], 64)
		if err == nil {
			result.IAB.ScreenScale = scale
		}

		applyFBOverrides(fb, result, database)

		return true
	}

	return false
}

func parseFBAttrs(attrs []string) map[string]string {
	fb := make(map[string]string, fbMapCap)

	for _, attr := range attrs {
		if key, val, ok := strings.Cut(attr, "/"); ok {
			fb[key] = val
		}
	}

	return fb
}

func applyFBOverrides(fb map[string]string, result *Result, database *db.Database) {
	if dev := fb["FBDV"]; dev != "" {
		result.Device.ModelRaw = dev

		if name, ok := database.AppleModels[dev]; ok {
			result.Device.Model = name
		}
	}

	if osName := fb["FBSN"]; osName != "" {
		result.OS.Name = osName
	}

	if osVer := fb["FBSV"]; osVer != "" {
		result.OS.Version = osVer
	}

	if id := fb["FBID"]; id != "" {
		switch strings.ToLower(id) {
		case "phone":
			result.Device.Type = DeviceMobile
		case "tablet":
			result.Device.Type = DeviceTablet
		}
	}
}

func parseInstagramIAB(tokens []tokenizer.Token, result *Result, database *db.Database) bool {
	return parseMetaStyleIAB(tokens, result, database, "Instagram", "Instagram")
}

// parseBarcelonaIAB parses Threads by Meta (internal codename: Barcelona) in-app browser UAs.
// Barcelona uses the same positional comment format as Instagram.
func parseBarcelonaIAB(tokens []tokenizer.Token, result *Result, database *db.Database) bool {
	return parseMetaStyleIAB(tokens, result, database, "Barcelona", "Threads")
}

// parseMetaStyleIAB handles Meta-style IAB UAs (Instagram, Barcelona/Threads).
// These share the format: "<tokenName> <version> (<model>; <os>; <locale>; ...; scale=X; WxH; ...)".
func parseMetaStyleIAB(
	tokens []tokenizer.Token, result *Result, database *db.Database,
	tokenName, appName string,
) bool {
	idx := -1

	for i, tok := range tokens {
		if tok.Kind == tokenizer.KindProduct && tok.Name == tokenName {
			idx = i

			break
		}
	}

	if idx < 0 {
		return false
	}

	appVersion := tokens[idx].Version
	isAndroid := false

	// Meta UAs use space-separated format: "Instagram 419.0.0.27.74 (...)"
	// The version is a separate product token when there's no slash.
	// Also detect "Android" token which changes the comment attribute layout.
	for j := idx + 1; j < len(tokens); j++ {
		if tokens[j].Kind != tokenizer.KindProduct {
			break
		}

		if appVersion == "" {
			appVersion = tokens[j].Name
		}

		if tokens[j].Name == "Android" {
			isAndroid = true
		}
	}

	result.IAB = IABInfo{
		App:        appName,
		AppVersion: appVersion,
	}

	for i := idx + 1; i < len(tokens); i++ {
		if tokens[i].Kind != tokenizer.KindComment {
			continue
		}

		if isAndroid {
			parseMetaAndroidAttrs(tokens[i].Attrs, result, database)
		} else {
			parseMetaStyleAttrs(tokens[i].Attrs, result, database)
		}

		break
	}

	return true
}

func parseMetaStyleAttrs(attrs []string, result *Result, database *db.Database) {
	if len(attrs) >= igMinModel {
		model := strings.TrimSpace(attrs[igAttrModel])
		if model != "" {
			result.Device.ModelRaw = model

			if name, ok := database.AppleModels[model]; ok {
				result.Device.Model = name
			}
		}
	}

	if len(attrs) >= igMinOS {
		osStr := strings.TrimSpace(attrs[igAttrOS])

		if parts := strings.SplitN(osStr, " ", 2); len(parts) == 2 { //nolint:mnd // split "iOS 18_3"
			result.OS.Name = parts[0]
			result.OS.Version = strings.ReplaceAll(parts[1], "_", ".")
		}
	}

	if len(attrs) >= igMinLocale {
		locale := strings.TrimSpace(attrs[igAttrLocale])

		if at := strings.IndexByte(locale, '@'); at >= 0 {
			suffix := locale[at+1:]

			locale = locale[:at]

			if rg, ok := strings.CutPrefix(suffix, "rg="); ok && len(rg) >= 2 {
				result.IAB.Region = strings.ToUpper(rg[:2])
			}
		}

		result.IAB.Locale = locale
	}

	if len(attrs) >= igMinScale {
		scaleStr := strings.TrimSpace(attrs[igAttrScale])

		if strings.HasPrefix(scaleStr, scalePrefix) {
			s, err := strconv.ParseFloat(scaleStr[scalePrefLen:], 64)
			if err == nil {
				result.IAB.ScreenScale = s
			}
		}
	}

	if len(attrs) >= igMinResolution {
		result.IAB.Resolution = strings.TrimSpace(attrs[igAttrResolution])
	}
}

// parseMetaAndroidAttrs handles Instagram/Barcelona Android comment format.
// Android layout: (apiLevel/ver; dpi; WxH; brand; model; codename; chipset; locale; buildId)
func parseMetaAndroidAttrs(attrs []string, result *Result, database *db.Database) {
	if len(attrs) >= igAndroidMinModel {
		model := strings.TrimSpace(attrs[igAndroidAttrModel])
		if model != "" {
			result.Device.ModelRaw = model

			if dev, ok := database.AndroidDB[model]; ok {
				result.Device.Model = dev.Brand + " " + dev.Model
			}
		}
	}

	if len(attrs) >= igAndroidMinResolution {
		result.IAB.Resolution = strings.TrimSpace(attrs[igAndroidAttrResolution])
	}

	if len(attrs) >= igAndroidMinLocale {
		locale := strings.TrimSpace(attrs[igAndroidAttrLocale])

		if at := strings.IndexByte(locale, '@'); at >= 0 {
			suffix := locale[at+1:]

			locale = locale[:at]

			if rg, ok := strings.CutPrefix(suffix, "rg="); ok && len(rg) >= 2 {
				result.IAB.Region = strings.ToUpper(rg[:2])
			}
		}

		result.IAB.Locale = locale
	}
}

func parseTikTokIAB(tokens []tokenizer.Token, result *Result) bool {
	hasTikTok := false

	for _, tok := range tokens {
		if tok.Kind == tokenizer.KindProduct && tok.Name == "musical_ly" {
			hasTikTok = true

			result.IAB = IABInfo{
				App:        "TikTok",
				AppVersion: tok.Version,
			}

			break
		}
	}

	if !hasTikTok {
		return false
	}

	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct {
			continue
		}

		switch tok.Name {
		case "NetType":
			result.IAB.NetType = tok.Version
		case "ByteLocale":
			result.IAB.Locale = tok.Version
		case "Region":
			result.IAB.Region = tok.Version
		}
	}

	return true
}

// parseMetaIABGeneric handles bare "MetaIAB" token found in newer Meta in-app browser UAs.
// These UAs are standard Chrome WebView strings with "MetaIAB" appended, no additional metadata.
func parseMetaIABGeneric(tokens []tokenizer.Token, result *Result) bool {
	for _, tok := range tokens {
		if tok.Kind == tokenizer.KindProduct && tok.Name == "MetaIAB" {
			result.IAB = IABInfo{App: "Meta"}

			return true
		}
	}

	return false
}
