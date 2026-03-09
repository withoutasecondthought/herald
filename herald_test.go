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
	assertEq(t, "OS.Version", r.OS.Version, "10.15.7")
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
	assertEq(t, "Browser.Name", r.Browser.Name, "Google Chrome")
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
