// Two Sum II - Input Array Is Sorted [Medium] — Two Pointers
// https://leetcode.com/problems/two-sum-ii-input-array-is-sorted/
//
// Submit with:  neet submit two-sum-ii-input-array-is-sorted

package main

func twoSum(numbers []int, target int) []int {
	i := 0
	j := len(numbers) - 1
	for i < j {
		sum := numbers[i] + numbers[j]
		if sum == target {
			return []int{i + 1, j + 1}
		} else if sum > target {
			j--
		} else {
			i++
		}
	}
	return []int{}
}
