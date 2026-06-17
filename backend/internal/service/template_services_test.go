package service

import (
	"strings"
	"testing"

	"github.com/spark-app/backend/internal/model"
)

func TestPersonalityReport(t *testing.T) {
	svc := NewPersonalityReportService(nil)

	dims := []model.PersonalityDimension{
		{Dimension: "extraversion", Score: 4.5},
		{Dimension: "agreeableness", Score: 3.2},
		{Dimension: "conscientiousness", Score: 4.0},
		{Dimension: "neuroticism", Score: 2.0},
		{Dimension: "openness", Score: 4.5},
	}

	report := svc.Generate(dims)

	if report.Title == "" {
		t.Error("title should not be empty")
	}
	if report.Summary == "" {
		t.Error("summary should not be empty")
	}
	if len(report.Traits) == 0 {
		t.Error("traits should not be empty")
	}
	if report.Advice == "" {
		t.Error("advice should not be empty")
	}

	// Verify all 5 dimensions are mentioned in summary
	for _, dim := range []string{"社交", "与人相处", "做事风格", "情绪", "新事物"} {
		if !strings.Contains(report.Summary, dim) {
			t.Errorf("summary should mention %s", dim)
		}
	}

	t.Logf("Title: %s", report.Title)
	t.Logf("Summary: %s", report.Summary)
	t.Logf("Traits: %v", report.Traits)
}

func TestPersonalityReportDescribe(t *testing.T) {
	svc := NewPersonalityReportService(nil)

	tests := []struct {
		dim    string
		score  float64
		expect string
	}{
		{"extraversion", 4.5, "外向活跃"},
		{"extraversion", 3.0, "适度外向"},
		{"extraversion", 2.0, "偏内向"},
		{"extraversion", 1.0, "内向沉静"},
		{"agreeableness", 4.5, "高度共情"},
		{"agreeableness", 2.0, "独立自主"},
		{"conscientiousness", 5.0, "高度自律"},
		{"conscientiousness", 1.0, "自由奔放"},
		{"neuroticism", 4.5, "敏感细腻"},
		{"neuroticism", 1.0, "淡定从容"},
		{"openness", 4.5, "好奇心强"},
		{"openness", 1.5, "追求确定"},
		{"unknown", 3.0, "均衡"},
	}

	for _, tc := range tests {
		result := svc.describe(tc.dim, tc.score)
		if result != tc.expect {
			t.Errorf("describe(%s, %.1f) = %q, want %q", tc.dim, tc.score, result, tc.expect)
		}
	}
}

func TestHoroscopeDaily(t *testing.T) {
	svc := NewHoroscopeService()

	dims := []model.PersonalityDimension{
		{Dimension: "extraversion", Score: 4.5},
		{Dimension: "agreeableness", Score: 3.5},
	}

	// Test all 12 zodiac signs produce non-empty output
	for _, sign := range []string{
		"水瓶座", "双鱼座", "白羊座", "金牛座", "双子座", "巨蟹座",
		"狮子座", "处女座", "天秤座", "天蝎座", "射手座", "摩羯座",
	} {
		result := svc.Daily(sign, dims)
		if result == "" {
			t.Errorf("Daily(%s) returned empty", sign)
		}
		if len(result) < 20 {
			t.Errorf("Daily(%s) too short: %q", sign, result)
		}
	}

	// Unknown zodiac should return fallback
	fallback := svc.Daily("未知", dims)
	if fallback == "" {
		t.Error("fallback should not be empty")
	}
}

func TestHoroscopeArchetypeClassification(t *testing.T) {
	tests := []struct {
		name string
		dims []model.PersonalityDimension
		arch archetype
	}{
		{
			"social",
			[]model.PersonalityDimension{
				{Dimension: "extraversion", Score: 4.5},
				{Dimension: "agreeableness", Score: 4.0},
			},
			archSocial,
		},
		{
			"outgoing",
			[]model.PersonalityDimension{
				{Dimension: "extraversion", Score: 4.5},
				{Dimension: "agreeableness", Score: 2.0},
			},
			archOutgoing,
		},
		{
			"artistic",
			[]model.PersonalityDimension{
				{Dimension: "extraversion", Score: 2.0},
				{Dimension: "openness", Score: 4.5},
				{Dimension: "neuroticism", Score: 3.5},
			},
			archArtistic,
		},
		{
			"pragmatic",
			[]model.PersonalityDimension{
				{Dimension: "extraversion", Score: 2.0},
				{Dimension: "conscientiousness", Score: 4.5},
				{Dimension: "neuroticism", Score: 2.0},
			},
			archPragmatic,
		},
		{
			"lowkey default",
			[]model.PersonalityDimension{
				{Dimension: "extraversion", Score: 2.0},
				{Dimension: "openness", Score: 2.0},
				{Dimension: "neuroticism", Score: 2.0},
				{Dimension: "conscientiousness", Score: 2.0},
			},
			archLowKey,
		},
	}

	for _, tc := range tests {
		result := classifyArchetype(tc.dims)
		if result != tc.arch {
			t.Errorf("%s: got arch %d, want %d", tc.name, result, tc.arch)
		}
	}
}

func TestIcebreaker(t *testing.T) {
	svc := NewIcebreakerService(NewZodiacService())

	userInterests := []model.InterestTag{
		{ID: 1, Name: "摄影", Category: "art"},
		{ID: 2, Name: "咖啡", Category: "lifestyle"},
	}
	targetInterests := []model.InterestTag{
		{ID: 2, Name: "咖啡", Category: "lifestyle"},
		{ID: 3, Name: "徒步", Category: "outdoor"},
	}
	personality := []model.PersonalityDimension{
		{Dimension: "extraversion", Score: 4.0},
	}

	icebreakers := svc.Generate("狮子座", "天秤座", userInterests, targetInterests, personality)

	if len(icebreakers) == 0 {
		t.Fatal("should generate at least 1 icebreaker")
	}
	for i, ib := range icebreakers {
		if ib == "" {
			t.Errorf("icebreaker %d is empty", i)
		}
		t.Logf("Icebreaker %d: %s", i+1, ib)
	}
}
