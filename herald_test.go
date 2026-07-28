package herald_test

import (
	"os"
	"path/filepath"
	"testing"

	herald "github.com/withoutasecondthought/herald"
)

func newTestParser(t *testing.T) *herald.Parser {
	t.Helper()

	p, err := herald.NewParser()
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	return p
}

// Common UA strings used across tests.
const (
	uaChromeWin = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/120.0.0.0 Safari/537.36"
	uaSafariIOS = "Mozilla/5.0 " +
		"(iPhone; CPU iPhone OS 18_7 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"Version/18.0 Mobile/15E148 Safari/604.1"
)

func TestParse_ChromeDesktop(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse(uaChromeWin)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBrowser)
	assertEq(t, "Browser.Name", r.Browser.Name, "Chrome")
	assertEq(t, "Browser.Version", r.Browser.Version, "120.0.0.0")
	assertEq(t, "Browser.Engine", r.Browser.Engine, "Blink")
	assertEq(t, "OS.Name", r.OS.Name, "Windows")
	assertEq(t, "OS.Version", r.OS.Version, "10")
	assertEq(t, "Device.Type", r.Device.Type, "desktop")
}

func TestParse_SafariIOS(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse(uaSafariIOS)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBrowser)
	assertEq(t, "Browser.Name", r.Browser.Name, "Safari")
	assertEq(t, "Browser.Version", r.Browser.Version, "18.0")
	assertEq(t, "Browser.Engine", r.Browser.Engine, "WebKit")
	assertEq(t, "OS.Name", r.OS.Name, "iOS")
	assertEq(t, "OS.Version", r.OS.Version, "18.7")
	assertEq(t, "Device.Type", r.Device.Type, "mobile")
}

func TestParse_Firefox(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse(
		"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) " +
			"Gecko/20100101 Firefox/115.0",
	)

	assertEq(t, "Browser.Name", r.Browser.Name, "Firefox")
	assertEq(t, "Browser.Version", r.Browser.Version, "115.0")
	assertEq(t, "Browser.Engine", r.Browser.Engine, "Gecko")
	assertEq(t, "OS.Name", r.OS.Name, "Linux")
}

func TestParse_Edge(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse(uaChromeWin + " Edg/120.0.2210.91")

	assertEq(t, "Browser.Name", r.Browser.Name, "Edge")
	assertEq(t, "Browser.Version", r.Browser.Version, "120.0.2210.91")
	assertEq(t, "Browser.Engine", r.Browser.Engine, "Blink")
}

func TestParse_Opera(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse(uaChromeWin + " OPR/106.0.0.0")

	assertEq(t, "Browser.Name", r.Browser.Name, "Opera")
	assertEq(t, "Browser.Version", r.Browser.Version, "106.0.0.0")
}

func TestParse_YandexBrowser(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/120.0.0.0 YaBrowser/24.1.0 Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "Browser.Name", r.Browser.Name, "Yandex Browser")
}

func TestParse_SamsungBrowser(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 13; SM-G991B) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"SamsungBrowser/23.0 Chrome/115.0.0.0 Mobile Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "Browser.Name", r.Browser.Name, "Samsung Browser")
	assertEq(t, "OS.Name", r.OS.Name, "Android")
	assertEq(t, "OS.Version", r.OS.Version, "13")
	assertEq(t, "Device.Type", r.Device.Type, "mobile")
	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "SM-G991B")
	assertEq(t, "Device.Model", r.Device.Model, "Samsung Galaxy S21")
}

func TestParse_ChromeIOS(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 " +
		"(iPhone; CPU iPhone OS 17_0 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"CriOS/120.0.6099.119 Mobile/15E148 Safari/604.1"
	r := p.Parse(ua)

	assertEq(t, "Browser.Name", r.Browser.Name, "Chrome")
	assertEq(t, "Browser.Engine", r.Browser.Engine, "WebKit")
	assertEq(t, "OS.Name", r.OS.Name, "iOS")
}

func TestParse_SafariMacOS(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 " +
		"(Macintosh; Intel Mac OS X 10_15_7) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"Version/17.0 Safari/605.1.15"
	r := p.Parse(ua)

	assertEq(t, "Browser.Name", r.Browser.Name, "Safari")
	assertEq(t, "OS.Name", r.OS.Name, "macOS")
	assertEq(t, "OS.Version", r.OS.Version, "")
	assertEq(t, "Device.Type", r.Device.Type, "desktop")
}

func TestParse_Googlebot(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (compatible; Googlebot/2.1; " +
		"+http://www.google.com/bot.html)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBot)
	assertEq(t, "Bot.Name", r.Bot.Name, "Googlebot")
	assertEq(t, "Bot.Owner", r.Bot.Owner, "Google")
	assertEq(t, "Bot.Category", r.Bot.Category, herald.BotCategorySearch)
	assertFloat(t, "Bot.Confidence", r.Bot.Confidence, 1.0)
}

func TestParse_GPTBot(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 AppleWebKit/537.36 " +
		"(KHTML, like Gecko; compatible; GPTBot/1.0; " +
		"+https://openai.com/gptbot)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBot)
	assertEq(t, "Bot.Name", r.Bot.Name, "GPTBot")
	assertEq(t, "Bot.Owner", r.Bot.Owner, "OpenAI")
	assertEq(t, "Bot.Category", r.Bot.Category, herald.BotCategoryAI)
}

func TestParse_ClaudeBot(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("ClaudeBot/1.0")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBot)
	assertEq(t, "Bot.Name", r.Bot.Name, "ClaudeBot")
	assertEq(t, "Bot.Owner", r.Bot.Owner, "Anthropic")
}

func TestParse_Bingbot(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (compatible; bingbot/2.0; " +
		"+http://www.bing.com/bingbot.htm)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBot)
	assertEq(t, "Bot.Name", r.Bot.Name, "Bing Bot")
}

func TestParse_CFNetwork(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("ut-1/3 CFNetwork/3860.400.51 Darwin/25.3.0")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeNativeApp)
	assertEq(t, "Native.Name", r.Native.Name, "ut-1")
	assertEq(t, "Native.Version", r.Native.Version, "3")
	assertEq(t, "Native.Runtime", r.Native.Runtime, "CFNetwork")
	assertEq(t, "OS.Name", r.OS.Name, "Darwin")
	assertEq(t, "OS.Version", r.OS.Version, "26")
}

func TestParse_OkHttp(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("okhttp/4.12.0")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeNativeApp)
	assertEq(t, "Native.Name", r.Native.Name, "OkHttp")
	assertEq(t, "Native.Version", r.Native.Version, "4.12.0")
	assertEq(t, "Native.Runtime", r.Native.Runtime, "OkHttp")
}

func TestParse_Dart(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("Dart/3.10 (dart:io)")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeHttpClient)
	assertEq(t, "Native.Name", r.Native.Name, "Dart")
	assertEq(t, "Native.Version", r.Native.Version, "3.10")
	assertEq(t, "Native.Runtime", r.Native.Runtime, "dart:io")
}

func TestParse_Curl(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("curl/7.68.0")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeHttpClient)
	assertEq(t, "Native.Name", r.Native.Name, "curl")
}

func TestParse_Empty(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeUnknown)

	if !r.Browser.IsEmpty() {
		t.Error("Browser should be empty")
	}

	if !r.OS.IsEmpty() {
		t.Error("OS should be empty")
	}

	if !r.Device.IsEmpty() {
		t.Error("Device should be empty")
	}
}

func TestParseWithHints_Windows11(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)

	hints := herald.ClientHints{
		UA:              `"Chromium";v="120", "Google Chrome";v="120"`,
		Platform:        "Windows",
		PlatformVersion: "15.0.0",
	}
	r := p.ParseWithHints(uaChromeWin, hints)

	assertEq(t, "OS.Name", r.OS.Name, "Windows")
	assertEq(t, "OS.Version", r.OS.Version, "11")
	assertEq(t, "Browser.Name", r.Browser.Name, "Chrome")
	assertEq(t, "Browser.Version", r.Browser.Version, "120")
}

func TestParseWithHints_Windows10(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)

	hints := herald.ClientHints{
		Platform:        "Windows",
		PlatformVersion: "10.0.0",
	}
	r := p.ParseWithHints(uaChromeWin, hints)

	assertEq(t, "OS.Version", r.OS.Version, "10")
}

func TestDetectType_Bot(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ct := p.DetectType("Googlebot/2.1")

	assertEq(t, "ClientType", ct, herald.ClientTypeBot)
}

func TestDetectType_Browser(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
		"Chrome/120.0.0.0"
	ct := p.DetectType(ua)

	assertEq(t, "ClientType", ct, herald.ClientTypeBrowser)
}

func TestDetectType_Empty(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ct := p.DetectType("")

	assertEq(t, "ClientType", ct, herald.ClientTypeUnknown)
}

func TestParse_TikTok(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 " +
		"(iPhone; CPU iPhone OS 17_0 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"musical_ly/33.0.0 NetType/WIFI ByteLocale/en Region/US"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "TikTok")
	assertEq(t, "IAB.NetType", r.IAB.NetType, "WIFI")
	assertEq(t, "IAB.Locale", r.IAB.Locale, "en")
	assertEq(t, "IAB.Region", r.IAB.Region, "US")
}

func TestParse_AndroidChrome(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 14; Pixel 8) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/120.0.0.0 Mobile Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "OS.Name", r.OS.Name, "Android")
	assertEq(t, "OS.Version", r.OS.Version, "14")
	assertEq(t, "Device.Type", r.Device.Type, "mobile")
	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "Pixel 8")
	assertEq(t, "Device.Model", r.Device.Model, "Google Pixel 8")
}

func TestParse_AndroidTablet(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 13; SM-X810) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/120.0.0.0 Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "Device.Type", r.Device.Type, "tablet")
	assertEq(t, "Device.Model", r.Device.Model, "Samsung Galaxy Tab S9+")
}

func TestParse_InstagramApp(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/23D127 " +
		"Instagram 419.0.0.27.74 (iPhone17,2; iOS 26_3; en_US; en; " +
		"scale=3.00; 1320x2868; IABMV/1; 895010607) Safari/604.1"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Instagram")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "419.0.0.27.74")
	assertEq(t, "IAB.Locale", r.IAB.Locale, "en_US")
	assertFloat(t, "IAB.ScreenScale", r.IAB.ScreenScale, 3.0)
	assertEq(t, "Bot.Name", r.Bot.Name, "")
}

func TestParse_InstagramAndroid(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 10; CDY-NX9A Build/HUAWEICDY-N29; wv) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 " +
		"Chrome/114.0.5735.196 Mobile Safari/537.36 " +
		"Instagram 401.0.0.48.79 Android " +
		"(29/10; 360dpi; 720x1448; HUAWEI; CDY-NX9A; HWCDY; kirin820; it_IT; 802602927)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Instagram")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "401.0.0.48.79")
	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "CDY-NX9A")
	assertEq(t, "IAB.Locale", r.IAB.Locale, "it_IT")
	assertEq(t, "IAB.Resolution", r.IAB.Resolution, "720x1448")
	assertEq(t, "Bot.Name", r.Bot.Name, "")
}

func TestParse_2ipBot(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("2ip bot/1.1 (+http://2ip.io)")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBot)
	assertEq(t, "Bot.Name", r.Bot.Name, "2ip Bot")
}

func TestParse_CriteoBot(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("CriteoBot/0.1 (+https://www.criteo.com/criteo-crawler/)")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBot)
	assertEq(t, "Bot.Name", r.Bot.Name, "CriteoBot")
	assertEq(t, "Bot.Owner", r.Bot.Owner, "Criteo")
}

func TestParse_Barcelona(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Barcelona 404.1.0.30.79 " +
		"(iPhone12,1; iOS 17_6_1; en_US; en; scale=2.00; 828x1792; 813788720)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Threads")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "404.1.0.30.79")
	assertEq(t, "IAB.Locale", r.IAB.Locale, "en_US")
	assertEq(t, "IAB.Resolution", r.IAB.Resolution, "828x1792")
	assertFloat(t, "IAB.ScreenScale", r.IAB.ScreenScale, 2.0)
	assertEq(t, "OS.Name", r.OS.Name, "iOS")
	assertEq(t, "OS.Version", r.OS.Version, "17.6.1")
	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "iPhone12,1")
}

func TestParse_EmbarcaderoURIClient(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("Embarcadero URI Client/1.0")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeHttpClient)
}

func TestParse_EESUpdate(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "EES Update (Windows; U; 64bit; BPC 12.0.2058.0; " +
		"OS: 10.0.26100 SP 0.0 NT; TDB 66727)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeHttpClient)
}

func TestParse_InstagramLocaleRegion(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 18_3 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/22D63 " +
		"Instagram 374.1.3.38.95 (iPhone14,5; iOS 18_3; en_GB@rg=npzzzz; en; " +
		"scale=3.00; 1170x2532; 715850892)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Instagram")
	assertEq(t, "IAB.Locale", r.IAB.Locale, "en_GB")
	assertEq(t, "IAB.Region", r.IAB.Region, "NP")
	assertEq(t, "IAB.Resolution", r.IAB.Resolution, "1170x2532")
	assertFloat(t, "IAB.ScreenScale", r.IAB.ScreenScale, 3.0)
	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "iPhone14,5")
	assertEq(t, "Device.Model", r.Device.Model, "iPhone 13")
	assertEq(t, "OS.Name", r.OS.Name, "iOS")
	assertEq(t, "OS.Version", r.OS.Version, "18.3")
}

func TestParse_InstagramAndroidDevice(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 14; SM-A546B Build/UP1A.231005.007; wv) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 " +
		"Chrome/131.0.6778.200 Mobile Safari/537.36 " +
		"Instagram 371.1.0.44.107 Android " +
		"(34/14; 450dpi; 1080x2177; samsung; SM-A546B; a54x; s5e8835; it_IT@rg=itzzzz; UP1A.231005.007)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Instagram")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "371.1.0.44.107")
	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "SM-A546B")
	assertEq(t, "Device.Model", r.Device.Model, "Samsung Galaxy A54")
	assertEq(t, "IAB.Locale", r.IAB.Locale, "it_IT")
	assertEq(t, "IAB.Region", r.IAB.Region, "IT")
	assertEq(t, "IAB.Resolution", r.IAB.Resolution, "1080x2177")
	assertEq(t, "OS.Name", r.OS.Name, "Android")
	assertEq(t, "OS.Version", r.OS.Version, "14")
}

func TestParse_WinHttpRequest(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("Mozilla/4.0 (compatible; Win32; WinHttp.WinHttpRequest.5)")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeHttpClient)
}

func TestParse_AppEngineGoogle(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; Trident/6.0; " +
		"AppEngine-Google; (+http://code.google.com/appengine; " +
		"appid: s~virustotalcloud))"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBot)
	assertEq(t, "Bot.Name", r.Bot.Name, "Google App Engine")
	assertEq(t, "Bot.Owner", r.Bot.Owner, "Google")
}

func TestParse_LocaleNotModel(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 10; zh-cn) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/90.0.4430.210 Mobile Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "")
	assertEq(t, "OS.Name", r.OS.Name, "Android")
	assertEq(t, "OS.Version", r.OS.Version, "10")
}

func TestParse_Aiohttp(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("Python/3.10 aiohttp/3.11.11")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeHttpClient)
}

func TestParse_GoResty(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("go-resty/2.16.2 (https://github.com/go-resty/resty)")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeHttpClient)
}

func TestParse_MetaExternalAgent(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "meta-externalagent/1.1 " +
		"(+https://developers.facebook.com/docs/sharing/webmasters/crawler)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBot)
	assertEq(t, "Bot.Name", r.Bot.Name, "Meta External Agent")
}

func TestParse_TurnitinBot(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("TurnitinBot (https://turnitin.com/robot/crawlerinfo.html)")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBot)
	assertEq(t, "Bot.Name", r.Bot.Name, "TurnitinBot")
}

func TestParse_SogouMobileBrowserNotBot(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 9.0; MHA-AL00) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 " +
		"Chrome/64.0.3282.137 Mobile Safari/537.36 " +
		"SogouMSE,SogouMobileBrowser/5.8.18"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBrowser)
}

func TestParse_SogouSpiderIsBot(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("Sogou web spider/4.0 (+http://www.sogou.com/docs/help/webmasters.htm#07)")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBot)
	assertEq(t, "Bot.Name", r.Bot.Name, "Sogou Spider")
}

func TestParse_Dalvik(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Dalvik/2.1.0 (Linux; U; Android 10; " +
		"Mi A2 Build/QKQ1.190910.002)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeNativeApp)
	assertEq(t, "OS.Name", r.OS.Name, "Android")
	assertEq(t, "OS.Version", r.OS.Version, "10")
}

func TestParse_InstagramAndroidEngine(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 14; SM-A546B Build/UP1A.231005.007; wv) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 " +
		"Chrome/131.0.6778.200 Mobile Safari/537.36 " +
		"Instagram 371.1.0.44.107 Android " +
		"(34/14; 450dpi; 1080x2177; samsung; SM-A546B; a54x; s5e8835; it_IT; UP1A.231005.007)"
	r := p.Parse(ua)

	assertEq(t, "Browser.Engine", r.Browser.Engine, "Blink")
}

func TestParse_InstagramIOSEngine(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/23D127 " +
		"Instagram 419.0.0.27.74 (iPhone17,2; iOS 26_3; en_US; en; " +
		"scale=3.00; 1320x2868; IABMV/1; 895010607) Safari/604.1"
	r := p.Parse(ua)

	assertEq(t, "Browser.Engine", r.Browser.Engine, "WebKit")
}

func TestParse_MetaIAB(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 10; 211033MI Build/QP1A.190711.020; wv) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 " +
		"Chrome/138.0.7204.179 Mobile Safari/537.36 MetaIAB"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Meta")
	assertEq(t, "Browser.Name", r.Browser.Name, "Chrome")
	assertEq(t, "OS.Name", r.OS.Name, "Android")
	assertEq(t, "Device.Type", r.Device.Type, "mobile")
}

func TestNewParser_Default(t *testing.T) {
	t.Parallel()

	p, err := herald.NewParser()
	if err != nil {
		t.Fatalf("NewParser(): %v", err)
	}

	r := p.Parse("Googlebot/2.1")
	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeBot)
}

func TestNewParser_WithOverrides(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	err := os.WriteFile(
		filepath.Join(dir, "android.json"),
		[]byte(`{"TEST-X1": {"brand": "TestBrand", "model": "TestPhone"}}`),
		0o600,
	)
	if err != nil {
		t.Fatalf("write override: %v", err)
	}

	p, err := herald.NewParser(herald.WithOverrides(dir))
	if err != nil {
		t.Fatalf("NewParser(WithOverrides): %v", err)
	}

	// Override model resolves.
	ua := "Mozilla/5.0 (Linux; Android 14; TEST-X1) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/120.0.0.0 Mobile Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "Device.Model", r.Device.Model, "TestBrand TestPhone")
	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "TEST-X1")

	// Built-in data still works.
	r2 := p.Parse("Googlebot/2.1")
	assertEq(t, "Bot.ClientType", r2.ClientType, herald.ClientTypeBot)
}

func TestParse_EdgeIOS(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"Version/17.0 EdgiOS/120.0.0.0 Mobile/15E148 Safari/604.1"
	r := p.Parse(ua)

	assertEq(t, "Browser.Name", r.Browser.Name, "Edge")
	assertEq(t, "Browser.Version", r.Browser.Version, "120.0.0.0")
	assertEq(t, "Browser.Engine", r.Browser.Engine, "WebKit")
}

func TestParse_FacebookAndroidFBIAB(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 15; SM-A055F Build/AP3A.240905.015.A2) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Version/4.0 Chrome/150.0.7871.124 Mobile Safari/537.36 " +
		"[FB_IAB/FB4A;FBAV/567.1.0.52.74;IABMV/1;]"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Facebook")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "567.1.0.52.74")
	assertEq(t, "Browser.Name", r.Browser.Name, "Facebook")
	assertEq(t, "OS.Version", r.OS.Version, "15")
	assertEq(t, "Device.Model", r.Device.Model, "Samsung Galaxy A05")
}

func TestParse_MessengerAndroid(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 6.0; ALE-L21 Build/HuaweiALE-L21; wv) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Version/4.0 Chrome/68.0.3440.91 Mobile Safari/537.36 " +
		"[FB_IAB/Orca-Android;FBAV/181.0.0.12.78;]"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Messenger")
	assertEq(t, "Browser.Name", r.Browser.Name, "Messenger")
}

func TestParse_FacebookLite(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 10; POT-LX1 Build/HUAWEIPOT-L21; wv) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Version/4.0 Chrome/120.0.0.0 Mobile Safari/537.36 " +
		"[FBAN/EMA;FBLC/nl_NL;FBAV/218.0.0.6.119;]"
	r := p.Parse(ua)

	assertEq(t, "IAB.App", r.IAB.App, "Facebook Lite")
	assertEq(t, "IAB.Locale", r.IAB.Locale, "nl_NL")
}

func TestParse_TikTokIOSUnderscore(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2_1 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"Mobile/15E148 BytedanceWebview/d8a21c6 musical_ly_32.8.0 JsSdk/2.0 " +
		"NetType/WIFI Channel/App ByteLocale/ar Region/SA"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "TikTok")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "32.8.0")
	assertEq(t, "IAB.Locale", r.IAB.Locale, "ar")
	assertEq(t, "IAB.Region", r.IAB.Region, "SA")
	assertEq(t, "Browser.Name", r.Browser.Name, "TikTok")
}

func TestParse_TikTokAndroidTrill(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 11; SM-A155M Build/RP1A.200720.012; wv) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Version/4.0 Chrome/91.0.4472.120 Mobile Safari/537.36 " +
		"trill_2023103040 JsSdk/1.0 NetType/WIFI Channel/googleplay " +
		"AppName/musical_ly app_version/31.3.4 ByteLocale/it-IT " +
		"ByteFullLocale/it-IT Region/IT BytedanceWebview/d8a21c6"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "TikTok")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "31.3.4")
	assertEq(t, "IAB.Locale", r.IAB.Locale, "it-IT")
}

func TestParse_InstagramAndroidFrozenBase(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 10; K; wv) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Version/4.0 Chrome/150.0.0.0 Mobile Safari/537.36 " +
		"Instagram 439.0.0.37.89 Android " +
		"(35/15; 450dpi; 1080x2400; Xiaomi/Redmi; 24094RAD4G; citrine; mt6855; it_IT; 1021815691; IABMV/1)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Instagram")
	assertEq(t, "OS.Version", r.OS.Version, "15")
	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "24094RAD4G")
	assertEq(t, "IAB.Locale", r.IAB.Locale, "it_IT")
}

func TestParse_ReducedAndroidChrome(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 10; K) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/151.0.0.0 Mobile Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "OS.Name", r.OS.Name, "Android")
	assertEq(t, "OS.Version", r.OS.Version, "")
	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "")
	assertEq(t, "Device.Type", r.Device.Type, "mobile")
}

func TestParse_ReducedAndroidTablet(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 10; K) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/151.0.0.0 Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "Device.Type", r.Device.Type, "tablet")
}

func TestParse_FirefoxAndroid(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("Mozilla/5.0 (Android 13; Mobile; rv:109.0) Gecko/109.0 Firefox/117.0")

	assertEq(t, "Browser.Name", r.Browser.Name, "Firefox")
	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "")
	assertEq(t, "Device.Type", r.Device.Type, "mobile")
}

func TestParse_ChromeOS(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (X11; CrOS x86_64 14541.0.0) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/141.0.0.0 Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "OS.Name", r.OS.Name, "Chrome OS")
	assertEq(t, "Device.Type", r.Device.Type, "desktop")
	assertEq(t, "Browser.Name", r.Browser.Name, "Chrome")
}

func TestParse_SafariIOS26FrozenUA(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"Version/26.0 Mobile/15E148 Safari/604.1"
	r := p.Parse(ua)

	assertEq(t, "OS.Name", r.OS.Name, "iOS")
	assertEq(t, "OS.Version", r.OS.Version, "26.0")
	assertEq(t, "Browser.Name", r.Browser.Name, "Safari")
}

func TestParse_SnapchatIAB(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 14; moto g play - 2024 Build/U1UD34M.16-56-2; wv) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Version/4.0 Chrome/131.0.6778.39 Mobile Safari/537.36 Snapchat/13.17.0.42"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Snapchat")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "13.17.0.42")
}

func TestParse_SnapchatNativeApp(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("Snapchat/11.1.1.66 (Pixel 3 XL; Android 11; gzip)")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeNativeApp)
	assertEq(t, "Native.Name", r.Native.Name, "Snapchat")
}

func TestParse_TelegramAndroid(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 14; K) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/134.0.6998.135 Mobile Safari/537.36 " +
		"Telegram-Android/11.9.0 (Samsung SM-A155M; Android 14; SDK 34; AVERAGE)"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Telegram")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "11.9.0")
	assertEq(t, "OS.Version", r.OS.Version, "14")
	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "SM-A155M")
}

func TestParse_LineAndroid(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 11; SM-G991B) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/95.0.4638.74 Mobile Safari/537.36 Line/11.5.2/IAB"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "LINE")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "11.5.2")
}

func TestParse_WeChat(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 12; ALN-AL00 Build/HUAWEIALN-AL00; wv) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Version/4.0 Chrome/134.0.6998.136 Mobile Safari/537.36 " +
		"XWEB/1340043 MMWEBSDK/20250201 MMWEBID/783 " +
		"MicroMessenger/8.0.57.2820(0x2800393F) WeChat/arm64 Weixin " +
		"NetType/WIFI Language/zh_CN ABI/arm64"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "WeChat")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "8.0.57.2820")
	assertEq(t, "IAB.Locale", r.IAB.Locale, "zh_CN")
	assertEq(t, "IAB.NetType", r.IAB.NetType, "WIFI")
}

func TestParse_TwitterIOS(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 26_5 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"Mobile/23F77 Twitter for iPhone/12.5"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Twitter")
	assertEq(t, "IAB.AppVersion", r.IAB.AppVersion, "12.5")
	assertEq(t, "Browser.Name", r.Browser.Name, "Twitter")
	assertEq(t, "Browser.Engine", r.Browser.Engine, "WebKit")
}

func TestParse_TwitterAndroid(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 14; moto g stylus 5G - 2024) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Version/4.0 Chrome/130.0.0.0 Mobile Safari/537.36 TwitterAndroid"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Twitter")
}

func TestParse_GSAiOS(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 18_1 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"GSA/355.0.755551703 Mobile/15E148 Safari/604.1"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Google App")
	assertEq(t, "Browser.Name", r.Browser.Name, "Google App")
}

func TestParse_PinterestIOSBlock(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"Mobile/15E148 Safari/604.1 [Pinterest/iOS]"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Pinterest")
}

func TestParse_PinterestAndroidNative(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("Pinterest for Android/10.9.3 (Infinix-X689F; 11)")

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeNativeApp)
	assertEq(t, "Native.Name", r.Native.Name, "Pinterest")
	assertEq(t, "Native.Version", r.Native.Version, "10.9.3")
}

func TestParse_MetaIABMVFallback(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 10; K; wv) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Version/4.0 Chrome/150.0.0.0 Mobile Safari/537.36 IABMV/1"
	r := p.Parse(ua)

	assertEq(t, "ClientType", r.ClientType, herald.ClientTypeIAB)
	assertEq(t, "IAB.App", r.IAB.App, "Meta")
}

func TestParse_MiBrowser(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 15; POCO M6 Build/AP3A.240905.015.A2) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/123.0.6312.118 Mobile Safari/537.36 XiaoMi/MiuiBrowser/14.44.1-gn"
	r := p.Parse(ua)

	assertEq(t, "Browser.Name", r.Browser.Name, "Mi Browser")
	assertEq(t, "Browser.Version", r.Browser.Version, "14.44.1-gn")
}

func TestParse_UCBrowser(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; U; Android 13; en-US; SM-A155F Build/TP1A.220624.014) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Version/4.0 Chrome/100.0.4896.58 UCBrowser/13.4.0.1306 Mobile Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "Browser.Name", r.Browser.Name, "UC Browser")
}

func TestParse_HuaweiBrowser(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 10; HarmonyOS; ELS-NX9; HMSCore 6.1.0.314) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/88.0.4324.93 HuaweiBrowser/11.1.5.310 Mobile Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "Browser.Name", r.Browser.Name, "Huawei Browser")
}

func TestParse_OperaMiniPresto(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("Opera/9.80 (Android; Opera Mini/33.0.2254/174.101; U; en) Presto/2.12.423 Version/12.16")

	assertEq(t, "Browser.Name", r.Browser.Name, "Opera Mini")
	assertEq(t, "Browser.Version", r.Browser.Version, "33.0.2254")
	assertEq(t, "Browser.Engine", r.Browser.Engine, herald.EnginePresto)
}

func TestParse_OperaMiniIOS(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"OPiOS/16.0.15.124050 Mobile/15E148 Safari/9537.53"
	r := p.Parse(ua)

	assertEq(t, "Browser.Name", r.Browser.Name, "Opera Mini")
}

func TestParse_DuckDuckGoAndroid(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 14; SM-S911B) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/147.0.7727.111 Mobile DuckDuckGo/5 Safari/537.36"
	r := p.Parse(ua)

	assertEq(t, "Browser.Name", r.Browser.Name, "DuckDuckGo")
	assertEq(t, "Browser.Engine", r.Browser.Engine, "Blink")
}

func TestParse_OperaGXMobile(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (Linux; Android 14; SM-S911B) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/150.0.7871.46 Mobile Safari/537.36 OPX/3.1"
	r := p.Parse(ua)

	assertEq(t, "Browser.Name", r.Browser.Name, "Opera GX")
}

func TestParseWithHints_ModelResolution(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)

	hints := herald.ClientHints{
		UA:       `"Chromium";v="151", "Google Chrome";v="151"`,
		Platform: "Android",
		Mobile:   true,
		Model:    "SM-S938B",
	}
	ua := "Mozilla/5.0 (Linux; Android 10; K) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/151.0.0.0 Mobile Safari/537.36"
	r := p.ParseWithHints(ua, hints)

	assertEq(t, "Device.ModelRaw", r.Device.ModelRaw, "SM-S938B")
	assertEq(t, "Device.Model", r.Device.Model, "Samsung Galaxy S25 Ultra")
}

func TestParseWithHints_AndroidTablet(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)

	hints := herald.ClientHints{
		UA:       `"Chromium";v="151", "Google Chrome";v="151"`,
		Platform: "Android",
		Mobile:   false,
	}
	ua := "Mozilla/5.0 (Linux; Android 10; K) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/151.0.0.0 Safari/537.36"
	r := p.ParseWithHints(ua, hints)

	assertEq(t, "Device.Type", r.Device.Type, "tablet")
}

func TestParseWithHints_Brave(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)

	hints := herald.ClientHints{
		UA: `"Chromium";v="136", "Brave";v="136", "Not.A/Brand";v="99"`,
	}
	r := p.ParseWithHints(uaChromeWin, hints)

	assertEq(t, "Browser.Name", r.Browser.Name, "Brave")
}

func TestParse_Darwin27(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	r := p.Parse("App/1.0 CFNetwork/3860.100.1 Darwin/27.0.0")

	assertEq(t, "OS.Name", r.OS.Name, "Darwin")
	assertEq(t, "OS.Version", r.OS.Version, "27")
}

func TestParse_VisionProFBDV(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 26_0 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"Mobile/15E148 [FBAN/FBIOS;FBAV/500.0.0.0.1;FBDV/RealityDevice14,1;FBSN/visionOS;FBSV/26.0]"
	r := p.Parse(ua)

	assertEq(t, "IAB.App", r.IAB.App, "Facebook")
	assertEq(t, "Device.Model", r.Device.Model, "Apple Vision Pro")
}

func TestParse_IPhone17eInstagram(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)
	ua := "Mozilla/5.0 (iPhone; CPU iPhone OS 26_1 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) " +
		"Mobile/15E148 Instagram 374.1.10.34.80 " +
		"(iPhone18,5; iOS 26_1; en_US; en-US; scale=3.00; 1179x2556; 715949514; IABMV/1)"
	r := p.Parse(ua)

	assertEq(t, "IAB.App", r.IAB.App, "Instagram")
	assertEq(t, "Device.Model", r.Device.Model, "iPhone 17e")
	assertEq(t, "OS.Version", r.OS.Version, "26.1")
}

func TestParse_NewBotPatterns(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)

	cases := []struct{ ua, name string }{
		{
			"Mozilla/5.0 AppleWebKit/537.36 " +
				"(KHTML, like Gecko; compatible; Claude-User/1.0; +claude-user@anthropic.com)",
			"Claude User",
		},
		{
			"Mozilla/5.0 AppleWebKit/537.36 " +
				"(KHTML, like Gecko; compatible; Perplexity-User/1.0; +https://perplexity.ai/perplexity-user)",
			"Perplexity User",
		},
		{
			"Mozilla/5.0 (compatible; Amazonbot/0.1; +https://developer.amazon.com/support/amazonbot)",
			"Amazonbot",
		},
		{"WhatsApp/2.23.20.0", "WhatsApp Link Preview"},
	}

	for _, c := range cases {
		r := p.Parse(c.ua)

		assertEq(t, "ClientType "+c.name, r.ClientType, herald.ClientTypeBot)
		assertEq(t, "Bot.Name "+c.name, r.Bot.Name, c.name)
	}
}

func TestLookup_NewDeviceModels(t *testing.T) {
	t.Parallel()

	p := newTestParser(t)

	if name, ok := p.LookupAppleModel("iPhone18,5"); !ok || name != "iPhone 17e" {
		t.Errorf("iPhone18,5: got %q, %v", name, ok)
	}

	if dev, ok := p.LookupAndroidModel("SM-S931B"); !ok || dev.Model != "Galaxy S25" {
		t.Errorf("SM-S931B: got %+v, %v", dev, ok)
	}

	if dev, ok := p.LookupAndroidModel("2409BRN2CY"); !ok || dev.Model != "Redmi 14C" {
		t.Errorf("2409BRN2CY: got %+v, %v", dev, ok)
	}
}

func assertEq[T comparable](t *testing.T, field string, got, want T) {
	t.Helper()

	if got != want {
		t.Errorf("%s: got %v, want %v", field, got, want)
	}
}

func assertFloat(t *testing.T, field string, got, want float64) {
	t.Helper()

	if got != want {
		t.Errorf("%s: got %f, want %f", field, got, want)
	}
}
