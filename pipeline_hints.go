package herald

import (
	"strconv"
	"strings"
)

const (
	win11MinPlatformVersion = 13
	chVersionPrefixLen      = 2 // "v=" length
)

// applyHints enriches the result with Client Hints data.
// Client Hints take priority over UA string data.
func applyHints(hints *ClientHints, result *Result) {
	if hints == nil || hints.IsEmpty() {
		return
	}

	applyPlatformHint(hints, result)
	applyBrowserHint(hints, result)

	if hints.Model != "" {
		result.Device.ModelRaw = hints.Model
	}

	if hints.Mobile && (result.Device.Type == "" || result.Device.Type == DeviceDesktop) {
		result.Device.Type = DeviceMobile
	}
}

func applyPlatformHint(hints *ClientHints, result *Result) {
	if hints.Platform == "" {
		return
	}

	osName, hasVersion := mapPlatformToOS(strings.ToLower(hints.Platform))
	if osName == "" {
		return
	}

	result.OS.Name = osName

	if !hasVersion || hints.PlatformVersion == "" {
		return
	}

	if osName == OSWindows {
		result.OS.Version = resolveWindowsVersion(hints.PlatformVersion)
	} else {
		result.OS.Version = hints.PlatformVersion
	}
}

func mapPlatformToOS(platform string) (name string, hasVersion bool) {
	switch platform {
	case "windows":
		return OSWindows, true
	case "macos":
		return OSmacOS, true
	case "android":
		return OSAndroid, true
	case "ios":
		return OSiOS, true
	case "linux":
		return OSLinux, false
	case "chrome os":
		return OSChromeOS, true
	default:
		return "", false
	}
}

func applyBrowserHint(hints *ClientHints, result *Result) {
	src := hints.FullVersionList
	if src == "" {
		src = hints.UA
	}

	if src == "" {
		return
	}

	name, version := parseCHUA(src)
	if name != "" {
		result.Browser.Name = name
		result.Browser.Version = version
	}
}

func resolveWindowsVersion(platformVersion string) string {
	parts := strings.SplitN(platformVersion, ".", 2) //nolint:mnd // split major.minor

	if len(parts) == 0 {
		return "10"
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return "10"
	}

	if major >= win11MinPlatformVersion {
		return "11"
	}

	return "10"
}

func parseCHUA(s string) (string, string) {
	entries := strings.Split(s, ",")

	bestName := ""
	bestVersion := ""

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)

		name, version := parseCHEntry(entry)
		if name == "" {
			continue
		}

		lower := strings.ToLower(name)
		if strings.Contains(lower, "not") || lower == "chromium" {
			if bestName == "" {
				bestName = name
				bestVersion = version
			}

			continue
		}

		bestName = name
		bestVersion = version
	}

	return bestName, bestVersion
}

func parseCHEntry(entry string) (string, string) {
	parts := strings.SplitN(entry, ";", 2) //nolint:mnd // split name;version

	if len(parts) == 0 {
		return "", ""
	}

	name := strings.Trim(strings.TrimSpace(parts[0]), "\"")

	version := ""
	if len(parts) > 1 {
		vPart := strings.TrimSpace(parts[1])

		if strings.HasPrefix(vPart, "v=") {
			version = strings.Trim(vPart[chVersionPrefixLen:], "\"")
		}
	}

	return name, version
}
