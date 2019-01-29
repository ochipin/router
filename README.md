ルーティングテーブル管理ライブラリ
===
ルーティングテーブルを生成するライブラリ。

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

	// Call で呼び出す前に、事前チェックをする
	// 第2引数以降は、復帰値の型を文字列で指定可能
	elem, _ := res.Get()
	fn, err := res.Valid(elem, args, "string")
	if err != nil {
		panic(err)
	}
	out = fn.Call(args)
	// Hello World! が表示される
	fmt.Println(out[0].String())

	// 直接名前を指定して、アクションを実行する
	out, err = res.Callname(elem, "Hello", []reflect.ValueOf{
		reflect.ValueOf("Hello"),
	}, "string")
	// Hello Hello が出力される
	fmt.Println(out[0].String())

	// HasName 関数を利用して、構造体、メソッドを所持しているか確認
	if router.HasName(elem, "Hello", "World") == true {
		// Hello.World を所持している
	}
}
```

Valid 関数の復帰値の error は次の構造体により構成されている。

```go
switch types := err.(type) {
// コールするメソッドの引数の数が不正
case *router.NotEnoughArgs:
	...
// コールするメソッドの引数の型が不正
case *router.IllegalArgs:
	...
// コールするメソッドの復帰値の数が不正
case *router.NotEnoughRets:
	...
// コールするメソッドの復帰値の型が不正
case *router.IllegalRets:
	...
}
```
各エラー型の構造体は、下記メンバ変数を所持している。

* Message  
エラーメッセージ本文
* Have  
関数の引数、復帰値に指定した型情報
* Want  
関数の引数、復帰値に指定しなければならない型情報

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

`SetStruct`関数を用いることで、登録されている構造体に値をセット可能。
ただし、Mix-inしているオブジェクトに限る。

```go
type Base struct {
	...
}
type Sample struct {
	Base
}

...

res, args, err = data.Caller("GET", "/World!")
if err != nil {
	panic(err)
}

act, _ := res.Get()
// act(Sample)が所持するBaseに値を設定する
err := router.SetStruct(act, &Base{
	// 値を設定
})
if err != nil {
	panic(err)
}
```

SetStruct が返却するエラー型は次のとおり。

```go
switch types := err.(type) {
// 指定された引数に nil が渡されたなどした場合
case *router.InvalidError:
	...
// 指定された引数が構造体型ではない場合
case *router.NoStruct:
	...
// 値をセットしようとしたが、ミックスインされていない
case *router.NoMixin:
	...
}
```