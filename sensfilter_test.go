package sensfilter

import (
	"fmt"
	"math/rand"
	"testing"
	"unicode/utf8"
)

type myWriter struct {
	list []*Result
}

func (_this *myWriter) Write(result2 *Result) (stop bool) {
	_this.list = append(_this.list, result2)
	return false
}

func (_this *myWriter) Len() int {
	return len(_this.list)
}

func (_this *Search) findB(s string, w ResultWriter) {
	n := len(s)
	trieRoot := _this.trieWriter.trie()
	skipper := _this.trieWriter.Skip()
next:
	for i := 0; i < n; {
		v, l := decodeStr(s[i:])
		// 如果该节点不存在，则跳过
		node, ok := trieRoot.next[v]
		if !ok {
			i += l
			continue
		}

		j := i
		var word []rune
		for {
			word = append(word, v)
			if node.end {
				res := &Result{string(word), s[i : j+l], i, j + l}
				stop := w.Write(res)
				if stop {
					return
				}
			}

			j += l
			// 跳过一些无意义的字符
			for {
				v, l = decodeStr(s[j:])
				if j < n && skipper.ShouldSkip(v) {
					j += l
				} else {
					break
				}
			}

			node = node.next[v]
			// 如果找不到直接break
			if node == nil {
				i = j
				continue next
			}
		}
	}

	return
}

func BenchmarkA(b *testing.B) {
	words := []string{"林茹", "林如", "临蓐", "空子", "霸王龙", "我是个SB", "是我"}
	obj := Strings(words)
	str := []byte("我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的")
	var res []*Result
	for i := 0; i < b.N; i++ {
		res = obj.findByAC(str, false)
	}
	fmt.Println(res)
}

func BenchmarkB(b *testing.B) {
	words := []string{"林茹", "林如", "临蓐", "空子", "霸王龙", "我是个SB", "是我"}
	obj := Strings(words)
	str := "我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的"
	for i := 0; i < b.N; i++ {
		mw := &myWriter{}
		obj.find([]byte(str), mw)
	}
	mw := &myWriter{}
	obj.find([]byte(str), mw)
	fmt.Println(mw.list)
}

func BenchmarkC(b *testing.B) {
	words := []string{"林茹", "林如", "临蓐", "空子", "霸王龙", "我是个SB", "是我"}
	obj := Strings(words)
	str := "我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的"
	for i := 0; i < b.N; i++ {
		mw := &myWriter{}
		obj.findB(str, mw)
	}
	mw := &myWriter{}
	obj.findB(string(str), mw)
	fmt.Println(mw.list)
}

func BenchmarkD(b *testing.B) {
	m := make(map[rune]struct{}, len(sortedSkipList))
	for _, v := range sortedSkipList {
		m[v] = struct{}{}
	}
	in := false
	for i := 0; i < b.N; i++ {
		val := rand.Intn(utf8.MaxRune)
		_, in = m[rune(val)]
	}
	fmt.Println(in)
}

func BenchmarkE(b *testing.B) {
	in := false
	skipper := &Skip{[]rune(sortedSkipList)}
	for i := 0; i < b.N; i++ {
		val := rand.Intn(utf8.MaxRune)
		in = skipper.ShouldSkip(rune(val))
	}
	fmt.Println(in)
}

func TestSearch_Find(t *testing.T) {
	words := []string{"dog", "cat", "apple", "orange", "chicken", "鸭子", "水果", "敏感词"}
	obj := Strings(words)

	str := "This is a sentence containing sensitive words such as dog, cat, and duck (鸭子 in Chinese)."

	res := obj.Find([]byte(str))

	type wantPair struct {
		word    string
		matched string
	}

	wants := []wantPair{
		{"dog", "dog"},
		{"cat", "cat"},
		{"鸭子", "鸭子"},
	}

	if len(wants) != len(res) {
		t.Fatalf("Incorrect number of matched sensitive words.")
	}

	for i, v := range res {
		want := wants[i]
		if v.Word != want.word || v.Matched != want.matched {
			t.Fatalf("Failed to match sensitive word: %s", want.word)
		}
	}
}

func TestSearch_Find2(t *testing.T) {
	words := []string{"abcef", "bcde", "bcd", "abcde"}
	obj := Strings(words)

	str := "#bc#d##abc*deff;;abcde"
	res := obj.Find([]byte(str))
	type wantPair struct {
		word    string
		matched string
	}

	wants := []wantPair{
		{"bcd", "bc#d"},
		{"abcde", "abc*de"},
		{"abcde", "abcde"},
	}

	for i, v := range res {
		want := wants[i]
		if v.Word != want.word || v.Matched != want.matched {
			t.Fatalf("Failed to match sensitive word: %s", want.word)
		}
	}
}

func TestSearch_Find3(t *testing.T) {
	words := []string{"林茹", "林如", "临蓐", "空子", "霸王龙", "我是个SB", "是我", "abcd", "bc"}
	obj := Strings(words)
	str := []byte("我空ss子sss我是霸**王*龙,我是我我是个(S)(B)真的abcccc")
	res := obj.Find(str)

	type wantPair struct {
		word    string
		matched string
	}

	wants := []wantPair{
		{"霸王龙", "霸**王*龙"},
		{"是我", "是我"},
		{"我是个SB", "我是个(S)(B"},
		{"bc", "bc"},
	}

	for i, v := range res {
		want := wants[i]
		if v.Word != want.word || v.Matched != want.matched {
			t.Fatalf("Failed to match sensitive word: %s", want.word)
		}
	}
}

func TestSearch_Replace(t *testing.T) {
	words := []string{"dog", "cat", "apple", "orange", "chicken", "鸭子", "水果", "敏感词"}
	obj := Strings(words)

	input := "I have a dog and a cat, and I love eating apples and oranges. I also like to eat chicken and duck (鸭子)."
	expectedOutput := "I have a *** and a ***, and I love eating *****s and ******s. I also like to eat ******* and duck (******)."

	output := obj.Replace([]byte(input), '*')

	if string(output) != expectedOutput {
		t.Fatalf("Unexpected output. Expected: %s. Got: %s.", expectedOutput, output)
	}
}

func TestSearch_HasSens(t *testing.T) {
	words := []string{"dog", "cat", "apple", "orange", "chicken", "鸭子", "水果", "敏感词"}
	obj := Strings(words)

	input1 := "This sentence does not contain any sensitive words."
	input2 := "This sentence contains the word dog."
	input3 := "这句话包含敏感词鸭子。"

	if obj.HasSens([]byte(input1)) {
		t.Fatalf("Expected no sensitive words, but found some.")
	}

	if !obj.HasSens([]byte(input2)) {
		t.Fatalf("Expected to find the word dog, but did not find it.")
	}

	if !obj.HasSens([]byte(input3)) {
		t.Fatalf("Expected to find the sensitive word 鸭子, but did not find it.")
	}
}

func TestSearch_FindWithSkip(t *testing.T) {
	words := []string{"TMD"}
	obj := Strings(words, "!*")

	str := "T***MD;T*M**D;T!MD;T#M#D"
	res := obj.Find([]byte(str))

	type wantPair struct {
		word    string
		matched string
	}

	wants := []wantPair{
		{"TMD", "T***MD"},
		{"TMD", "T*M**D"},
		{"TMD", "T!MD"},
	}

	if len(wants) != len(res) {
		t.Fatalf("The number of matched sensitive words is incorrect.wants len:%d,result len:%d", len(wants), len(res))
	}

	for i, v := range res {
		want := wants[i]
		if v.Word != want.word || v.Matched != want.matched {
			t.Fatalf("Unable to match sensitive word：%s", want.word)
		}
	}
}
