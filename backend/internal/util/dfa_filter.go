package util

import "strings"

// DFAFilter is a DFA-based sensitive word filter.
type DFAFilter struct {
	root *dfaNode
}

type dfaNode struct {
	children map[rune]*dfaNode
	isEnd    bool
}

func NewDFAFilter(words []string) *DFAFilter {
	f := &DFAFilter{root: &dfaNode{children: make(map[rune]*dfaNode)}}
	for _, w := range words {
		f.AddWord(w)
	}
	return f
}

func (f *DFAFilter) AddWord(word string) {
	node := f.root
	for _, r := range word {
		if node.children[r] == nil {
			node.children[r] = &dfaNode{children: make(map[rune]*dfaNode)}
		}
		node = node.children[r]
	}
	node.isEnd = true
}

// Contains returns true if text contains any sensitive word.
func (f *DFAFilter) Contains(text string) bool {
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		node := f.root
		for j := i; j < len(runes); j++ {
			next := node.children[runes[j]]
			if next == nil {
				break
			}
			if next.isEnd {
				return true
			}
			node = next
		}
	}
	return false
}

// Replace replaces all sensitive words with '*'.
func (f *DFAFilter) Replace(text string) string {
	runes := []rune(text)
	mask := make([]bool, len(runes))

	for i := 0; i < len(runes); i++ {
		node := f.root
		endPos := -1
		for j := i; j < len(runes); j++ {
			next := node.children[runes[j]]
			if next == nil {
				break
			}
			if next.isEnd {
				endPos = j
			}
			node = next
		}
		if endPos >= 0 {
			for k := i; k <= endPos; k++ {
				mask[k] = true
			}
		}
	}

	var result strings.Builder
	for i, r := range runes {
		if mask[i] {
			result.WriteRune('*')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Find returns the first sensitive word found, or empty string.
func (f *DFAFilter) Find(text string) string {
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		node := f.root
		for j := i; j < len(runes); j++ {
			next := node.children[runes[j]]
			if next == nil {
				break
			}
			if next.isEnd {
				return string(runes[i : j+1])
			}
			node = next
		}
	}
	return ""
}
