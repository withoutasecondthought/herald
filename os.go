package herald

import (
	"strconv"
	"strings"

	"github.com/withoutasecondthought/herald/db"
	"github.com/withoutasecondthought/herald/tokenizer"
)

//nolint:gochecknoglobals // read-only lookup table
var windowsVersions = map[string]string{
	"10.0": "10",
	"6.3":  "8.1",
	"6.2":  "8",
	"6.1":  "7",
	"6.0":  "Vista",
	"5.1":  "XP",
	"5.0":  "2000",
}

const macOSXPrefix = "Mac OS X "

// frozenMacOSVersion is the value all modern browsers permanently report on macOS;
// it is a freeze sentinel, not the real OS version.
const frozenMacOSVersion = "10.15.7"

// iosFreezeSafariMajor marks Safari 26+, whose UAs freeze the OS token at 18_6/18_7
// while the Version/ token keeps reporting the real coupled iOS version.
const iosFreezeSafariMajor = 26

// frozenIOSBuild is the Mobile/ token value real Safari has pinned since iOS 11.3;
// WKWebView UAs carry the actual OS build instead.
const frozenIOSBuild = "15E148"

// detectOS parses OS information from comment tokens and product tokens.
func detectOS(tokens []tokenizer.Token, result *Result, database *db.Database) {
	if detectOSFromComments(tokens, result) {
		if result.OS.Name == OSiOS {
			result.OS.Build = extractIOSBuild(tokens)
		}

		fixFrozenIOSVersion(tokens, result)

		return
	}

	detectOSFromDarwin(tokens, result, database)
}

func detectOSFromComments(tokens []tokenizer.Token, result *Result) bool {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindComment {
			continue
		}

		for _, attr := range tok.Attrs {
			attr = strings.TrimSpace(attr)

			if found := matchOSAttr(attr, tok.Attrs, result); found {
				return true
			}
		}
	}

	return false
}

func matchOSAttr(attr string, allAttrs []string, result *Result) bool {
	if strings.Contains(attr, "CPU iPhone OS") || strings.Contains(attr, "CPU OS") {
		if v := parseIOSVersion(attr); v != "" {
			result.OS = OS{Name: OSiOS, Version: v}

			return true
		}
	}

	if matchAndroidAttr(attr, allAttrs, result) {
		return true
	}

	if matchWindowsAttr(attr, result) {
		return true
	}

	if matchMacintoshAttr(attr, allAttrs, result) {
		return true
	}

	if strings.HasPrefix(attr, "CrOS") {
		result.OS = OS{Name: OSChromeOS}

		return true
	}

	if strings.HasPrefix(attr, OSLinux) && result.OS.Name == "" {
		result.OS = OS{Name: OSLinux}
	}

	return false
}

func matchAndroidAttr(attr string, allAttrs []string, result *Result) bool {
	v, ok := strings.CutPrefix(attr, OSAndroid)
	if !ok {
		return false
	}

	ver := strings.TrimSpace(v)
	if ver == "10" && hasFrozenModelK(allAttrs) {
		ver = ""
	}

	result.OS = OS{Name: OSAndroid, Version: ver}

	return true
}

func matchWindowsAttr(attr string, result *Result) bool {
	ntVer, ok := strings.CutPrefix(attr, "Windows NT")
	if !ok {
		return false
	}

	ntVer = strings.TrimSpace(ntVer)

	winVer := ntVer
	if mapped, ok := windowsVersions[ntVer]; ok {
		winVer = mapped
	}

	result.OS = OS{Name: OSWindows, Version: winVer}

	return true
}

func matchMacintoshAttr(attr string, allAttrs []string, result *Result) bool {
	if !strings.HasPrefix(attr, "Macintosh") {
		return false
	}

	result.OS = OS{Name: OSmacOS}

	for _, a2 := range allAttrs {
		a2 = strings.TrimSpace(a2)

		if v := parseMacOSVersion(a2); v != "" && v != frozenMacOSVersion {
			result.OS.Version = v
		}
	}

	return true
}

// hasFrozenModelK reports whether the comment carries the reduced-UA placeholder model
// ("Android 10; K" / "Android 10; K; wv"), meaning the version is a freeze constant.
func hasFrozenModelK(attrs []string) bool {
	for _, a := range attrs {
		if strings.TrimSpace(a) == "K" {
			return true
		}
	}

	return false
}

func fixFrozenIOSVersion(tokens []tokenizer.Token, result *Result) {
	if result.OS.Name != OSiOS {
		return
	}

	frozen := strings.HasPrefix(result.OS.Version, "18.6") || strings.HasPrefix(result.OS.Version, "18.7")
	if !frozen {
		return
	}

	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct || tok.Name != "Version" || tok.Version == "" {
			continue
		}

		major, _, _ := strings.Cut(tok.Version, ".")

		n, err := strconv.Atoi(major)
		if err == nil && n >= iosFreezeSafariMajor {
			result.OS.Version = tok.Version
		}

		return
	}
}

func extractIOSBuild(tokens []tokenizer.Token) string {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct || tok.Name != "Mobile" {
			continue
		}

		if tok.Version == frozenIOSBuild || !isIOSBuild(tok.Version) {
			return ""
		}

		return tok.Version
	}

	return ""
}

const iosBuildMinLen = 4

func isIOSBuild(v string) bool {
	if len(v) < iosBuildMinLen {
		return false
	}

	return isASCIIDigit(v[0]) && isASCIIDigit(v[1]) && v[2] >= 'A' && v[2] <= 'Z' && isASCIIDigit(v[3])
}

func isASCIIDigit(b byte) bool { return b >= '0' && b <= '9' }

func detectOSFromDarwin(tokens []tokenizer.Token, result *Result, database *db.Database) {
	for _, tok := range tokens {
		if tok.Kind != tokenizer.KindProduct || tok.Name != ProductDarwin {
			continue
		}

		if v := tok.Version; v != "" {
			major := strings.SplitN(v, ".", 2)[0] //nolint:mnd // split major.minor

			if dv, ok := database.DarwinMap[major]; ok {
				result.OS = OS{Name: OSDarwin, Version: dv.IOS}
			}
		}

		return
	}
}

func parseIOSVersion(s string) string {
	_, rest, found := strings.Cut(s, "OS ")
	if !found {
		return ""
	}

	if likeIdx := strings.Index(rest, " like"); likeIdx >= 0 {
		rest = rest[:likeIdx]
	}

	rest = strings.TrimSpace(rest)

	return strings.ReplaceAll(rest, "_", ".")
}

func parseMacOSVersion(s string) string {
	_, v, found := strings.Cut(s, macOSXPrefix)
	if !found {
		return ""
	}

	v = strings.TrimSpace(v)

	return strings.ReplaceAll(v, "_", ".")
}
