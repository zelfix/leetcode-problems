// Baseball Game [Easy] — Stack
// https://leetcode.com/problems/baseball-game/
//
// Submit with:  neet submit baseball-game

package main

import "strconv"

func calPoints(operations []string) int {
	var stack []int
	for _, op := range operations {
		switch op {
		case "+":
			n := len(stack)
			sum := stack[n-1] + stack[n-2]
			stack = append(stack, sum)
		case "C":
			stack = stack[:len(stack)-1]
		case "D":
			stack = append(stack, stack[len(stack)-1]*2)
		default:
			num, _ := strconv.Atoi(op)
			stack = append(stack, num)
		}
	}
	sum := 0
	for _, n := range stack {
		sum += n
	}
	return sum
}
