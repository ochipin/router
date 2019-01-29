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
		return nil, nil, &NotRoutes{
			Message: fmt.Sprintf("'[%s]: %s' - not found", method, path),
			Method:  method,
			Path:    path,
		}
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
		return nil, nil, &NotRoutes{
			Message: fmt.Sprintf("'[%s]: %s' - not found", method, path),
			Method:  method,
			Path:    path,
		}
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
	Get() (reflect.Value, error)
	Call([]reflect.Value, ...string) ([]reflect.Value, error)
	Name() (string, string)
	Valid(reflect.Value, []reflect.Value, ...string) (reflect.Value, error)
	Callname(reflect.Value, string, []reflect.Value, ...string) ([]reflect.Value, error)
}

// Action : アクション登録用構造体
type Action struct {
	Ctlname    string
	Actname    string
	Controller interface{}
}

// Get : アクションを実行するCallerを取得する
func (action *Action) Get() (reflect.Value, error) {
	// 登録済みのコントローラの型が存在するかチェックする
	typ := reflect.TypeOf(action.Controller)
	if typ == nil {
		return reflect.Value{}, fmt.Errorf("'%s.%s' - is nil", action.Ctlname, action.Actname)
	}

	// 登録済みのコントローラの型が構造体型かチェックする
	if typ.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("'%s.%s' - not struct type", action.Ctlname, action.Actname)
	}

	caller := reflect.New(typ)
	// コールする関数情報が不正ではないかチェックする
	if caller.MethodByName(action.Actname).IsValid() == false {
		return reflect.Value{}, fmt.Errorf("'%s.%s' - function undefined", action.Ctlname, action.Actname)
	}

	return caller, nil
}

// Valid : 与えられた引数の数、型、復帰値の数、型がコールするメソッドと一致している場合、メソッド情報を返却する
func (action *Action) Valid(caller reflect.Value, args []reflect.Value, ret ...string) (reflect.Value, error) {
	// メソッド情報を取得
	fn := caller.MethodByName(action.Actname)
	return action.valid(fn, action.Actname, args, ret...)
}

func (action *Action) valid(fn reflect.Value, actname string, args []reflect.Value, ret ...string) (reflect.Value, error) {
	if fn.IsValid() == false {
		return reflect.Value{}, &InvalidError{
			Message: fmt.Sprintf("'%s.%s' function does not exists", action.Ctlname, actname),
		}
	}
	typ := reflect.TypeOf(fn.Interface())
	// 型情報を取得し、引数の数が一致するかチェックする
	if len(args) != typ.NumIn() {
		var have []string
		// 呼び出し元の引数情報を取得
		for i := 0; i < len(args); i++ {
			have = append(have, args[i].Type().String())
		}
		var want []string
		// 呼び出し先の引数情報を取得
		for i := 0; i < typ.NumIn(); i++ {
			want = append(want, typ.In(i).String())
		}
		// エラーを返却する
		return reflect.Value{}, &NotEnoughArgs{
			Message: fmt.Sprintf("not enough arguments in call to '%s.%s'. have = %d, want = %d",
				action.Ctlname, actname, len(args), typ.NumIn()),
			Have: fmt.Sprintf("(%s)", strings.Join(have, ", ")),
			Want: fmt.Sprintf("(%s)", strings.Join(want, ", ")),
		}
	}

	// 引数の型情報が正しいかチェックする
	for i := 0; i < typ.NumIn(); i++ {
		// 型情報が一致している場合、次の型情報へ
		if typ.In(i) == args[i].Type() {
			continue
		}

		var errflag = true
		if typ.In(i).Kind() == reflect.Interface {
			// コールする関数の引数の型情報がinterfaceの場合
			func() {
				defer func() {
					if err := recover(); err != nil {
						errflag = true
					}
				}()
				errflag = false
				reflect.New(typ.In(i)).Elem().Set(args[i])
			}()
		} else if typ.In(i).Kind() == args[i].Type().Kind() {
			// コールする関数の引数のKindは同じ場合
			func() {
				defer func() {
					if err := recover(); err != nil {
						errflag = true
					}
				}()
				errflag = false
				v := args[i].Convert(typ.In(i))
				args[i] = v
			}()
		}
		if errflag == false {
			continue
		}
		var have, want []string
		for i := 0; i < len(args); i++ {
			have = append(have, args[i].Type().String())
			want = append(want, typ.In(i).String())
		}
		return reflect.Value{}, &IllegalArgs{
			Message: fmt.Sprintf("cannot use (type %s) as type %s in argument to '%s.%s'",
				args[i].Type().Name(), typ.In(i).Name(), action.Ctlname, actname),
			Have: fmt.Sprintf("(%s)", strings.Join(have, ", ")),
			Want: fmt.Sprintf("(%s)", strings.Join(want, ", ")),
		}
	}

	// ret が無指定の場合、復帰値の検証は行わない
	if len(ret) == 0 {
		return fn, nil
	}

	// ret が指定されている場合、復帰値の数を検証する
	if typ.NumOut() != len(ret) {
		var have []string
		// 呼び出し先の引数情報を取得
		for i := 0; i < typ.NumIn(); i++ {
			have = append(have, typ.In(i).String())
		}
		return reflect.Value{}, &NotEnoughRets{
			Message: fmt.Sprintf("not enough arguments to return '%s.%s'. have = %d, want = %d",
				action.Ctlname, actname, len(ret), typ.NumOut()),
			Have: fmt.Sprintf("(%s)", strings.Join(have, ", ")),
			Want: fmt.Sprintf("(%s)", strings.Join(ret, ", ")),
		}
	}

	// 復帰値の型を検証する
	for i := 0; i < typ.NumOut(); i++ {
		// 型情報が一致している場合、次の型情報へ
		if typ.Out(i).String() == ret[i] {
			continue
		}
		var have []string
		// 呼び出し先の引数情報を取得
		for i := 0; i < typ.NumIn(); i++ {
			have = append(have, typ.In(i).String())
		}
		return reflect.Value{}, &IllegalRets{
			Message: fmt.Sprintf("cannot use (type %s) as type %s in return argument '%s.%s'",
				typ.Out(i).String(), ret[i], action.Ctlname, actname),
			Have: fmt.Sprintf("(%s)", strings.Join(have, ", ")),
			Want: fmt.Sprintf("(%s)", strings.Join(ret, ", ")),
		}
	}

	return fn, nil
}

// Callname : 指定した名前で、メソッドを実行する
func (action *Action) Callname(elem reflect.Value, methodname string, args []reflect.Value, ret ...string) ([]reflect.Value, error) {
	method := elem.MethodByName(methodname)
	fn, err := action.valid(method, methodname, args, ret...)
	if err != nil {
		return nil, err
	}
	return fn.Call(args), nil
}

// Call : 関数をコールする
func (action *Action) Call(args []reflect.Value, ret ...string) ([]reflect.Value, error) {
	// アクション情報を取得
	caller, err := action.Get()
	if err != nil {
		return nil, err
	}

	fn, err := action.Valid(caller, args, ret...)
	if err != nil {
		return nil, err
	}
	return fn.Call(args), nil
}

// Name : コントローラ名とアクション名を返却する
func (action *Action) Name() (string, string) {
	return action.Ctlname, action.Actname
}

// SetStruct : 指定された reflect.Value が所持する構造体に i を格納する
func SetStruct(action reflect.Value, i interface{}) error {
	// action が有効か否かを判定する
	if action.IsValid() == false {
		return &InvalidError{
			Message: fmt.Sprintf("detects incorrect argument value. action = invalid"),
		}
	}
	// ポインタの場合は、中身を指す
	for action.Kind() == reflect.Ptr {
		action = action.Elem()
	}
	// action が有効か否かを判定する
	if action.IsValid() == false {
		return &InvalidError{
			Message: fmt.Sprintf("detects incorrect argument value. action = invalid"),
		}
	}
	// 構造体ではない場合、エラーを返却する
	if action.Kind() != reflect.Struct {
		kind := action.Kind().String()
		name := action.Type().String()
		return &NoStruct{
			Message: fmt.Sprintf("argument is not struct type. action is type = %s, kind = %s", kind, name),
			Kind:    kind,
			Type:    name,
		}
	}

	// 与えられた引数 i をreflect.Value型に変更する
	v := reflect.ValueOf(i)
	// 有効か否かを判定する
	if v.IsValid() == false {
		return &InvalidError{
			Message: fmt.Sprintf("detects incorrect argument value. i = invalid"),
		}
	}
	// i から基本型名を取得する (ex: router.Structname)
	basicname := v.Type().String()
	// i がポインタで渡された場合は、中身を指す
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.IsValid() == false {
		return &InvalidError{
			Message: fmt.Sprintf("detects incorrect argument value. i = invalid"),
		}
	}
	// i が構造体ではない場合、エラーを返却する
	if v.Kind() != reflect.Struct {
		kind := v.Kind().String()
		name := v.Type().String()
		return &NoStruct{
			Message: fmt.Sprintf("argument is not struct type. i is type = %s, kind = %s", kind, name),
			Kind:    kind,
			Type:    name,
		}
	}
	// i のメンバ名(ミックスイン名)を取得 (ex: Structname)
	membername := v.Type().Name()

	if action.Type().String() == basicname {
		action.Set(reflect.ValueOf(i))
		return nil
	}

	// 引数で指定した構造体がミックスインされているか確認する
	for {
		action = action.FieldByName(membername)
		// 指定された型情報がミックスインされていない場合、エラーを返却する
		if action.IsValid() == false {
			return &NoMixin{
				Message:    fmt.Sprintf("have not mix-in \"%s\"", basicname),
				Basicname:  basicname,
				Membername: membername,
			}
		}
		// 型情報が一致した場合、ループを抜ける
		if action.Type().String() == basicname {
			break
		}
		// 型が一致せず、actionがポインタの場合中身を指すようにする
		for action.Kind() == reflect.Ptr {
			action = action.Elem()
		}
		// Invalid の場合は、エラーを返却する
		if action.IsValid() == false {
			return &NoMixin{
				Message:    fmt.Sprintf("have not mix-in \"%s\"", basicname),
				Basicname:  basicname,
				Membername: membername,
			}
		}
		// Kindがstructではない場合、エラーを返却する
		if action.Kind() != reflect.Struct {
			return &NoMixin{
				Message:    fmt.Sprintf("have not mix-in \"%s\"", basicname),
				Basicname:  basicname,
				Membername: membername,
			}
		}
	}

	action.Set(reflect.ValueOf(i))
	return nil
}

// HasName : 指定された名前の構造体、メソッドを所持しているか確認する
func HasName(elem reflect.Value, ctlname, actname string) bool {
	if elem.IsValid() == false {
		return false
	}
	var e = elem
	for elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	if elem.Kind() != reflect.Struct {
		return false
	}
	return e.Type().String() == ctlname && e.MethodByName(actname).IsValid()
}

// NotRoutes : 指定したパス、またはメソッドが存在しない場合のエラー型
type NotRoutes struct {
	Message string
	Path    string
	Method  string
}

func (err *NotRoutes) Error() string {
	return err.Message
}

// NotEnoughArgs : コールするメソッドの引数の数が一致しない場合のエラー型
type NotEnoughArgs struct {
	Message string
	Have    string
	Want    string
}

func (err *NotEnoughArgs) Error() string {
	return err.Message
}

// IllegalArgs : コールするメソッドの引数の型が一致しない場合のエラー型
type IllegalArgs struct {
	Message string
	Have    string
	Want    string
}

func (err *IllegalArgs) Error() string {
	return err.Message
}

// NotEnoughRets : コールするメソッドの復帰値の数が一致しない場合のエラー型
type NotEnoughRets struct {
	Message string
	Have    string
	Want    string
}

func (err *NotEnoughRets) Error() string {
	return err.Message
}

// IllegalRets : コールするメソッドの復帰値の型が一致しない場合のエラー型
type IllegalRets struct {
	Message string
	Have    string
	Want    string
}

func (err *IllegalRets) Error() string {
	return err.Message
}

// InvalidError : SetStructに渡した引数が不正な場合のエラー型
type InvalidError struct {
	Message string
}

func (err *InvalidError) Error() string {
	return err.Message
}

// NoStruct : SetStructに渡した引数が構造体型ではない場合のエラー型
type NoStruct struct {
	Message string
	Kind    string
	Type    string
}

func (err *NoStruct) Error() string {
	return err.Message
}

// NoMixin : 指定した構造体がミックスインされていない場合のエラー型
type NoMixin struct {
	Message    string
	Membername string
	Basicname  string
}

func (err *NoMixin) Error() string {
	return err.Message
}
