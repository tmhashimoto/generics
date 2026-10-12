// Package main — Go ジェネリクスの実務例
//
// ジェネリクスとは:
//   「処理の流れは同じで、扱う型だけが違う」コードを、型ごとにコピペせず
//   1つの定義で書ける仕組みです。
//
// 書き方の基本:
//   func 名前[型パラメータ 制約](引数...) 戻り値 { ... }
//   例: func Filter[T any](items []T, keep func(T) bool) []T
//
// 「制約」って何？:
//   型パラメータ T に「どんな型を入れてよいか」を決めるルールです。
//
//   any
//     「どんな型でもOK」。いちばんゆるい制約。
//     まずはこれだけ覚えれば、多くのジェネリクスは読めます。
//
//   comparable（このあと例 4・5 で登場）
//     「== や != で比べられる型」だけOK、という意味の制約です。
//     たとえば int や string は a == b と書けます。
//     いっぽう []int（スライス）は == で比べられないので対象外です。
//     map のキーは内部で「同じキーか？」を比べる必要があるため、
//     map を使う型（Set や Cache）では comparable が必要になります。
//
// 読む順番のすすめ:
//   1〜3 は any だけ → 4〜5 で comparable が出てくる
//
// 実行:
//
//	go run .
package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	fmt.Println("Go ジェネリクス実務例")
	fmt.Println("まず any（どんな型でもOK）→ あとで comparable（== できる型）")
	fmt.Println()

	demoFilterMap() // any
	demoAPIClient() // any
	demoResult()    // any
	demoSet()       // comparable（ここで初めて登場）
	demoCache()     // comparable
}

// ---------------------------------------------------------------------------
type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
	Role   string `json:"role"`
}

type Order struct {
	ID     int `json:"id"`
	UserID int `json:"user_id"`
	Total  int `json:"total"`
}

// ---------------------------------------------------------------------------
// 1. Filter / Map — スライス変換の定番（制約は any）
//
// 【なぜジェネリクス？】
//   「条件に合う要素だけ残す」「別の形に写す」は、要素の型が User でも Order
//   でもアルゴリズムは同じです。
//
// 【any】
//   T any は「T はどんな型でもよい」という意味です。
//   Filter の中では if や append しかしていないので、特別な制約は不要です。
// ---------------------------------------------------------------------------

// Filter は keep が true を返した要素だけを新しいスライスに集めます。
func Filter[T any](items []T, keep func(T) bool) []T {
	out := make([]T, 0, len(items))
	for _, item := range items {
		if keep(item) {
			out = append(out, item)
		}
	}
	return out
}

// Map は各要素に f を適用し、別の型のスライスを返します。
// T と U が別々なので、構造体 → 文字列 など変換に使えます。
func Map[T, U any](items []T, f func(T) U) []U {
	out := make([]U, len(items))
	for i, item := range items {
		out[i] = f(item)
	}
	return out
}

func demoFilterMap() {
	fmt.Println("=== 1. Filter[T any] / Map[T, U any] ===")
	fmt.Println("制約: any（どんな型でもOK）")
	fmt.Println("用途: 一覧の絞り込み・名前だけのリスト作成など")
	fmt.Println()

	users := []User{
		{ID: 1, Name: "Alice", Active: true, Role: "admin"},
		{ID: 2, Name: "Bob", Active: false, Role: "member"},
		{ID: 3, Name: "Carol", Active: true, Role: "member"},
	}

	active := Filter(users, func(u User) bool { return u.Active })
	names := Map(active, func(u User) string { return u.Name })
	fmt.Printf("  全ユーザー → Activeのみ → 名前だけ\n")
	fmt.Printf("  → %v\n", names)

	orders := []Order{
		{ID: 1, UserID: 1, Total: 500},
		{ID: 2, UserID: 3, Total: 12000},
		{ID: 3, UserID: 1, Total: 8000},
	}
	expensive := Filter(orders, func(o Order) bool { return o.Total >= 10000 })
	fmt.Printf("  合計 10000円以上の注文: %+v\n", expensive)
	fmt.Println()
}

// ---------------------------------------------------------------------------
// 2. DecodeJSON — API レスポンスの型付きデコード（制約は any）
//
// 【なぜジェネリクス？】
//   エンドポイントごとに返す型が違う（User / []Order / ...）ので、
//   「何にデコードするか」を呼び出し側の [T] で指定します。
// ---------------------------------------------------------------------------

func DecodeJSON[T any](raw string) (T, error) {
	var out T
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return out, err
	}
	return out, nil
}

func demoAPIClient() {
	fmt.Println("=== 2. DecodeJSON[T any] ===")
	fmt.Println("制約: any（どんな型でもOK）")
	fmt.Println("用途: REST API の JSON を型付きで取り出す")
	fmt.Println()

	userJSON := `{"id":1,"name":"Alice","active":true,"role":"admin"}`
	user, err := DecodeJSON[User](userJSON)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}
	fmt.Printf("  DecodeJSON[User]     → %+v\n", user)

	ordersJSON := `[{"id":1,"user_id":1,"total":500},{"id":2,"user_id":3,"total":12000}]`
	orders, err := DecodeJSON[[]Order](ordersJSON)
	if err != nil {
		fmt.Println("  error:", err)
		return
	}
	fmt.Printf("  DecodeJSON[[]Order]  → %+v\n", orders)
	fmt.Println()
}

// ---------------------------------------------------------------------------
// 3. Result — 成功/失敗を1つの値として返す（制約は any）
//
// 【なぜジェネリクス？】
//   成功時の中身の型（User など）だけが変わるので、Result[T] にすると共通化できます。
//   普段の Go は (T, error) で十分なことが多いです。補助的な書き方、という位置づけです。
// ---------------------------------------------------------------------------

type Result[T any] struct {
	Value T
	Err   error
}

func Ok[T any](v T) Result[T] {
	return Result[T]{Value: v}
}

func Fail[T any](err error) Result[T] {
	var zero T
	return Result[T]{Value: zero, Err: err}
}

func (r Result[T]) UnwrapOr(fallback T) T {
	if r.Err != nil {
		return fallback
	}
	return r.Value
}

func findUser(users []User, id int) Result[User] {
	for _, u := range users {
		if u.ID == id {
			return Ok(u)
		}
	}
	return Fail[User](fmt.Errorf("user %d not found", id))
}

func demoResult() {
	fmt.Println("=== 3. Result[T any] ===")
	fmt.Println("制約: any（どんな型でもOK）")
	fmt.Println("用途: 検索の成功/失敗を1つの値にまとめる")
	fmt.Println()

	users := []User{
		{ID: 1, Name: "Alice", Active: true, Role: "admin"},
		{ID: 2, Name: "Bob", Active: false, Role: "member"},
	}

	found := findUser(users, 1)
	fmt.Printf("  findUser(1)  → %s\n", found.UnwrapOr(User{Name: "unknown"}).Name)

	missing := findUser(users, 99)
	fmt.Printf("  findUser(99) → fallback=%s\n", missing.UnwrapOr(User{Name: "guest"}).Name)
	if missing.Err != nil {
		fmt.Printf("  エラー詳細: %v\n", missing.Err)
	}
	fmt.Println()
}

// ---------------------------------------------------------------------------
// 4. Set — 重複のない集合（ここで comparable が必要）
//
// 【comparable をざっくり】
//   comparable = 「== で同じか比べられる型」
//
//   OK な例:  int, string, bool, それらだけの struct
//   NG な例:  スライス, マップ, 関数  （これらは == できない）
//
// 【なぜここで必要？】
//   Set の中身は map です。
//   map はキーを探すとき「このキーと同じか？」を比べます。
//   そのためキーの型は comparable である必要があります。
//
//   もし Set[T any] と書くと、コンパイルエラーになります。
//   （any だと「== できない型」も来うるから、map のキーにできない）
//
// 【なぜジェネリクス？】
//   注文ID(int)用・権限名(string)用と別実装にせず、1つの Set で済ませたいからです。
// ---------------------------------------------------------------------------

// Set は要素の重複を許さない集合です。中身は map[T]struct{} です。
type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable](items ...T) *Set[T] {
	s := &Set[T]{m: make(map[T]struct{}, len(items))}
	for _, item := range items {
		s.Add(item)
	}
	return s
}

func (s *Set[T]) Add(v T)      { s.m[v] = struct{}{} }
func (s *Set[T]) Has(v T) bool { _, ok := s.m[v]; return ok }
func (s *Set[T]) Len() int     { return len(s.m) }

func (s *Set[T]) Values() []T {
	out := make([]T, 0, len(s.m))
	for v := range s.m {
		out = append(out, v)
	}
	return out
}

func demoSet() {
	fmt.Println("=== 4. Set[T comparable] ===")
	fmt.Println("制約: comparable（== で比べられる型だけ）")
	fmt.Println("理由: 中で map を使うため。キーは「同じか？」を比べる必要がある")
	fmt.Println("例:   int / string はOK。スライスはNG")
	fmt.Println()

	orderIDs := NewSet(101, 102, 101, 103, 102)
	fmt.Printf("  入力: 101, 102, 101, 103, 102\n")
	fmt.Printf("  → 注文ID数: %d（重複除去後）\n", orderIDs.Len())
	fmt.Printf("  → 103があるか: %v\n", orderIDs.Has(103))

	roles := NewSet("admin", "editor")
	fmt.Printf("  権限 Set: admin, editor\n")
	fmt.Printf("  → admin: %v / viewer: %v\n", roles.Has("admin"), roles.Has("viewer"))
	fmt.Println()
}

// ---------------------------------------------------------------------------
// 5. Cache — キーと値の型を決められるキャッシュ（キーに comparable）
//
// 【制約の使い分け】
//   K comparable … キー。map のキーなので == できる型が必要
//   V any        … 値。ただ保存するだけなのでどんな型でもよい
//
// 【注意】
//   学習用の単純実装です。本番では排他制御や TTL を足すことが多いです。
// ---------------------------------------------------------------------------

type Cache[K comparable, V any] struct {
	data map[K]V
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{data: make(map[K]V)}
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.data[key] = value
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	v, ok := c.data[key]
	return v, ok
}

func demoCache() {
	fmt.Println("=== 5. Cache[K comparable, V any] ===")
	fmt.Println("制約: キー=comparable / 値=any")
	fmt.Println("用途: DBやAPIの結果を一時的に覚えておく")
	fmt.Println()

	userCache := NewCache[int, User]()
	userCache.Set(1, User{ID: 1, Name: "Alice", Active: true, Role: "admin"})

	if u, ok := userCache.Get(1); ok {
		fmt.Printf("  Get(1)  → hit:  %+v\n", u)
	}
	if _, ok := userCache.Get(99); !ok {
		fmt.Println("  Get(99) → miss: id=99 は未登録")
	}

	fmt.Println()
	fmt.Println("--- まとめ ---")
	fmt.Println("any        : どんな型でもOK（まずはこれ）")
	fmt.Println("comparable : == できる型だけ（map のキーに使うとき）")
	fmt.Println("使うとき   : 処理は同じで、型だけ違う")
	fmt.Println("控えるとき : 型ごとに振る舞いが大きく違う → interface の方が向く")
}
