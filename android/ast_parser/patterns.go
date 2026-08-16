package ast_parser

import (
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

type stmtKind int

const (
	stmtUnknown stmtKind = iota
	stmtReadField
	stmtVectorDeserialize
	stmtFlagFieldRead
	stmtBoolFromMask
	stmtObjectField
	stmtOptionalTernary
)

type stmtInfo struct {
	Kind      stmtKind
	Field     string
	Prim      string
	FlagField string
	Bit       int
	BoolVar   string
	ObjType   string
}

func textOf(n *sitter.Node, src []byte) string {
	if n == nil {
		return ""
	}
	return n.Utf8Text(src)
}

func childByKind(n *sitter.Node, kind string) *sitter.Node {
	if n == nil {
		return nil
	}
	for i := uint(0); i < n.ChildCount(); i++ {
		c := n.Child(i)
		if c != nil && c.Kind() == kind {
			return c
		}
	}
	return nil
}

func fieldNameFromLHS(n *sitter.Node, src []byte) string {
	if n == nil || n.Kind() != "field_access" {
		return ""
	}
	name := n.Child(2)
	return textOf(name, src)
}

func streamMethodName(call *sitter.Node, src []byte) string {
	if call == nil || call.Kind() != "method_invocation" {
		return ""
	}

	if call.Child(0) != nil && call.Child(0).Kind() != "identifier" {
		return textOf(call.Child(2), src)
	}

	return textOf(call.Child(2), src)
}

func classifyAssignment(stmt *sitter.Node, src []byte) stmtInfo {
	if stmt == nil {
		return stmtInfo{}
	}
	assign := stmt
	if stmt.Kind() == "expression_statement" {
		assign = stmt.Child(0)
	}
	if assign == nil || assign.Kind() != "assignment_expression" {
		return stmtInfo{}
	}
	lhs := assign.Child(0)
	rhs := assign.Child(2)
	field := fieldNameFromLHS(lhs, src)
	if field == "" || rhs == nil {
		return stmtInfo{}
	}

	switch rhs.Kind() {
	case "method_invocation":
		return classifyInvocationRHS(field, rhs, src)
	case "binary_expression":

		if ff, bit, ok := parseMaskNotZero(rhs, src); ok {
			return stmtInfo{Kind: stmtBoolFromMask, Field: field, BoolVar: field, FlagField: ff, Bit: bit}
		}
	case "ternary_expression":
		if info, ok := parseOptionalTernary(field, rhs, src); ok {
			return info
		}
	}
	return stmtInfo{}
}

var boxedPrimitives = map[string]bool{
	"Boolean": true, "Integer": true, "Long": true,
	"Double": true, "Float": true, "Short": true, "Byte": true,
}

func unboxValueOf(call *sitter.Node, src []byte) *sitter.Node {
	if streamMethodName(call, src) != "valueOf" {
		return nil
	}
	obj := call.Child(0)
	if obj == nil || !boxedPrimitives[textOf(obj, src)] {
		return nil
	}
	args := childByKind(call, "argument_list")
	if args == nil {
		return nil
	}
	for i := uint(0); i < args.ChildCount(); i++ {
		if c := args.Child(i); c != nil && c.Kind() == "method_invocation" {
			return c
		}
	}
	return nil
}

func classifyInvocationRHS(field string, call *sitter.Node, src []byte) stmtInfo {
	if inner := unboxValueOf(call, src); inner != nil {
		return classifyInvocationRHS(field, inner, src)
	}
	obj := call.Child(0)
	name := streamMethodName(call, src)

	if obj != nil && textOf(obj, src) == "Vector" && strings.HasPrefix(name, "deserialize") {
		return stmtInfo{Kind: stmtVectorDeserialize, Field: field}
	}

	if name == "TLdeserialize" {
		info := stmtInfo{Kind: stmtObjectField, Field: field}

		tn, hint := scopedName(obj, src)
		if hint == "" && !strings.HasPrefix(tn, "TL_") {
			if t, _ := FixTypeName(tn, "", false); t != "" {
				info.ObjType = t
			}
		}
		return info
	}

	if prim, ok := parseStreamPrimitive(name); ok {
		if field == "flags" || strings.HasPrefix(field, "flags") {
			if prim == "int" {
				return stmtInfo{Kind: stmtFlagFieldRead, Field: field}
			}
		}
		return stmtInfo{Kind: stmtReadField, Field: field, Prim: prim}
	}
	return stmtInfo{}
}

func fieldAndMask(fieldNode, maskNode *sitter.Node, src []byte) (string, int, bool) {
	bit, ok := bitOfMask(textOf(maskNode, src))
	if !ok {
		return "", 0, false
	}
	field := fieldNameFromLHS(fieldNode, src)
	if field == "" {
		field = textOf(fieldNode, src)
	}
	return field, bit, true
}

func parseMaskNotZero(bin *sitter.Node, src []byte) (string, int, bool) {
	left := bin.Child(0)
	if left == nil {
		return "", 0, false
	}
	if left.Kind() == "parenthesized_expression" {
		left = left.Child(1)
	}
	if left == nil || left.Kind() != "binary_expression" {
		return "", 0, false
	}
	op := left.Child(1)
	if textOf(op, src) != "&" {
		return "", 0, false
	}
	c0 := left.Child(0)
	c2 := left.Child(2)

	if field, bit, ok := fieldAndMask(c0, c2, src); ok {
		return field, bit, true
	}
	if field, bit, ok := fieldAndMask(c2, c0, src); ok {
		return field, bit, true
	}
	return "", 0, false
}

func parsePlainAndMask(bin *sitter.Node, src []byte) (string, int, bool) {
	if bin == nil {
		return "", 0, false
	}
	op := bin.Child(1)
	if textOf(op, src) != "&" {
		return "", 0, false
	}
	c0, c2 := bin.Child(0), bin.Child(2)
	if field, bit, ok := fieldAndMask(c0, c2, src); ok {
		return field, bit, true
	}
	if field, bit, ok := fieldAndMask(c2, c0, src); ok {
		return field, bit, true
	}
	return "", 0, false
}

func parseOptionalTernary(field string, tern *sitter.Node, src []byte) (stmtInfo, bool) {
	cond := tern.Child(0)
	thenE := tern.Child(2)
	if cond == nil || thenE == nil {
		return stmtInfo{}, false
	}
	ff, bit, ok := parseMaskNotZero(cond, src)
	if !ok {
		return stmtInfo{}, false
	}
	if thenE.Kind() == "method_invocation" {
		if prim, ok := parseStreamPrimitive(streamMethodName(thenE, src)); ok {
			return stmtInfo{Kind: stmtOptionalTernary, Field: field, Prim: prim, FlagField: ff, Bit: bit}, true
		}
	}
	return stmtInfo{}, false
}

func parseWriteSideTernary(tern *sitter.Node, src []byte) (boolName, rawBase string, bit int, ok bool) {
	if tern == nil || tern.Kind() != "ternary_expression" {
		return "", "", 0, false
	}
	cond := tern.Child(0)
	thenE := tern.Child(2)
	if thenE == nil || thenE.Kind() != "binary_expression" {
		return "", "", 0, false
	}
	if textOf(thenE.Child(1), src) != "|" {
		return "", "", 0, false
	}
	left, right := thenE.Child(0), thenE.Child(2)

	var maskNode, baseNode *sitter.Node
	if right != nil && right.Kind() == "decimal_integer_literal" {
		maskNode = right
		baseNode = left
	} else if left != nil && left.Kind() == "decimal_integer_literal" {
		maskNode = left
		baseNode = right
	}
	if maskNode == nil || baseNode == nil {
		return "", "", 0, false
	}
	b, okBit := bitOfMask(textOf(maskNode, src))
	if !okBit {
		return "", "", 0, false
	}

	var bn string
	if cond != nil {
		switch cond.Kind() {
		case "field_access":
			bn = fieldNameFromLHS(cond, src)
		case "identifier":
			bn = textOf(cond, src)
		}
	}
	if bn == "" {
		return "", "", 0, false
	}

	var base string
	switch baseNode.Kind() {
	case "field_access":
		base = fieldNameFromLHS(baseNode, src)
	case "identifier":
		base = textOf(baseNode, src)
	}
	if base == "" {
		base = "flags"
	}
	return bn, base, b, true
}

func parseSetFlagCall(call *sitter.Node, src []byte) (boolName, rawBase string, bit int, ok bool) {
	if call == nil || call.Kind() != "method_invocation" {
		return "", "", 0, false
	}
	if streamMethodName(call, src) != "setFlag" {
		return "", "", 0, false
	}
	args := childByKind(call, "argument_list")
	if args == nil {
		return "", "", 0, false
	}

	var real []*sitter.Node
	for i := uint(0); i < args.ChildCount(); i++ {
		c := args.Child(i)
		if c == nil {
			continue
		}
		switch c.Kind() {
		case "(", ")", ",":
		default:
			real = append(real, c)
		}
	}
	if len(real) != 3 {
		return "", "", 0, false
	}
	b, okBit := bitOfMask(textOf(real[1], src))
	if !okBit {
		return "", "", 0, false
	}
	switch real[0].Kind() {
	case "field_access":
		rawBase = fieldNameFromLHS(real[0], src)
	case "identifier":
		rawBase = textOf(real[0], src)
	}
	if rawBase == "" {
		rawBase = "flags"
	}
	switch real[2].Kind() {
	case "field_access":
		boolName = fieldNameFromLHS(real[2], src)
	case "identifier":
		boolName = textOf(real[2], src)
	}
	if boolName == "" {
		return "", "", 0, false
	}
	return boolName, rawBase, b, true
}

func parseSetFlagBase(call *sitter.Node, src []byte) (rawBase string, ok bool) {
	if call == nil || call.Kind() != "method_invocation" {
		return "", false
	}
	if streamMethodName(call, src) != "setFlag" {
		return "", false
	}
	args := childByKind(call, "argument_list")
	if args == nil {
		return "", false
	}
	var real []*sitter.Node
	for i := uint(0); i < args.ChildCount(); i++ {
		c := args.Child(i)
		if c == nil {
			continue
		}
		switch c.Kind() {
		case "(", ")", ",":
		default:
			real = append(real, c)
		}
	}
	if len(real) != 3 {
		return "", false
	}
	switch real[0].Kind() {
	case "field_access":
		rawBase = fieldNameFromLHS(real[0], src)
	case "identifier":
		rawBase = textOf(real[0], src)
	}
	if rawBase == "" {
		rawBase = "flags"
	}
	return rawBase, true
}

func parseMethodBlock(language *sitter.Language, body string) (*sitter.Node, []byte, func()) {
	wrapper := "class T { void m() {\n" + body + "\n} }"
	srcBytes := []byte(wrapper)
	parser := sitter.NewParser()
	_ = parser.SetLanguage(language)
	tree := parser.Parse(srcBytes, nil)
	closer := func() { tree.Close(); parser.Close() }
	root := tree.RootNode()

	var block *sitter.Node
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n == nil || block != nil {
			return
		}
		if n.Kind() == "block" {
			block = n
			return
		}
		for i := uint(0); i < n.ChildCount(); i++ {
			walk(n.Child(i))
		}
	}
	walk(root)
	return block, srcBytes, closer
}
