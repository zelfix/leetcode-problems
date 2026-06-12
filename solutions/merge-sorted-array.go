// Merge Sorted Array [Easy] — Two Pointers
// https://leetcode.com/problems/merge-sorted-array/
//
// Submit with:  neet submit merge-sorted-array

package main

func merge(nums1 []int, m int, nums2 []int, n int) {
	i := m - 1
	j := n - 1
	k := m + n - 1
	for ; j >= 0; k-- {
		if i < 0 || nums1[i] < nums2[j] {
			nums1[k] = nums2[j]
			j--
		} else {
			nums1[k] = nums1[i]
			i--
		}
	}
}
