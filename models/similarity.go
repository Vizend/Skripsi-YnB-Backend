package models

import (
	"strings"
)

// Levenshtein distance 
func levenshtein(a, b string) int {
	ra := []rune(strings.ToLower(strings.TrimSpace(a)))
	rb := []rune(strings.ToLower(strings.TrimSpace(b)))
	da := make([]int, len(rb)+1)
	db := make([]int, len(rb)+1)
	for j := 0; j <= len(rb); j++ {
		da[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		db[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			db[j] = min3(db[j-1]+1, da[j]+1, da[j-1]+cost)
		}
		copy(da, db)
	}
	return da[len(rb)]
}
func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
func simRatio(a, b string) float64 {
	maxLen := float64(max(len([]rune(a)), len([]rune(b))))
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(levenshtein(a, b))/maxLen
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
