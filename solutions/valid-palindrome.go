// Valid Palindrome [Easy] — Two Pointers
// https://leetcode.com/problems/valid-palindrome/
//
// Submit with:  neet submit valid-palindrome

package main

import "unicode"

func isPalindrome(s string) bool {
	i := 0
	j := len(s) - 1
	for i < j {
		l := rune(s[i])
		isAlnumL := unicode.IsLetter(l) || unicode.IsNumber(l)
		if !isAlnumL {
			i++
			continue
		}

		r := rune(s[j])
		isAlnumR := unicode.IsLetter(r) || unicode.IsNumber(r)
		if !isAlnumR {
			j--
			continue
		}
		if unicode.ToLower(rune(s[i])) != unicode.ToLower(rune(s[j])) {
			return false
		}
		i++
		j--
	}
	return true
}
