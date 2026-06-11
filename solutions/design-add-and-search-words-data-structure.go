// Design Add and Search Words Data Structure [Medium] — Tries
// https://leetcode.com/problems/design-add-and-search-words-data-structure/
//
// Submit with:  neet submit design-add-and-search-words-data-structure

package main

type WordDictionary struct {
	branches [26]*WordDictionary
	isWord   bool
}

func Constructor() WordDictionary {
	return WordDictionary{}
}

func (this *WordDictionary) AddWord(word string) {
	if len(word) == 0 {
		this.isWord = true
		return
	}
	if this.branches[word[0]-'a'] == nil {
		this.branches[word[0]-'a'] = &WordDictionary{}
	}
	this.branches[word[0]-'a'].AddWord(word[1:])
}

func (this *WordDictionary) Search(word string) bool {
	if len(word) == 0 {
		return this.isWord
	}
	if word[0] == '.' {
		for _, w := range this.branches {
			if w != nil {
				if w.Search(word[1:]) {
					return true
				}
			}
		}
		return false
	}
	if this.branches[word[0]-'a'] == nil {
		return false
	}
	return this.branches[word[0]-'a'].Search(word[1:])
}

/**
 * Your WordDictionary object will be instantiated and called as such:
 * obj := Constructor();
 * obj.AddWord(word);
 * param_2 := obj.Search(word);
 */
