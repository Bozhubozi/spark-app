package service

import (
	"testing"
)

func TestZodiacCompatibility(t *testing.T) {
	svc := NewZodiacService()

	// Perfect match (same element, same triplicity)
	score1 := svc.Compatibility("狮子座", "白羊座") // both 火象
	if score1 < 70 {
		t.Errorf("狮子+白羊 should be high compatibility, got %d", score1)
	}

	// Good match
	score2 := svc.Compatibility("巨蟹座", "天蝎座") // both 水象
	if score2 < 70 {
		t.Errorf("巨蟹+天蝎 should be high, got %d", score2)
	}

	// Challenging match
	score3 := svc.Compatibility("狮子座", "金牛座") // 火象 vs 土象
	if score3 > 60 {
		t.Errorf("狮子+金牛 should be lower, got %d", score3)
	}

	// All scores should be in range 0-100
	for _, a := range []string{"白羊座", "金牛座", "双子座", "巨蟹座", "狮子座", "处女座"} {
		for _, b := range []string{"天秤座", "天蝎座", "射手座", "摩羯座", "水瓶座", "双鱼座"} {
			s := svc.Compatibility(a, b)
			if s < 0 || s > 100 {
				t.Errorf("Compatibility(%s, %s) = %d, out of range", a, b, s)
			}
			// Should be symmetric
			s2 := svc.Compatibility(b, a)
			if s != s2 {
				t.Errorf("Compatibility not symmetric: %s+%s=%d, %s+%s=%d", a, b, s, b, a, s2)
			}
		}
	}
}

func TestZodiacReport(t *testing.T) {
	svc := NewZodiacService()

	report := svc.Report("狮子座", "天秤座")
	if report == "" {
		t.Error("report should not be empty")
	}
	if len(report) < 20 {
		t.Errorf("report too short: %q", report)
	}
	t.Logf("Report: %s", report)
}

func TestZodiacIndex(t *testing.T) {
	svc := NewZodiacService()
	// Verify all 12 signs are valid (Compatibility uses indexOf internally)
	signs := []string{"白羊座", "金牛座", "双子座", "巨蟹座", "狮子座", "处女座",
		"天秤座", "天蝎座", "射手座", "摩羯座", "水瓶座", "双鱼座"}
	for _, s := range signs {
		score := svc.Compatibility(s, s)
		if score < 80 {
			t.Errorf("same sign (%s) should be >= 80, got %d", s, score)
		}
	}
}
