// Merge Strings Alternately [Easy] — Two Pointers
// https://leetcode.com/problems/merge-strings-alternately/
//
// Submit with:  neet submit merge-strings-alternately

package main

import "strings"

func mergeAlternately(word1 string, word2 string) string {
	i := 0
	j := 0
	var sb strings.Builder
	for i < len(word1) && j < len(word2) {
		sb.WriteByte(word1[i])
		i++
		sb.WriteByte(word2[j])
		j++
	}
	if i == len(word1) {
		sb.WriteString(word2[j:])
	}
	if j == len(word2) {
		sb.WriteString(word1[i:])
	}
	return sb.String()
}
