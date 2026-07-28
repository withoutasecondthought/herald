package herald

import (
	"strconv"
	"strings"

	"github.com/withoutasecondthought/herald/db"
	"github.com/withoutasecondthought/herald/tokenizer"
)

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
	igAndroidAttrOS         = 0
	igAndroidAttrResolution = 2
	igAndroidAttrBrand      = 3
	igAndroidAttrModel      = 4
	igAndroidAttrLocale     = 7

	igAndroidMinOS         = 1
	igAndroidMinResolution = 3
	igAndroidMinModel      = 5
	igAndroidMinLocale     = 8
)

const (
	appTikTok           = "TikTok"
	appTikTokLite       = "TikTok Lite"
	appTwitter          = "Twitter"
	tokenMusicalLy      = "musical_ly"
	tokenTwitterAndroid = "TwitterAndroid"
	keyIABMV            = "IABMV"
)

// fbAppNames maps FBAN/FB_IAB raw values to app names; unknown values pass through raw.
//
//nolint:gochecknoglobals // read-only lookup table
var fbAppNames = map[string]string{
	"FBIOS":               "Facebook",
	"FBForIPhone":         "Facebook",
	"FB4A":                "Facebook",
	"FBW":                 "Facebook",
	"MESSENGER":           "Messenger",
	"MessengerForiOS":     "Messenger",
	"Orca-Android":        "Messenger",
	"MessengerLiteForiOS": "Messenger Lite",
	"EMA":                 "Facebook Lite",
}

//nolint:gochecknoglobals // read-only lookup table
var iabProductApps = []struct{ token, app string }{
	{"Snapchat", "Snapchat"},
	{"Line", "LINE"},
	{"GSA", "Google App"},
}

// detectIAB parses in-app browser info from Facebook, Instagram, Threads, TikTok,
// Telegram, WeChat, Twitter, Snapchat, LINE, Google app, and Pinterest UAs.
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

	if parseTelegramIAB(tokens, result, database) {
		return true
	}

	if parseWeChatIAB(tokens, result) {
		return true
	}

	if parseTwitterIAB(tokens, result) {
		return true
	}

	if parseSimpleTokenIAB(tokens, result) {
		return true
	}

	if parseBracketIAB(tokens, result) {
		return true
	}

	return parseMetaIABGeneric(tokens, result)
}

func fbAppName(raw string) string {
	if name, ok := fbAppNames[raw]; ok {
		return name
	}

	return raw
}

func parseFacebookIAB(tokens []tokenizer.Token, result *Result, database *db.Database) bool {
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

		result.IAB = IABInfo{
			App:        fbAppName(app),
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
		} else if adev, ok := database.AndroidDB[dev]; ok {
			result.Device.Model = adev.Brand + " " + adev.Model
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
		result.IAB.Locale = parseIABLocale(strings.TrimSpace(attrs[igAttrLocale]), result)
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
// The apiLevel/ver attr carries the real Android version even when the base UA is frozen.
func parseMetaAndroidAttrs(attrs []string, result *Result, database *db.Database) {
	if len(attrs) >= igAndroidMinOS {
		if _, ver, ok := strings.Cut(strings.TrimSpace(attrs[igAndroidAttrOS]), "/"); ok && ver != "" {
			result.OS = OS{Name: OSAndroid, Version: ver}
		}
	}

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
		result.IAB.Locale = parseIABLocale(strings.TrimSpace(attrs[igAndroidAttrLocale]), result)
	}
}

func parseIABLocale(locale string, result *Result) string {
	at := strings.IndexByte(locale, '@')
	if at < 0 {
		return locale
	}

	suffix := locale[at+1:]

	locale = locale[:at]

	if rg, ok := strings.CutPrefix(suffix, "rg="); ok && len(rg) >= 2 {
		result.IAB.Region = strings.ToUpper(rg[:2])
	}

	return locale
}

// tikTokApp matches TikTok product tokens: "musical_ly_32.8.0" (iOS), "trill_<build>" plus
// "AppName/musical_ly app_version/x.y.z" (Android), "musically_go"/"ultralite" (TikTok Lite),
// and the legacy slash form "musical_ly/x.y.z".
func tikTokApp(tokens []tokenizer.Token) (string, string, bool) {
	app := ""
	version := ""
	fallbackVersion := ""

	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct {
			continue
		}

		switch tok.Name {
		case tokenMusicalLy:
			if tok.Version != "" {
				app = appTikTok
				version = tok.Version
			}
		case "AppName":
			switch tok.Version {
			case tokenMusicalLy, "trill":
				app = appTikTok
			case "musically_go", "ultralite":
				app = appTikTokLite
			}
		case "app_version":
			version = tok.Version
		default:
			app, fallbackVersion = matchTikTokPrefix(tok.Name, app, fallbackVersion)
		}
	}

	if version == "" {
		version = fallbackVersion
	}

	return app, version, app != ""
}

func isTikTokToken(tok tokenizer.Token) bool {
	if tok.Name == tokenMusicalLy && tok.Version != "" {
		return true
	}

	if strings.HasPrefix(tok.Name, "musical_ly_") ||
		strings.HasPrefix(tok.Name, "trill_") ||
		strings.HasPrefix(tok.Name, "musically_go_") {
		return true
	}

	if tok.Name == "AppName" {
		switch tok.Version {
		case tokenMusicalLy, "trill", "musically_go", "ultralite":
			return true
		}
	}

	return false
}

func matchTikTokPrefix(name, app, fallbackVersion string) (string, string) {
	if suffix, ok := strings.CutPrefix(name, "musical_ly_"); ok {
		return appTikTok, suffix
	}

	if suffix, ok := strings.CutPrefix(name, "musically_go_"); ok {
		return appTikTokLite, suffix
	}

	if suffix, ok := strings.CutPrefix(name, "trill_"); ok {
		if app == "" {
			app = appTikTok
		}

		if fallbackVersion == "" {
			fallbackVersion = suffix
		}
	}

	return app, fallbackVersion
}

func parseTikTokIAB(tokens []tokenizer.Token, result *Result) bool {
	app, version, ok := tikTokApp(tokens)
	if !ok {
		return false
	}

	result.IAB = IABInfo{App: app, AppVersion: version}

	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct {
			continue
		}

		switch tok.Name {
		case "NetType":
			result.IAB.NetType = tok.Version
		case "ByteLocale":
			result.IAB.Locale = tok.Version
		case "ByteFullLocale":
			if result.IAB.Locale == "" {
				result.IAB.Locale = tok.Version
			}
		case "Region":
			result.IAB.Region = tok.Version
		}
	}

	return true
}

// parseTelegramIAB handles "Telegram-Android/11.9.0 (Samsung SM-A155M; Android 14; SDK 34; AVERAGE)".
// The trailing comment carries the real model and OS version even when the base UA is frozen.
// Telegram iOS sends a plain Safari UA with no token — not detectable.
func parseTelegramIAB(tokens []tokenizer.Token, result *Result, database *db.Database) bool {
	idx := -1

	for i, tok := range tokens {
		if tok.Kind == tokenizer.KindProduct && tok.Name == "Telegram-Android" {
			idx = i

			break
		}
	}

	if idx < 0 {
		return false
	}

	result.IAB = IABInfo{App: "Telegram", AppVersion: tokens[idx].Version}

	for i := idx + 1; i < len(tokens); i++ {
		if tokens[i].Kind != tokenizer.KindComment {
			continue
		}

		applyTelegramDeviceAttrs(tokens[i].Attrs, result, database)

		break
	}

	return true
}

// applyTelegramDeviceAttrs reads "(<Build.MANUFACTURER> <Build.MODEL>; Android <ver>; ...)".
func applyTelegramDeviceAttrs(attrs []string, result *Result, database *db.Database) {
	if len(attrs) < 2 || !strings.HasPrefix(attrs[1], OSAndroid+" ") {
		return
	}

	result.OS = OS{Name: OSAndroid, Version: strings.TrimSpace(attrs[1][len(OSAndroid):])}

	model := strings.TrimSpace(attrs[0])
	if _, rest, ok := strings.Cut(model, " "); ok && rest != "" {
		model = rest
	}

	if model == "" {
		return
	}

	result.Device.ModelRaw = model

	if dev, ok := database.AndroidDB[model]; ok {
		result.Device.Model = dev.Brand + " " + dev.Model
	}
}

func parseWeChatIAB(tokens []tokenizer.Token, result *Result) bool {
	found := false

	for _, tok := range tokens {
		if tok.Kind == tokenizer.KindProduct && tok.Name == "MicroMessenger" {
			result.IAB = IABInfo{App: "WeChat", AppVersion: tok.Version}
			found = true

			break
		}
	}

	if !found {
		return false
	}

	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct {
			continue
		}

		switch tok.Name {
		case "NetType":
			result.IAB.NetType = tok.Version
		case "Language":
			result.IAB.Locale = tok.Version
		}
	}

	return true
}

// parseTwitterIAB handles "TwitterAndroid" (bare token) and "Twitter for iPhone/12.5" forms.
func parseTwitterIAB(tokens []tokenizer.Token, result *Result) bool {
	for i, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct {
			continue
		}

		if tok.Name == tokenTwitterAndroid {
			result.IAB = IABInfo{App: appTwitter, AppVersion: tok.Version}

			return true
		}

		if tok.Name == appTwitter && twitterForDevice(tokens, i, result) {
			return true
		}
	}

	return false
}

func twitterForDevice(tokens []tokenizer.Token, idx int, result *Result) bool {
	dev, ok := twitterForTarget(tokens, idx)
	if !ok {
		return false
	}

	result.IAB = IABInfo{App: appTwitter, AppVersion: dev.Version}

	return true
}

func twitterForTarget(tokens []tokenizer.Token, idx int) (tokenizer.Token, bool) {
	if idx+2 >= len(tokens) {
		return tokenizer.Token{}, false
	}

	next, dev := tokens[idx+1], tokens[idx+2]
	if next.Kind != tokenizer.KindProduct || next.Name != "for" || dev.Kind != tokenizer.KindProduct {
		return tokenizer.Token{}, false
	}

	if dev.Name != "iPhone" && dev.Name != "iPad" {
		return tokenizer.Token{}, false
	}

	return dev, true
}

func parseSimpleTokenIAB(tokens []tokenizer.Token, result *Result) bool {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct {
			continue
		}

		for _, e := range iabProductApps {
			if tok.Name != e.token {
				continue
			}

			version := strings.TrimSuffix(tok.Version, "/IAB")

			result.IAB = IABInfo{App: e.app, AppVersion: version}

			return true
		}
	}

	return false
}

// parseBracketIAB handles non-Facebook bracket blocks like "[Pinterest/iOS]",
// mapping the block's first key to the app name.
func parseBracketIAB(tokens []tokenizer.Token, result *Result) bool {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindFBBlock || len(tok.Attrs) == 0 {
			continue
		}

		key, _, found := strings.Cut(tok.Attrs[0], "/")
		if !found || key == "" || strings.HasPrefix(key, "FB") || key == keyIABMV {
			continue
		}

		result.IAB = IABInfo{App: key}

		return true
	}

	return false
}

// parseMetaIABGeneric handles bare "MetaIAB" and "IABMV" markers found in newer Meta
// in-app browser UAs that carry no other app token.
func parseMetaIABGeneric(tokens []tokenizer.Token, result *Result) bool {
	for _, tok := range tokens {
		switch tok.Kind {
		case tokenizer.KindProduct:
			if tok.Name == "MetaIAB" || tok.Name == keyIABMV {
				result.IAB = IABInfo{App: "Meta"}

				return true
			}
		case tokenizer.KindFBBlock:
			for _, attr := range tok.Attrs {
				if strings.HasPrefix(attr, keyIABMV+"/") {
					result.IAB = IABInfo{App: "Meta"}

					return true
				}
			}
		case tokenizer.KindComment:
		}
	}

	return false
}
