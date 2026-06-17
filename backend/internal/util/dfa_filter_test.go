package util

import (
	"testing"
)

func TestDFAFilterContains(t *testing.T) {
	f := NewDFAFilter([]string{"傻逼", "垃圾", "fuck", "妈的"})

	tests := []struct {
		text     string
		expected bool
	}{
		{"你好", false},
		{"你是一个傻逼", true},
		{"这个垃圾东西", true},
		{"what the fuck is this", true},
		{"他妈的在这里", true},
		{"正常文本没有敏感词", false},
		{"", false},
		{"傻  逼", false}, // spaces break the match (DFA works on runes)
		{"傻逼!", true},    // punctuation after is fine
	}

	for _, tc := range tests {
		result := f.Contains(tc.text)
		if result != tc.expected {
			t.Errorf("Contains(%q) = %v, want %v", tc.text, result, tc.expected)
		}
	}
}

func TestDFAFilterReplace(t *testing.T) {
	f := NewDFAFilter([]string{"傻逼", "垃圾", "妈的"})

	tests := []struct {
		input    string
		expected string
	}{
		{"你好世界", "你好世界"},
		{"你是个傻逼吧", "你是个**吧"},
		{"这个垃圾东西", "这个**东西"},
		{"他妈的在这", "他**在这"},
		{"", ""},
	}

	for _, tc := range tests {
		result := f.Replace(tc.input)
		if result != tc.expected {
			t.Errorf("Replace(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestDFAFilterFind(t *testing.T) {
	f := NewDFAFilter([]string{"傻逼", "色情", "赌博"})

	if word := f.Find("你好世界"); word != "" {
		t.Errorf("Find should return empty, got %q", word)
	}
	if word := f.Find("这是色情内容"); word != "色情" {
		t.Errorf("Find should return '色情', got %q", word)
	}
	if word := f.Find("涉及赌博广告"); word != "赌博" {
		t.Errorf("Find should return '赌博', got %q", word)
	}
}

func TestDFAFilterOverlappingWords(t *testing.T) {
	// Test words that share prefixes
	f := NewDFAFilter([]string{"赌博", "赌场", "赌博网站"})

	if !f.Contains("这里有赌博") {
		t.Error("should detect 赌博")
	}
	if !f.Contains("澳门赌场") {
		t.Error("should detect 赌场")
	}
	if !f.Contains("赌博网站已被查封") {
		t.Error("should detect 赌博网站")
	}

	// Replace should catch the longest match (赌博网站 = 4 chars)
	result := f.Replace("赌博网站真好")
	if result != "****真好" {
		t.Errorf("Replace overlapping: %q", result)
	}
}

func TestBuildDFAFilter(t *testing.T) {
	// Test the BuildDFAFilter function from sensitive_words.go
	f := BuildDFAFilter()
	if f == nil {
		t.Fatal("BuildDFAFilter should return non-nil")
	}
	// Should not panic on common text
	_ = f.Contains("你好")
	_ = f.Replace("正常聊天内容")
	_ = f.Find("正常")
}
