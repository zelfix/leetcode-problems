// Valid Palindrome II [Easy] — Two Pointers
// https://leetcode.com/problems/valid-palindrome-ii/
//
// Submit with:  neet submit valid-palindrome-ii

package main

func isPalin(s string, lo int, high int) bool {
	i := lo
	j := high
	for i < j {
		if s[i] != s[j] {
			return false
		}
		i++
		j--
	}
	return true
}

func validPalindrome(s string) bool {
	i := 0
	j := len(s) - 1
	for i < j {
		l := rune(s[i])
		r := rune(s[j])
		if l != r {
			return isPalin(s, i+1, j) || isPalin(s, i, j-1)
		}
		i++
		j--
	}
	return true
}
