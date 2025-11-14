package setlang

// Expr is the top-level expression node.
// Grammar: UnionExpr
type Expr struct {
	Union *UnionExpr `@@`
}

// UnionExpr represents union operations (|).
// Grammar: IntersectExpr ("|" IntersectExpr)*
type UnionExpr struct {
	Left  *IntersectExpr `@@`
	Right []*UnionTail   `@@*`
}

// UnionTail represents the right-hand side of a union operation.
type UnionTail struct {
	Op    string         `@"|"`
	Right *IntersectExpr `@@`
}

// IntersectExpr represents intersection operations (&).
// Grammar: DiffExpr ("&" DiffExpr)*
type IntersectExpr struct {
	Left  *DiffExpr        `@@`
	Right []*IntersectTail `@@*`
}

// IntersectTail represents the right-hand side of an intersection operation.
type IntersectTail struct {
	Op    string    `@"&"`
	Right *DiffExpr `@@`
}

// DiffExpr represents difference operations (~).
// Grammar: Primary ("~" Primary)*
type DiffExpr struct {
	Left  *Primary    `@@`
	Right []*DiffTail `@@*`
}

// DiffTail represents the right-hand side of a difference operation.
type DiffTail struct {
	Op    string   `@"~"`
	Right *Primary `@@`
}

// Primary represents atomic expressions: function calls, identifiers, or parenthesized expressions.
type Primary struct {
	FuncCall *FuncCall `  @@`
	Ident    *string   `| @Ident`
	SubExpr  *Expr     `| "(" @@ ")"`
}

// FuncCall represents a function call with arguments.
// Grammar: Ident "(" (Arg ("," Arg)*)? ")"
type FuncCall struct {
	Name string `@Ident`
	Args []*Arg `"(" ( @@ ( "," @@ )* )? ")"`
}

// Arg represents a function argument, which can be either:
// - A nested expression (for functions that take set expressions as arguments)
// - A string literal (for functions that take string parameters)
// - An identifier (for simple arguments)
type Arg struct {
	Expr   *Expr   `  @@`
	StrLit *string `| @String`
	Ident  *string `| @Ident`
}

// Node is an interface that all AST nodes implement.
// This allows for uniform handling of different node types during evaluation.
type Node interface {
	node()
}

func (*Expr) node()          {}
func (*UnionExpr) node()     {}
func (*UnionTail) node()     {}
func (*IntersectExpr) node() {}
func (*IntersectTail) node() {}
func (*DiffExpr) node()      {}
func (*DiffTail) node()      {}
func (*Primary) node()       {}
func (*FuncCall) node()      {}
func (*Arg) node()           {}
