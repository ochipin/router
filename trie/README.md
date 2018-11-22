Trieライブラリ
================================================
Trie木でオブジェクトを管理します。

```go
package main

import (
    "fmt"
    "github.com/ochipin/trie"
)

func main() {
    r := new(trie.Trie)

    r.Add("key1", "value1")
    r.Add("key2", "value2")

    fmt.Println(r.Get("key1")) // value1
    fmt.Println(r.Get("key2")) // value2
    fmt.Println(r.Get("key3")) // <nil>
    fmt.Println(r.Get("key"))  // <nil>
}
```

`Get`で取得する値は `interface{}` 型です。
