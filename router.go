package router

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/ochipin/router/trie"
)

// New : ルーティングテーブル設定用構造体を生成する
func New() *RouteTable {
	rt := &RouteTable{
		regex:   make(map[string]string),
		classes: make(map[string]interface{}),
		routes:  make(map[string]map[string]*Route),
	}
	rt.Generator = rt
	return rt
}

// Route : ルーティングパスの情報を取り扱う構造体
type Route struct {
	ctlname string // コントローラ名
	actname string // アクション名
	prior   bool   // 処理優先度。正規表現を使用されていた場合、優先度は低となる
}

// RouteTable : ルーティングテーブル設定構造体
type RouteTable struct {
	regex     map[string]string            // 正規表現登録用オブジェクト
	classes   map[string]interface{}       // 構造体登録用オブジェクト
	routes    map[string]map[string]*Route // ルーティングパス登録用オブジェクト
	Generator Generator
}

// MixinClass : 指定したコントローラがミックスインされているか確認する
// ex) r.MixinClass(Base{}, "mux.Controller") // true : mix-in されている
func (rt *RouteTable) MixinClass(i interface{}, name string) bool {
	// クラス情報を格納するマップが作成されていない場合は、false を返却する
	if rt.classes == nil {
		return false
	}

	// pkgname.Controller 形式の文字列ではない場合、falseを返却する
	idx := strings.Index(name, ".")
	if idx == -1 {
		return false
	}

	// 型情報を取得する
	typ := reflect.TypeOf(i)
	// 指定された型情報が存在しない場合はfalseを返却する
	v, ok := rt.classes[typ.Name()]
	if !ok {
		return false
	}

	// 型名とnameが一致した場合はtrueを返却する
	val := reflect.ValueOf(v)
	if val.Type().String() == name {
		return true
	}

	search := name[idx+1:]
	for {
		val = val.FieldByName(search)
		if val.IsValid() == false {
			return false
		}
		if val.Type().String() == name {
			break
		}
	}

	return true
}

// BoolClass : 指定したコントローラが登録されているか確認する
func (rt *RouteTable) BoolClass(i interface{}) bool {
	if rt.classes == nil {
		return false
	}

	typ := reflect.TypeOf(i)
	if typ.Kind() != reflect.Struct {
		return false
	}
	if v, ok := rt.classes[typ.Name()]; ok {
		return reflect.TypeOf(v).String() == typ.String()
	}

	return false
}

// GetRegexp : 登録されている正規表現情報を返却する
func (rt *RouteTable) GetRegexp(id string) string {
	if rt.regex != nil {
		if v, ok := rt.regex[":"+id]; ok {
			return v
		}
	}
	return ""
}

// GetRouter : 登録されているルート情報を返却する
func (rt *RouteTable) GetRouter(method, path string) string {
	if rt.routes == nil {
		return ""
	}
	routes, ok := rt.routes[method]
	if !ok {
		return ""
	}
	route, ok := routes[path]
	if !ok {
		return ""
	}

	return route.ctlname + "." + route.actname
}

// TableList : 登録されているルート情報を返却する
func (rt *RouteTable) TableList() map[string][][]string {
	var list = make(map[string][][]string)

	// id: [0-9]+ 等の登録されている正規表現オブジェクトを取得する
	for k, v := range rt.regex {
		list["REGEXP"] = append(list["REGEXP"], []string{k, v})
	}
	if _, ok := list["REGEXP"]; !ok {
		list["REGEXP"] = [][]string{}
	}

	// ルーティングテーブル情報を取得する
	for m, v := range rt.routes {
		for p, route := range v {
			list["ROUTER"] = append(list["ROUTER"], []string{m, p, route.ctlname + "." + route.actname})
		}
	}
	if _, ok := list["ROUTER"]; !ok {
		list["ROUTER"] = [][]string{}
	}

	return list
}

// Action : Action を生成する
func (rt *RouteTable) Action(ctlname, actname string, controller interface{}) Result {
	return &Action{
		Ctlname:    ctlname,
		Actname:    actname,
		Controller: controller,
	}
}

// SetRegexp : 正規表現形式のルートパスを設定する際に、使用される正規表現を登録する(複数)
func (rt *RouteTable) SetRegexp(regex map[string]string) error {
	if regex == nil {
		return fmt.Errorf("argument is nil")
	}

	rt.regex = make(map[string]string)

	for k, v := range regex {
		if err := rt.AddRegexp(k, v); err != nil {
			return err
		}
	}

	return nil
}

// AddRegexp : 正規表現形式のルートパスを設定する際に、使用される正規表現を登録する(単体)
func (rt *RouteTable) AddRegexp(id, regex string) error {
	if id == "" {
		return fmt.Errorf("key name is empty")
	}
	if _, err := regexp.Compile(regex); err != nil {
		return fmt.Errorf("'%s' - invalid regexp. '%s' not used", id, regex)
	}
	rt.regex[":"+id] = regex
	return nil
}

// SetClass : 該当するルートパスへアクセスした際に、実行するコントローラを登録する(複数)
func (rt *RouteTable) SetClass(i []interface{}) error {
	// 複数指定の場合、与えられた構造体をすべて登録する
	for _, v := range i {
		if err := rt.AddClass(v); err != nil {
			return err
		}
	}
	return nil
}

// AddClass : 該当するルートパスへアクセスした際に、実行するコントローラを登録する(単体)
func (rt *RouteTable) AddClass(i interface{}) error {
	typ := reflect.TypeOf(i)
	if typ == nil {
		return fmt.Errorf("invalid argument. is nil")
	}
	// 構造体形式ではない場合、エラーを返却する
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("invalid argument. '%s' is not struct", typ.Name())
	}
	// 与えられた構造体を登録する
	rt.classes[typ.Name()] = i
	return nil
}

// Register : ルートパスを登録する
func (rt *RouteTable) Register(method, path, name string) error {
	// コントローラ名、アクション名を抜き出す
	names := strings.Split(name, ".")
	if len(names) != 2 {
		if name == "" {
			return fmt.Errorf("controller.action name is empty")
		}
		return fmt.Errorf("'%s' - invalid controller.action name", name)
	}

	// path が空文字列の場合はエラーを返却する
	if path == "" {
		return fmt.Errorf("path is empty")
	}

	// プライオリティ値を図る
	prior := strings.Count(path, ":") == 0

	// GET, POSTなどのリクエストメソッドを受け取る箱がない場合は作成する
	if _, ok := rt.routes[method]; !ok {
		rt.routes[method] = make(map[string]*Route)
	}
	// ルーティングテーブルを作成する
	rt.routes[method][path] = &Route{
		ctlname: names[0],
		actname: names[1],
		prior:   prior,
	}

	return nil
}

// Create : 登録されたルートパスを
func (rt *RouteTable) Create() (Router, error) {
	var result = make(Router)

	// ex) map[GET]map[/:id]*Route を GET, map[/:id]*Route として処理する
	for method, routes := range rt.routes {
		// ルーティング構造体を生成
		routing := &Routing{
			access: new(trie.Trie),
			regexp: make(map[*regexp.Regexp]interface{}),
		}
		result[method] = routing
		// map[/:id]*Route を /:id, *Route として処理する
		for path, route := range routes {
			// コントローラオブジェクトを取得する
			controller, ok := rt.classes[route.ctlname]
			if !ok {
				return nil, fmt.Errorf("'%s' - controller not registered", route.ctlname)
			}
			// アクションオブジェクトを生成する
			if rt.Generator == nil {
				return nil, fmt.Errorf("action generator is nil pointer")
			}
			action := rt.Generator.Action(route.ctlname, route.actname, controller)
			// アクションオブジェクトが正しい設定値であるか検証する
			if _, err := action.Get(); err != nil {
				return nil, err
			}
			// パスを設定する
			if !route.prior {
				// 優先度が低い場合、パス内の:<name>を正規表現文字列に置き換える
				var p = path
				for name, reg := range rt.regex {
					p = strings.Replace(p, name, reg, -1)
				}
				// 正しく正規表現が置き換えられたかチェックする
				if strings.Count(p, ":") != 0 {
					return nil, fmt.Errorf("regexp in '%s' path is not registered", path)
				}
				// 正規表現を使用したアクセスパスを生成する
				regexp, err := regexp.Compile("^" + p + "$")
				if err != nil {
					return nil, fmt.Errorf("'%s.%s' - %s", route.ctlname, route.actname, err)
				}
				routing.regexp[regexp] = action
			} else {
				// 優先度が高い場合、固定パスを登録する
				routing.access.Add(path, action)
			}
		}
	}
	return result, nil
}

// Routing : ルーティングパス構造体
type Routing struct {
	access *trie.Trie                     // 固定パス
	regexp map[*regexp.Regexp]interface{} // 正規表現形式のパス
}

// Router : 各ルーティングパスを、メソッド(GET/POST)単位で取り扱うマップ
type Router map[string]*Routing

// Caller : 関数実行用オブジェクトを返却する
func (r Router) Caller(method, path string) (Result, []reflect.Value, error) {
	var args []reflect.Value

	// 指定されたメソッド名に該当するルーティング構造体を取得する
	routing, ok := r[method]
	if !ok {
		return nil, nil, fmt.Errorf("'%s' - method not found", method)
	}

	// 指定されたパスを固定パスとしてアクションを取得する
	i := routing.access.Get(path)
	// 固定パスとして取得できない場合、正規表現形式のパスとしてアクションを取得する
	if i == nil {
		for reg, obj := range routing.regexp {
			// マッチしない場合は、次の正規表現へ
			if !reg.MatchString(path) {
				continue
			}
			// マッチした場合は、正規表現で引っかかった文字列のみを抜き出す
			strs := reg.FindAllStringSubmatch(path, -1)
			if len(strs) > 0 {
				// 抜き出した文字列を配列へ格納する
				for i := 1; i < len(strs[0]); i++ {
					args = append(args, reflect.ValueOf(strs[0][i]))
				}
			}
			i = obj
			break
		}
	}

	// アクションの取得失敗の場合、nil を返却する
	if i == nil {
		return nil, nil, fmt.Errorf("'%s' - path not found", path)
	}

	// interface{} を Action 構造体へ変換する
	action, ok := i.(Result)
	if !ok {
		return nil, nil, fmt.Errorf("action struct is invalid")
	}

	return action, args, nil
}

// Generator : 生成するアクションオブジェクトのジェネレータ
type Generator interface {
	Action(string, string, interface{}) Result
}

// Result : Action() が返却するResult型
type Result interface {
	Get() (*reflect.Value, error)
	Call([]reflect.Value) ([]reflect.Value, error)
	Name() (string, string)
}

// Action : アクション登録用構造体
type Action struct {
	Ctlname    string
	Actname    string
	Controller interface{}
}

// Get : アクションを実行するCallerを取得する
func (action *Action) Get() (*reflect.Value, error) {
	// 登録済みのコントローラの型が存在するかチェックする
	typ := reflect.TypeOf(action.Controller)
	if typ == nil {
		return nil, fmt.Errorf("'%s.%s' - is nil", action.Ctlname, action.Actname)
	}

	// 登録済みのコントローラの型が構造体型かチェックする
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("'%s.%s' - not struct type", action.Ctlname, action.Actname)
	}

	caller := reflect.New(typ)
	// コールする関数情報が不正ではないかチェックする
	if caller.MethodByName(action.Actname).IsValid() == false {
		return nil, fmt.Errorf("'%s.%s' - function undefined", action.Ctlname, action.Actname)
	}

	return &caller, nil
}

// Call : 関数をコールする
func (action *Action) Call(args []reflect.Value) ([]reflect.Value, error) {
	// メソッド情報を取得
	caller, err := action.Get()
	if err != nil {
		return nil, err
	}
	fn := caller.MethodByName(action.Actname)

	// 型情報を取得し、引数があるかチェックする
	typ := reflect.TypeOf(fn.Interface())
	if len(args) != typ.NumIn() {
		return nil, fmt.Errorf("'%s.%s' - argument error", action.Ctlname, action.Actname)
	}

	// 関数をコール
	return fn.Call(args), nil
}

// Name : コントローラ名とアクション名を返却する
func (action *Action) Name() (string, string) {
	return action.Ctlname, action.Actname
}
