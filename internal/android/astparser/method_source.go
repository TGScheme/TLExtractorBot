package astparser

import (
	sitter "github.com/tree-sitter/go-tree-sitter"
)

func (w *walker) walkSerializeBlock(block *sitter.Node) {
	w.walkBlock(block, 1)
}

func (w *walker) returnTypeFromDeserialize(block *sitter.Node) string {
	var result string
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n == nil || result != "" {
			return
		}
		if n.Kind() == "return_statement" {
			result = w.parseReturnExpr(n.Child(1))
			return
		}
		for i := uint(0); i < n.ChildCount(); i++ {
			walk(n.Child(i))
		}
	}
	walk(block)
	return result
}

func (w *walker) parseReturnExpr(expr *sitter.Node) string {
	if expr == nil {
		return ""
	}
	if expr.Kind() != "method_invocation" {
		return ""
	}
	obj := expr.Child(0)
	name := streamMethodName(expr, w.src)

	if obj != nil && textOf(obj, w.src) == "Vector" && name == "deserialize" {
		elem := w.vectorElementType(expr)
		if elem != "" {
			return "Vector<" + elem + ">"
		}
		return ""
	}

	if name == "TLdeserialize" {
		typeName, hint := scopedName(obj, w.src)
		t, _ := FixTypeName(typeName, hint, false)
		return t
	}
	return ""
}

func (w *walker) vectorElementType(call *sitter.Node) string {
	args := childByKind(call, "argument_list")
	if args == nil {
		return ""
	}
	for i := uint(0); i < args.ChildCount(); i++ {
		c := args.Child(i)
		if c == nil {
			continue
		}
		if c.Kind() == "method_reference" {
			typeName, hint := scopedName(c.Child(0), w.src)
			t, _ := FixTypeName(typeName, hint, false)
			return t
		}
	}
	return ""
}

func scopedName(n *sitter.Node, src []byte) (string, string) {
	if n == nil {
		return "", ""
	}
	if n.Kind() == "field_access" || n.Kind() == "scoped_type_identifier" {
		return textOf(n.Child(n.ChildCount()-1), src), textOf(n.Child(0), src)
	}
	return textOf(n, src), ""
}
