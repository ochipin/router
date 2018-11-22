package trie

import "fmt"

// Trie : トライ木でURLを管理する構造体
type Trie struct {
	childs   map[rune]*Trie // 次の要素
	endpoint bool           // 最後尾に到達した時点で true となる
	object   interface{}    // 登録するオブジェクト
}

// Add : 新規ノードを追加する
func (t *Trie) Add(path string, object interface{}) error {
	// 引数のpathが空文字列の場合関数を抜ける
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	// 次要素管理マップが作られていない場合作成する
	if t.childs == nil {
		t.childs = make(map[rune]*Trie)
	}

	// ノード追加処理
	i := t
	for _, v := range path {
		trie, ok := i.childs[v]
		if !ok {
			trie = new(Trie)
			trie.childs = make(map[rune]*Trie)
			i.childs[v] = trie
		}
		i = trie
	}

	// 既に登録済みの場合復帰
	if i.endpoint {
		return fmt.Errorf("'%s' is already exists", path)
	}

	// ノードに要素を追加する
	i.endpoint = true
	i.object = object

	return nil
}

// Get : ノードを取り出す
func (t *Trie) Get(path string) interface{} {
	if path == "" {
		return nil
	}

	// ノード検索開始
	i := t
	for _, v := range path {
		trie, ok := i.childs[v]
		if !ok {
			return nil
		}
		i = trie
	}

	// 検索結果を返却する
	return i.object
}
