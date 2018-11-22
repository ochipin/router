ルーティングテーブル管理ライブラリ
===

```go
package main

import (
	"fmt"

	"github.com/ochipin/router"
)

type Sample struct{}

func (s Sample) Index() {
	fmt.Println("Index Function called")
}

func (s Sample) Hello(args string) string {
	fmt.Println("Hello Function called")
	return fmt.Sprintf("Hello %s", args)
}

func main() {
	// ルーティングテーブル構築用インスタンスを生成
	r := router.New()

	// Sample構造体を登録
	r.SetClass([]interface{}{Sample{}})

	// 使用する正規表現を登録する。キー名、正規表現として登録すること。
	// 正規表現で抜き出した値がほしい場合は、()で囲む
	r.AddRegexp("id", "([Wa-z!]+)")

	// アクセスパスを設定する。
	// '/' の場合は、Sample.Index をコールする
	r.Register("GET", "/", "Sample.Index")
	// 正規表現アクセスパスを使用する場合は、':' の後に登録した正規表現キー名を指定する。
	// '/:id'、 すなわち /([Wa-z!]+) の場合 Sample.Hello をコール
	r.Register("GET", "/:id", "Sample.Hello")

	// 構造体、正規表現、ルーティングを設定した後、ルータを作成する
	data, err := r.Create()
	if err != nil {
		panic(err)
	}

	// '/' にアクセスして、Sample.Index をコールする準備開始
	res, args, err := data.Caller("GET", "/")
	if err != nil {
		panic(err)
	}
	// res は Result 型で、Call 関数を実行することで、Sample.Index をコール可能
	if _, err := res.Call(args); err != nil {
		panic(err)
	}

	// '/World!' にアクセスして、Sample.Hello をコールする準備開始
	// args には、 /:id に設定した正規表現を抜き出した値が格納される。
	// AddRegexp, または SetRegexp で正規表現を設定する際に、()で囲むことを忘れずに。
	res, args, err = data.Caller("GET", "/World!")
	if err != nil {
		panic(err)
	}

	// Sample.Hello を引数 World! を添えてコール
	// out には、Sample.Hello の復帰値が格納される
	out, err := res.Call(args)
	if err != nil {
		panic(err)
	}
	// Hello World! が表示される
	fmt.Println(out[0].String())
}
```

```go
/*
 * 独自アクションジェネレータを実装
 */
type MyGenerator struct{}

func (act MyGenerator) Action(ctlname, actname string, i interface{}) router.Result {
	action := &MyAction{}
	action.Ctlname = ctlname
	action.Actname = actname
	action.Controller = i
	return action
}

type MyAction struct {
	router.Action
}

func (act *MyAction) Call(args []reflect.Value) ([]reflect.Value, error) {
	caller, err := act.Get()
	if err != nil {
		return nil, err
	}

	fmt.Println("MyAction Call()")
	return caller.Call(args), nil
}

func main() {
	// ルーティングテーブル構築用インスタンスを生成
	r := router.New()
	// 独自アクションジェネレータを登録
	r.Generator = &MyGenerator{}
	...
```