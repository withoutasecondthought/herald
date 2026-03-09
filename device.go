package herald

import (
	"strings"

	"github.com/withoutasecondthought/herald/db"
	"github.com/withoutasecondthought/herald/tokenizer"
)

// detectDevice determines device type, raw model, and human-readable model name.
func detectDevice(tokens []tokenizer.Token, result *Result, database *db.Database) {
	switch result.OS.Name {
	case OSiOS, OSDarwin:
		detectAppleDevice(tokens, result, database)
	case OSAndroid:
		detectAndroidDevice(tokens, result, database)
	case OSWindows, OSmacOS, OSLinux:
		result.Device = Device{Type: DeviceDesktop}
	}
}

func detectAppleDevice(tokens []tokenizer.Token, result *Result, database *db.Database) {
	rawModel := findAppleModelRaw(tokens)
	deviceType := resolveAppleDeviceType(rawModel, result)

	result.Device = Device{
		Type:     deviceType,
		ModelRaw: rawModel,
	}

	if rawModel != "" {
		if name, ok := database.AppleModels[rawModel]; ok {
			result.Device.Model = name
		}
	}
}

func findAppleModelRaw(tokens []tokenizer.Token) string {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindComment {
			continue
		}

		for _, attr := range tok.Attrs {
			attr = strings.TrimSpace(attr)

			isApple := strings.HasPrefix(attr, "iPhone") ||
				strings.HasPrefix(attr, "iPad") ||
				strings.HasPrefix(attr, "iPod")

			if isApple && strings.Contains(attr, ",") {
				return attr
			}
		}
	}

	return ""
}

func resolveAppleDeviceType(rawModel string, result *Result) string {
	if strings.HasPrefix(rawModel, "iPad") {
		return DeviceTablet
	}

	if (result.OS.Name == OSmacOS || result.OS.Name == OSDarwin) && rawModel == "" {
		return DeviceDesktop
	}

	if rawModel == "" {
		if strings.Contains(result.Raw, "iPad") {
			return DeviceTablet
		}
	}

	return DeviceMobile
}

func detectAndroidDevice(tokens []tokenizer.Token, result *Result, database *db.Database) {
	rawModel := findAndroidModelRaw(tokens)
	deviceType := resolveAndroidDeviceType(rawModel, result.Raw)

	result.Device = Device{
		Type:     deviceType,
		ModelRaw: rawModel,
	}

	if rawModel != "" {
		if dev, ok := database.AndroidDB[rawModel]; ok {
			result.Device.Model = dev.Brand + " " + dev.Model
		}
	}
}

func findAndroidModelRaw(tokens []tokenizer.Token) string {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindComment {
			continue
		}

		foundAndroid := false

		for _, attr := range tok.Attrs {
			attr = strings.TrimSpace(attr)

			if strings.HasPrefix(attr, OSAndroid) {
				foundAndroid = true

				continue
			}

			if foundAndroid && attr != "" && !isLocaleCode(attr) {
				model := stripBrandPrefix(attr)

				if buildIdx := strings.Index(model, " Build/"); buildIdx >= 0 {
					model = model[:buildIdx]
				}

				return strings.TrimSpace(model)
			}
		}
	}

	return ""
}

func resolveAndroidDeviceType(rawModel, rawUA string) string {
	if strings.Contains(rawUA, "Tablet") {
		return DeviceTablet
	}

	if strings.HasPrefix(rawModel, "SM-T") || strings.HasPrefix(rawModel, "SM-X") {
		return DeviceTablet
	}

	return DeviceMobile
}

// isLocaleCode returns true if s looks like a BCP 47 locale tag (e.g., "zh-cn", "en-US").
// This prevents locale strings from being misidentified as device models.
func isLocaleCode(s string) bool {
	if len(s) != 5 { //nolint:mnd // xx-XX is always 5 chars
		return false
	}

	sep := s[2]

	return (sep == '-' || sep == '_') &&
		isLowerAlpha(s[0]) && isLowerAlpha(s[1]) &&
		isAlpha(s[3]) && isAlpha(s[4])
}

func isLowerAlpha(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func stripBrandPrefix(model string) string {
	prefixes := []string{
		"SAMSUNG ", "samsung ",
		"HUAWEI ", "huawei ",
		"OPPO ", "Oppo ",
		"vivo ", "Vivo ",
		"Motorola ", "motorola ",
		"Sony ", "sony ",
		"Infinix ", "infinix ",
		"TECNO ", "Tecno ",
		"Realme ", "realme ",
		"LGE ",
	}

	for _, p := range prefixes {
		if strings.HasPrefix(model, p) {
			return model[len(p):]
		}
	}

	return model
}
