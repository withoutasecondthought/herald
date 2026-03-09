package herald_test

import "testing"

func BenchmarkParse_ChromeDesktop(b *testing.B) {
	p := newTestParser(&testing.T{})
	ua := uaChromeWin

	b.ResetTimer()

	for b.Loop() {
		p.Parse(ua)
	}
}

func BenchmarkParse_SafariIOS(b *testing.B) {
	p := newTestParser(&testing.T{})
	ua := uaSafariIOS

	b.ResetTimer()

	for b.Loop() {
		p.Parse(ua)
	}
}

func BenchmarkParse_Bot(b *testing.B) {
	p := newTestParser(&testing.T{})
	ua := "Mozilla/5.0 (compatible; Googlebot/2.1; " +
		"+http://www.google.com/bot.html)"

	b.ResetTimer()

	for b.Loop() {
		p.Parse(ua)
	}
}

func BenchmarkParse_CFNetwork(b *testing.B) {
	p := newTestParser(&testing.T{})
	ua := "ut-1/3 CFNetwork/3860.400.51 Darwin/25.3.0"

	b.ResetTimer()

	for b.Loop() {
		p.Parse(ua)
	}
}

func BenchmarkParse_Empty(b *testing.B) {
	p := newTestParser(&testing.T{})

	for b.Loop() {
		p.Parse("")
	}
}

func BenchmarkDetectType_Chrome(b *testing.B) {
	p := newTestParser(&testing.T{})
	ua := uaChromeWin

	b.ResetTimer()

	for b.Loop() {
		p.DetectType(ua)
	}
}

func BenchmarkDetectType_Bot(b *testing.B) {
	p := newTestParser(&testing.T{})
	ua := "Googlebot/2.1"

	b.ResetTimer()

	for b.Loop() {
		p.DetectType(ua)
	}
}
