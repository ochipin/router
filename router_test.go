package router

import (
	"fmt"
	"reflect"
	"testing"
)

type Sample struct{ Method string }

func (s *Sample) Index()                   { fmt.Println("Index page") }
func (s *Sample) Hello(a, b string) string { return fmt.Sprintf("Hello %s %s", a, b) }
func (s *Sample) World()                   { fmt.Println("OK!! World") }
func (s *Sample) TheTest(a string)         { fmt.Println("The Test") }

func Test__ROUTER_SUCCESS1(t *testing.T) {
	// ルーティングテーブル作成用オブジェクトを生成
	var data = New()

	// パスとコントローラをそれぞれ登録
	if err := data.Register("GET", "/", "Sample.Index"); err != nil {
		t.Fatal(err)
	}
	if err := data.Register("POST", "/:id", "Sample.Hello"); err != nil {
		t.Fatal(err)
	}
	if err := data.Register("DELETE", "/9088", "Sample.World"); err != nil {
		t.Fatal(err)
	}
	if err := data.Register("PUT", "/:n/:n", "Sample.Hello"); err != nil {
		t.Fatal(err)
	}
	// コントローラ名の部分に空文字列を渡す。エラーとなる
	if err := data.Register("GET", "/sample", ""); err == nil {
		t.Fatal("data.Register is error")
	}
	// コントローラ名の部分にコントローラ名のみを渡す。エラーとなる
	if err := data.Register("GET", "/sample", "Sample"); err == nil {
		t.Fatal("data.Register is error")
	}
	// パス部分に空文字列を渡す。エラーとなる
	if err := data.Register("GET", "", "Sample.Hello"); err == nil {
		t.Fatal("data.Register is error")
	}

	// Sample 構造体をコントローラとして登録する。 nil が含まれているため、エラーとなる
	if err := data.SetClass([]interface{}{Sample{}, nil}); err == nil {
		t.Fatal("data.SetController is error")
	}
	// 構造体ではない型を登録する。エラーとなる
	if err := data.SetClass([]interface{}{200}); err == nil {
		t.Fatal("data.SetController is error")
	}
	// Sample 構造体をコントローラとして登録する
	if err := data.SetClass([]interface{}{Sample{}}); err != nil {
		t.Fatal("data.SetController is error")
	}

	// 不正な正規表現オブジェクトの登録。エラーとなる。
	if err := data.SetRegexp(map[string]string{"": "[0-9]+"}); err == nil {
		t.Fatal(err)
	}
	// 不正な正規表現オブジェクトの登録。エラーとなる
	if err := data.SetRegexp(nil); err == nil {
		t.Fatal(err)
	}
	// 正規表現オブジェクトを登録
	data.SetRegexp(map[string]string{
		"id": "[0-9]+",
	})
	data.AddRegexp("n", "([0-9]+)")

	// 登録した情報から、ルーティングテーブルを生成する
	router, err := data.Create()
	if err != nil {
		t.Fatal(err)
	}

	// ルーティングテーブルから、ありえない名前でCallerを取得する。エラーとなる。
	if _, _, err := router.Caller("HEADER", ""); err == nil {
		t.Fatal("router.Caller is error")
	}
	// ルーティングテーブルから、どれにもマッチしないパスを指定する。エラーとなる。
	if _, _, err := router.Caller("GET", "/sample"); err == nil {
		t.Fatal("router.Caller is error")
	}
	// ルーティングテーブルから、Callerを取得する
	caller, args, err := router.Caller("PUT", "/1/2")
	if err != nil {
		t.Fatal(err)
	}

	ctlname, actname := caller.Name()
	if ctlname != "Sample" || actname != "Hello" {
		t.Fatal("Caller.Name is error")
	}
	// 関数をコール。実装している関数と、引数があっていないためエラーとなる
	if _, err := caller.Call([]reflect.Value{}); err == nil {
		t.Fatal("Caller.Call is error")
	}
	// 関数をコール
	result, err := caller.Call(args)
	if err != nil {
		t.Fatal(err)
	}

	// 復帰値を検証
	if len(result) != 1 {
		t.Fatal("return type not string")
	}

	fmt.Println(result[0].String())
}

func Test__ROUTER_FAILED1(t *testing.T) {
	// ルーティングテーブル作成用オブジェクトを生成
	var data = New()

	// パスとコントローラを登録
	if err := data.Register("GET", "/", "Sample.Method"); err != nil {
		t.Fatal(err)
	}

	// Sample 構造体をコントローラとして登録する
	if err := data.SetClass([]interface{}{Sample{}}); err != nil {
		t.Fatal("data.SetController is error")
	}

	// 登録した情報から、ルーティングテーブルを生成するが、登録されたアクションが関数ではないためエラーとなる
	if _, err := data.Create(); err == nil {
		t.Fatal(err)
	}
}

func Test__ROUTER_FAILED2(t *testing.T) {
	// ルーティングテーブル作成用オブジェクトを生成
	var data = New()

	// パスとコントローラを登録
	if err := data.Register("GET", "/", "Sample2.Method"); err != nil {
		t.Fatal(err)
	}

	// Sample 構造体をコントローラとして登録する
	if err := data.SetClass([]interface{}{Sample{}}); err != nil {
		t.Fatal("data.SetController is error")
	}

	// 登録した情報から、ルーティングテーブルを生成するが、登録されたコントローラが存在しないためエラーとなる
	if _, err := data.Create(); err == nil {
		t.Fatal(err)
	}
}

func Test__ROUTER_FAILED3(t *testing.T) {
	// ルーティングテーブル作成用オブジェクトを生成
	var data = New()

	// パスとコントローラを登録
	if err := data.Register("GET", "/:id", "Sample.Index"); err != nil {
		t.Fatal(err)
	}
	// 正規表現を登録
	data.AddRegexp("n", "([0-9]+)")

	// Sample 構造体をコントローラとして登録する
	if err := data.SetClass([]interface{}{Sample{}}); err != nil {
		t.Fatal("data.SetController is error")
	}

	// 登録した情報から、ルーティングテーブルを生成するが、登録された正規表現がないためエラーとなる
	if _, err := data.Create(); err == nil {
		t.Fatal(err)
	}
}

func Test__ROUTER_FAILED4(t *testing.T) {
	// ルーティングテーブル作成用オブジェクトを生成
	var data = New()

	// パスとコントローラを登録
	if err := data.Register("GET", "/:id", "Sample.Index"); err != nil {
		t.Fatal(err)
	}
	// 誤った正規表現を登録
	data.AddRegexp("n", "([0-9]+")

	// Sample 構造体をコントローラとして登録する
	if err := data.SetClass([]interface{}{Sample{}}); err != nil {
		t.Fatal("data.SetController is error")
	}

	// 登録した情報から、ルーティングテーブルを生成するが、登録された正規表現がないためエラーとなる
	if _, err := data.Create(); err == nil {
		t.Fatal(err)
	}
}

func Test__ROUTER_FAILED5(t *testing.T) {
	// ルーティングテーブル作成用オブジェクトを生成
	var data = New()

	// パスとコントローラを登録するが、不正な正規表現が含まれてしまっている場合
	if err := data.Register("GET", "/:id/(.+", "Sample.Index"); err != nil {
		t.Fatal(err)
	}
	// 正規表現を登録
	data.AddRegexp("id", "([0-9]+)")

	// Sample 構造体をコントローラとして登録する
	if err := data.SetClass([]interface{}{Sample{}}); err != nil {
		t.Fatal("data.SetController is error")
	}

	// 登録した情報から、ルーティングテーブルを生成するが、不正なパスのためエラーとなる
	if _, err := data.Create(); err == nil {
		t.Fatal(err)
	}
}

func Test__ROUTER_FAILED6(t *testing.T) {
	// ルーティングテーブル作成用オブジェクトを生成
	var data = New()

	// パスとコントローラを登録する
	data.Register("GET", "/:n", "Sample.Index")
	// 正規表現を登録
	data.AddRegexp("n", "([0-9]+)")

	// Sample 構造体をコントローラとして登録する
	if err := data.SetClass([]interface{}{Sample{}}); err != nil {
		t.Fatal("data.SetController is error")
	}

	// 登録した情報から、ルーティングテーブルを生成する
	router, _ := data.Create()
	if _, _, err := router.Caller("GET", "/OK"); err == nil {
		t.Fatal(err)
	}
}

type Value int
type Base struct {
	Sample
	Age Value
}

func Test__ROUTER_MIXIN(t *testing.T) {
	var rt RouteTable
	if rt.MixinClass(Base{}, "router.Sample") != false {
		t.Fatal("MixinClass: Error")
	}
	var r = New()

	if r.MixinClass(Base{}, "router.Sample") != false {
		t.Fatal("MixinClass: Error")
	}
	r.AddClass(Base{})

	if r.MixinClass(Base{}, "router.Sample") != true {
		t.Fatal("MixinClass: Error")
	}
	if r.MixinClass(Base{}, "router.Base") != true {
		t.Fatal("MixinClass: Error")
	}
	if r.MixinClass(Base{}, "Base") != false {
		t.Fatal("MixinClass: Error")
	}
	if r.MixinClass(200, "router.Value") != false {
		t.Fatal("MixinClass: Error")
	}
	if r.MixinClass(Base{}, "router.Name") != false {
		t.Fatal("MixinClass: Error")
	}
}

func Test__ROUTER_BOOL(t *testing.T) {
	var rt RouteTable
	if rt.BoolClass(Base{}) != false {
		t.Fatal("BoolClass: Error")
	}
	var r = New()
	r.AddClass(Base{})

	if r.BoolClass(200) != false {
		t.Fatal("BoolClass: Error")
	}

	if r.BoolClass(Sample{}) != false {
		t.Fatal("BoolClass: Error")
	}

	if r.BoolClass(Base{}) != true {
		t.Fatal("BoolClass: Error")
	}
}

func Test__ROUTER_REGEXP(t *testing.T) {
	var rt RouteTable
	if rt.GetRegexp("name") != "" {
		t.Fatal("GetRegexp: Error")
	}

	r := New()
	r.AddRegexp("id", "[0-9]+")
	if r.GetRegexp("id") == "" {
		t.Fatal("GetRegexp: Error")
	}
}

func Test__ROUTER_GETROUTER(t *testing.T) {
	var rt RouteTable
	if rt.GetRouter("GET", "/") != "" {
		t.Fatal("GetRouter: Error")
	}

	r := New()
	r.Register("GET", "/", "Controller.Name")
	if r.GetRouter("GET", "/") == "" {
		t.Fatal("GetRouter: Error")
	}
	if r.GetRouter("PUT", "/") != "" {
		t.Fatal("GetRouter: Error")
	}
	if r.GetRouter("GET", "/name") != "" {
		t.Fatal("GetRouter: Error")
	}
}

func Test__ROUTER_TABLELIST(t *testing.T) {
	r := New()
	m := r.TableList()
	if len(m["REGEXP"]) != 0 || len(m["ROUTER"]) != 0 {
		t.Fatal("TableList: ERROR")
	}

	r.AddRegexp("id", "[0-9]+")
	r.Register("GET", "/", "Base.Index")
	m = r.TableList()
	if len(m["REGEXP"]) == 0 || len(m["ROUTER"]) == 0 {
		t.Fatal("TableList: ERROR")
	}
}
