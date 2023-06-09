package sensfilter

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
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

func (_this *Search) find(s []byte, w ResultWriter) {
	n := len(s)
	trieRoot := _this.trieWriter.trie()
	skipper := _this.trieWriter.Skip()
next:
	for i := 0; i < n; {
		v, l := decodeBytes(s[i:])
		// 如果该节点不存在，则跳过
		node, ok := trieRoot.next[v]
		if !ok {
			i += l
			continue
		}

		j := i
		var (
			word []rune
			res  *Result
		)
		for {
			word = append(word, v)
			if node.end {
				res = &Result{string(word), string(s[i : j+l]), i, j + l}
			}

			j += l
			// 跳过一些无意义的字符
			for {
				v, l = decodeBytes(s[j:])
				if j < n && skipper.ShouldSkip(v) {
					j += l
				} else {
					break
				}
			}

			node = node.next[v]
			// 如果找不到直接break
			if node == nil {
				if res == nil {
					j = i + 1
				} else if w.Write(res) {
					return
				}
				i = j
				continue next
			}
		}
	}

	return
}

// decodeStr 使用 utf8 解码字符串并返回 rune 和字节数
func decodeStr(s string) (r rune, size int) {
	return utf8.DecodeRuneInString(s)
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

func TestSearch_ReplaceRune(t *testing.T) {
	words := []string{"dog", "cat", "apple", "orange", "chicken", "鸭子", "水果", "敏感词"}
	obj := Strings(words)

	input := "I have a dog and a cat, and I love eating apples and oranges. I also like to eat chicken and duck (鸭子)."
	expectedOutput := "I have a *** and a ***, and I love eating *****s and ******s. I also like to eat ******* and duck (**)."

	output := obj.ReplaceRune([]byte(input), '*')

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

func TestSearch_FindWithSkip2(t *testing.T) {
	words := []string{"T*M!!!D"}
	obj := NewSearch(SetSortedSkip("!*"))
	obj.TrieWriter().InsertWords(words).BuildFail()

	if fmt.Sprintf("%v", obj.TrieWriter().Skip()) != "!*" {
		t.Fatalf("skipper is not match.")
	}

	str := "T***MD;T*M**D;T!MD;T#M#DFUCK"
	_, _ = obj.TrieWriter().WriteString("FUCK")
	obj.TrieWriter().BuildFail()
	res := obj.Find([]byte(str))

	type wantPair struct {
		word    string
		matched string
	}

	wants := []wantPair{
		{"TMD", "T***MD"},
		{"TMD", "T*M**D"},
		{"TMD", "T!MD"},
		{"FUCK", "FUCK"},
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

	if obj.TrieWriter().String() != "TMD\nFUCK" {
		t.Fatalf("trie writer sensitive words is not match.")
	}
}

func TestTrieWriter_InsertReader(t *testing.T) {
	writer := NewTrieWriter()
	skipper := &Skip{}
	skipper.Set("*!")
	writer.setSkip(skipper)
	_, _ = writer.insertReader(strings.NewReader("TMD\nfuck\n霸**王龙\n真的好吗"), make([]byte, 1024), '\n')
	want := []string{"TMD", "fuck", "霸王龙", "真的好吗"}
	res := writer.Array()
	if len(want) != len(res) {
		t.Fatalf("Incorrect number of sensitive words.want len:%d,result len:%d", len(want), len(res))
	}

	sort.Strings(res)
	sort.Strings(want)
	for i, v := range res {
		word := want[i]
		if word != v {
			t.Fatalf("Unable to match sensitive word: %s, result: %s", word, v)
		}
	}
}

func TestTrieWriter_InsertFile(t *testing.T) {
	obj, err := File("./example/word")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("./example/word")

	want := strings.Split(string(data), "\n")
	res := obj.TrieWriter().Array()
	sort.Strings(res)
	sort.Strings(want)
	if len(want) != len(res) {
		t.Fatalf("Incorrect number of sensitive words.want len:%d,result len:%d", len(want), len(res))
	}
}

func randWords(length int) []string {
	res := make([]string, length)
	str := []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz赵钱孙李周吴郑王冯陈褚卫蒋沈韩杨朱秦尤许何吕施张孔曹严华金魏陶姜戚谢邹喻柏水窦章云苏潘葛奚范彭郎鲁韦昌马苗凤花方俞任袁柳酆")
	strLen := len(str)
	exists := map[string]struct{}{}
	for i := 0; i < length; {
		l := 3 + rand.Intn(18)
		b := bytes.Buffer{}
		for l > 0 {
			idx := rand.Intn(strLen)
			b.WriteRune(str[idx])
			l--
		}
		item := b.String()
		if _, ok := exists[item]; ok {
			continue
		}
		exists[item] = struct{}{}
		res[i] = item
		i++
	}
	return res
}

func TestTrieWriter_InsertReaderBoundaryTesting(t *testing.T) {
	writer := NewTrieWriter()
	writer.setSkip(&Skip{})
	readerBuf := make([]byte, 1024*4)
	words := randWords(len(readerBuf) >> 1)
	reader := &bytes.Buffer{}
	var want []string
	for _, word := range words {
		if reader.Len()+len(word) > len(readerBuf) {
			left := len(readerBuf) - reader.Len()
			if left > 0 {
				s := make([]byte, 0, left)
				for left > 0 {
					s = append(s, 'e')
					left--
				}
				want = append(want, string(s))
				reader.Write(s)
			}
			break
		}
		want = append(want, word)
		reader.WriteString(word + "\n")
	}
	fmt.Println("reader len:", reader.Len())

	_, _ = writer.insertReader(reader, readerBuf, '\n')

	res := writer.Array()
	if len(want) != len(res) {
		t.Fatalf("Incorrect number of sensitive words.want len:%d,result len:%d", len(want), len(res))
	}

	sort.Strings(res)
	sort.Strings(want)
	for i, v := range res {
		word := want[i]
		if word != v {
			t.Fatalf("Unable to match sensitive word: %s, result: %s", word, v)
		}
	}
}

func TestTrieWriter_InsertScanner(t *testing.T) {
	writer := NewTrieWriter()
	writer.setSkip(&Skip{})
	readerBuf := make([]byte, 1024*4)
	words := randWords(len(readerBuf) >> 1)
	reader := &bytes.Buffer{}
	var want []string
	for _, word := range words {
		if reader.Len()+len(word) > len(readerBuf) {
			left := len(readerBuf) - reader.Len()
			if left > 0 {
				s := make([]byte, 0, left)
				for left > 0 {
					s = append(s, 'e')
					left--
				}
				want = append(want, string(s))
				reader.Write(s)
			}
			break
		}
		want = append(want, word)
		reader.WriteString(word + "\n")
	}

	writer.InsertScanner(bufio.NewScanner(reader))

	res := writer.Array()
	if len(want) != len(res) {
		t.Fatalf("Incorrect number of sensitive words.want len:%d,result len:%d", len(want), len(res))
	}

	sort.Strings(res)
	sort.Strings(want)
	for i, v := range res {
		word := want[i]
		if word != v {
			t.Fatalf("Unable to match sensitive word: %s, result: %s", word, v)
		}
	}
}
