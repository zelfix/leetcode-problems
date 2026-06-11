// Implement Trie (Prefix Tree) [Medium] — Tries
// https://leetcode.com/problems/implement-trie-prefix-tree/
//
// Submit with:  neet submit implement-trie-prefix-tree

package main

type Trie struct {
	branches [26]*Trie
	isWord   bool
}

func Constructor() Trie {
	return Trie{}
}

func (this *Trie) Insert(word string) {
	if len(word) == 0 {
		this.isWord = true
		return
	}
	if this.branches[word[0]-'a'] == nil {
		this.branches[word[0]-'a'] = &Trie{}
	}
	this.branches[word[0]-'a'].Insert(word[1:])
}

func (this *Trie) Search(word string) bool {
	if len(word) == 0 {
		return this.isWord
	}
	if this.branches[word[0]-'a'] == nil {
		return false
	}
	return this.branches[word[0]-'a'].Search(word[1:])
}

func (this *Trie) StartsWith(prefix string) bool {
	if len(prefix) == 0 {
		return true
	}
	if this.branches[prefix[0]-'a'] == nil {
		return false
	}
	return this.branches[prefix[0]-'a'].StartsWith(prefix[1:])
}

/**
 * Your Trie object will be instantiated and called as such:
 * obj := Constructor();
 * obj.Insert(word);
 * param_2 := obj.Search(word);
 * param_3 := obj.StartsWith(prefix);
 */
