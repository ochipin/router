package trie

import (
	"fmt"
	"testing"
)

func Test__TRIEY(t *testing.T) {
	// トライオブジェクトを生成
	trie := new(Trie)

	// 空文字列のキー名を指定した場合はエラーとなる
	if err := trie.Add("", nil); err == nil {
		t.Fatal("trie.Add: fatal")
	}

	// name, age をセット
	trie.Add("name", 100)
	trie.Add("name2", 30)

	// 値を取り出し正しいか検証する
	if v := trie.Get("name"); fmt.Sprint(v) != "100" {
		t.Fatal("trie.Get: fatal")
	}
	if v := trie.Get("name2"); fmt.Sprint(v) != "30" {
		t.Fatal("trie.Get: fatal")
	}

	// 存在しないキー名を指定し、Get が nil を返却するか検証する
	if v := trie.Get(""); v != nil {
		t.Fatal("trie.Get: fatal")
	}
	if v := trie.Get("fullname"); v != nil {
		t.Fatal("trie.Get: fatal")
	}

	// すでに登録済みのキー名を指定してセットした場合、エラーになるか検証する
	if err := trie.Add("name", 200); err == nil {
		t.Fatal("trie.Add: fatal")
	}

	if v := trie.Get("name"); fmt.Sprint(v) != "100" {
		t.Fatal("trie.Get: fatal")
	}
}
