// Valid Parentheses [Easy] — Stack
// https://leetcode.com/problems/valid-parentheses/
//
// Submit with:  neet submit valid-parentheses

package main

func isValid(s string) bool {
	var stack []rune
	var brackets map[rune]rune = map[rune]rune{
		')': '(',
		'}': '{',
		']': '[',
	}
	for _, ch := range s {
		if open, ok := brackets[ch]; ok {
			if len(stack) > 0 && stack[len(stack)-1] == open {
				stack = stack[:len(stack)-1]
			} else {
				return false
			}
		} else {
			stack = append(stack, ch)
		}
	}
	return len(stack) == 0
}
