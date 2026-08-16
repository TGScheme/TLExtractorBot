package ast_parser

import (
	"strconv"
	"strings"

	javaTypes "TLExtractor/java/types"
	"TLExtractor/telegram/scheme/types"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

type walker struct {
	astClass *javaTypes.AstClass
	src      []byte
	flags    *flagContext
	params   []types.Parameter
	seen     map[string]bool

	pendingBool []pendingBool
	localAlias  map[string]string
	localValue  map[string]bool

	localField map[string]string

	syntheticFlag map[string]string

	locals map[string]int

	flagInsertIdx map[string]int

	writeSideBools []pendingBool
}

type pendingBool struct {
	name      string
	flagField string
	bit       int
}

func newWalker(astClass *javaTypes.AstClass, src []byte) *walker {
	return &walker{
		astClass:      astClass,
		src:           src,
		flags:         newFlagContext(),
		seen:          make(map[string]bool),
		localAlias:    make(map[string]string),
		localValue:    make(map[string]bool),
		localField:    make(map[string]string),
		syntheticFlag: make(map[string]string),
		locals:        make(map[string]int),
		flagInsertIdx: make(map[string]int),
	}
}

func (w *walker) result() []types.Parameter {
	byField := make(map[string][]types.Parameter)
	var order []string
	for _, pb := range w.pendingBool {
		if w.flags.isBoolGuarded(pb.flagField, pb.bit) || w.seen[pb.name] {
			continue
		}
		if _, ok := byField[pb.flagField]; !ok {
			order = append(order, pb.flagField)
		}
		byField[pb.flagField] = append(byField[pb.flagField], types.Parameter{
			Name: pb.name,
			Type: flagPrefix(pb.flagField, pb.bit, "true"),
		})
	}
	out := make([]types.Parameter, 0, len(w.params)+len(w.pendingBool))
	placed := make(map[string]bool)
	for _, p := range w.params {
		out = append(out, p)
		if p.Type == "#" {
			out = append(out, byField[p.Name]...)
			placed[p.Name] = true
		}
	}

	for _, f := range order {
		if !placed[f] {
			out = append(out, byField[f]...)
		}
	}
	return regroupTrueBools(out)
}

func regroupTrueBools(params []types.Parameter) []types.Parameter {
	flagOf := func(t string) (string, bool) {
		if !strings.HasSuffix(t, "?true") {
			return "", false
		}
		dot := strings.Index(t, ".")
		if dot <= 0 {
			return "", false
		}
		return t[:dot], true
	}
	byField := make(map[string][]types.Parameter)
	var order []string
	rest := make([]types.Parameter, 0, len(params))
	for _, p := range params {
		if ff, ok := flagOf(p.Type); ok {
			if _, seen := byField[ff]; !seen {
				order = append(order, ff)
			}
			byField[ff] = append(byField[ff], p)
		} else {
			rest = append(rest, p)
		}
	}
	if len(byField) == 0 {
		return params
	}
	out := make([]types.Parameter, 0, len(params))
	placed := make(map[string]bool)
	for _, p := range rest {
		out = append(out, p)
		if p.Type == "#" {
			out = append(out, byField[p.Name]...)
			placed[p.Name] = true
		}
	}
	for _, ff := range order {
		if !placed[ff] {
			out = append(out, byField[ff]...)
		}
	}
	return out
}

func (w *walker) emit(name, tlType string) {
	if name == "" || name == "constructor" || w.seen[name] {
		return
	}
	w.seen[name] = true
	w.params = append(w.params, types.Parameter{Name: name, Type: tlType})
}

func (w *walker) ensureFlagField(name string) {
	if name == "" || w.flags.isFlagField(name) {
		return
	}
	w.flags.declareFlagField(name)
	idx, positioned := w.flagInsertIdx[name]
	if !positioned || w.seen[name] || idx > len(w.params) {
		w.emit(name, "#")
		return
	}
	w.seen[name] = true
	w.params = append(w.params, types.Parameter{})
	copy(w.params[idx+1:], w.params[idx:])
	w.params[idx] = types.Parameter{Name: name, Type: "#"}
}

type flagState struct {
	active      bool
	flagField   string
	bit         int
	boolGuarded bool
}

func (w *walker) typed(name, base string, fs flagState) string {
	if fs.active {
		return flagPrefix(fs.flagField, fs.bit, base)
	}
	return base
}

func (w *walker) walkBlock(block *sitter.Node, loopNesting int) {
	if block == nil {
		return
	}
	w.walkStatements(block, loopNesting, flagState{})
}

func (w *walker) walkStatements(block *sitter.Node, loopNesting int, fs flagState) {
	for i := uint(0); i < block.ChildCount(); i++ {
		w.walkStatement(block.Child(i), loopNesting, fs)
	}
}

func (w *walker) walkStatement(stmt *sitter.Node, loopNesting int, fs flagState) {
	if stmt == nil {
		return
	}
	switch stmt.Kind() {
	case "local_variable_declaration":
		w.handleLocalVar(stmt)
	case "expression_statement":
		w.handleExpression(stmt, loopNesting, fs)
	case "if_statement":
		w.handleIf(stmt, loopNesting, fs)
	case "for_statement", "enhanced_for_statement", "while_statement":
		body := childByKind(stmt, "block")
		w.walkStatements(body, loopNesting+1, fs)
	case "try_statement":
		body := childByKind(stmt, "block")
		w.walkStatements(body, loopNesting, fs)
	}
}

func (w *walker) handleLocalVar(stmt *sitter.Node) {
	declarator := childByKind(stmt, "variable_declarator")
	if declarator == nil {
		return
	}
	name := textOf(declarator.Child(0), w.src)
	val := declarator.Child(2)
	if val == nil {
		return
	}
	if _, dup := w.locals[name]; !dup {
		w.locals[name] = len(w.params)
	}

	if val.Kind() == "method_invocation" {
		mName := streamMethodName(val, w.src)
		if prim, ok := parseStreamPrimitive(mName); ok && prim == "int" {
			w.localAlias[name] = name
		}
		obj := val.Child(0)
		_, isPrim := parseStreamPrimitive(mName)
		if mName == "TLdeserialize" ||
			(obj != nil && textOf(obj, w.src) == "Vector" && strings.HasPrefix(mName, "deserialize")) ||
			isPrim {
			w.localValue[name] = true
		}

		if ff, bit, ok := w.parseHasFlagCall(val); ok {
			w.flags.mapBool(name, ff, bit)
		}

		if bools, flagField, ok := w.collectInlineSetFlags(val); ok {
			w.localAlias[name] = flagField
			for _, b := range bools {
				w.registerWriteSideBool("", b.name, flagField, b.bit)
			}
		} else if rawBase, ok := parseSetFlagBase(val, w.src); ok {
			flagField := rawBase
			if ff, aliasOK := w.localAlias[rawBase]; aliasOK {
				flagField = ff
			}
			w.localAlias[name] = flagField
		}
	}

	if val.Kind() == "field_access" {
		if field := fieldNameFromLHS(val, w.src); field != "" {
			w.localField[name] = field
		}
	}

	if val.Kind() == "binary_expression" {
		if ff, bit, ok := parseMaskNotZero(val, w.src); ok {
			w.flags.mapBool(name, w.resolveFlagField(ff), bit)
		} else if ff, bit, ok := parsePlainAndMask(val, w.src); ok {
			w.flags.mapBool(name, w.resolveFlagField(ff), bit)
		}
	}

	if val.Kind() == "ternary_expression" {
		if bn, rawBase, bit, ok := parseWriteSideTernary(val, w.src); ok {
			w.registerWriteSideBool(name, bn, rawBase, bit)
		}
	}
}

func (w *walker) registerWriteSideBool(localName, boolName, rawBase string, bit int) {
	flagField := rawBase
	if ff, aliasOK := w.localAlias[rawBase]; aliasOK {
		flagField = ff
	}
	if localName != "" {
		w.localAlias[localName] = flagField
	}
	field, ok := w.resolveBoolName(boolName)
	if !ok {
		return
	}
	w.flags.mapBool(field, flagField, bit)
	w.writeSideBools = append(w.writeSideBools, pendingBool{
		name: field, flagField: flagField, bit: bit,
	})
}

func (w *walker) resolveFlagField(name string) string {
	if w.flags.isFlagField(name) {
		return name
	}
	if real, ok := w.localAlias[name]; ok && w.flags.isFlagField(real) {
		return real
	}
	if canon, ok := w.syntheticFlag[name]; ok {
		return canon
	}

	if idx, isLocal := w.locals[name]; isLocal {
		canon := w.nextFlagName()
		w.syntheticFlag[name] = canon
		w.flagInsertIdx[canon] = idx
		return canon
	}
	if first := w.flags.firstFlagField(); first != "" {
		return first
	}
	return name
}

func (w *walker) nextFlagName() string {
	taken := func(n string) bool {
		if w.flags.isFlagField(n) {
			return true
		}
		for _, v := range w.syntheticFlag {
			if v == n {
				return true
			}
		}
		return false
	}
	if !taken("flags") {
		return "flags"
	}
	for i := 2; ; i++ {
		if candidate := "flags" + strconv.Itoa(i); !taken(candidate) {
			return candidate
		}
	}
}

func (w *walker) resolveBoolName(name string) (string, bool) {
	if field, ok := w.localField[name]; ok {
		name = field
	}
	if _, ok := fieldVar(w.astClass, name); !ok {
		return "", false
	}
	return name, true
}

func (w *walker) handleExpression(stmt *sitter.Node, loopNesting int, fs flagState) {
	if loopNesting > 0 {
		if w.trySerializeFlagsWrite(stmt) {
			return
		}
		if name, base, ok := w.classifyLoopElement(stmt); ok {
			w.emit(name, w.typed(name, base, fs))
		}
		return
	}
	info := classifyAssignment(stmt, w.src)
	switch info.Kind {
	case stmtReadField:
		w.emit(info.Field, w.typed(info.Field, info.Prim, fs))
	case stmtObjectField, stmtVectorDeserialize:
		base, ok := fieldType(w.astClass, info.Field)
		if info.ObjType != "" {
			base, ok = info.ObjType, true
		}
		if ok {
			w.emit(info.Field, w.typed(info.Field, base, fs))
		}
	case stmtFlagFieldRead:
		w.flags.declareFlagField(info.Field)
		w.emit(info.Field, "#")
	case stmtBoolFromMask:

		w.flags.mapBool(info.BoolVar, w.resolveFlagField(info.FlagField), info.Bit)
		w.pendingBool = append(w.pendingBool, pendingBool{
			name: info.Field, flagField: w.resolveFlagField(info.FlagField),
			bit: info.Bit,
		})
	case stmtOptionalTernary:
		flagField := w.resolveFlagField(info.FlagField)
		w.ensureFlagField(flagField)
		w.emit(info.Field, flagPrefix(flagField, info.Bit, info.Prim))
	default:

		if !w.tryFlagAssignFromAlias(stmt) {
			if !w.tryBoolFromMaskLocal(stmt) {
				if !w.tryHasFlagAssign(stmt, fs) {
					if !w.tryHasFlagTernary(stmt) {
						w.tryHoistedFieldAssign(stmt, fs)
					}
				}
			}
		}
	}
}

func (w *walker) trySerializeFlagsWrite(stmt *sitter.Node) bool {
	expr := stmt.Child(0)
	if expr == nil || expr.Kind() != "method_invocation" {
		return false
	}
	if streamMethodName(expr, w.src) != "writeInt32" {
		return false
	}
	args := childByKind(expr, "argument_list")
	if args == nil {
		return false
	}
	arg := args.Child(1)
	if arg == nil {
		return false
	}
	var flagField string
	switch arg.Kind() {
	case "field_access":

		name := fieldNameFromLHS(arg, w.src)
		if name == "flags" || strings.HasPrefix(name, "flags") {
			flagField = name
		}
	case "identifier":

		name := textOf(arg, w.src)
		if ff, ok := w.localAlias[name]; ok && (ff == "flags" || strings.HasPrefix(ff, "flags")) {
			flagField = ff
		}
	case "method_invocation":

		if bools, ff, ok := w.collectInlineSetFlags(arg); ok {
			flagField = ff
			for _, b := range bools {
				w.registerWriteSideBool("", b.name, ff, b.bit)
			}
		}
	}
	if flagField == "" {
		return false
	}

	w.ensureFlagField(flagField)

	for _, pb := range w.writeSideBools {
		if pb.flagField == flagField {
			w.pendingBool = append(w.pendingBool, pendingBool{
				name: pb.name, flagField: pb.flagField, bit: pb.bit,
			})
		}
	}

	remaining := w.writeSideBools[:0]
	for _, pb := range w.writeSideBools {
		if pb.flagField != flagField {
			remaining = append(remaining, pb)
		}
	}
	w.writeSideBools = remaining
	return true
}

func (w *walker) collectInlineSetFlags(call *sitter.Node) (bools []pendingBool, flagField string, ok bool) {
	node := call
	for node != nil && node.Kind() == "method_invocation" && streamMethodName(node, w.src) == "setFlag" {
		if bn, _, bit, okc := parseSetFlagCall(node, w.src); okc {
			bools = append(bools, pendingBool{name: bn, bit: bit})
		}
		args := childByKind(node, "argument_list")
		if args == nil {
			break
		}
		var base *sitter.Node
		for i := uint(0); i < args.ChildCount(); i++ {
			c := args.Child(i)
			if c == nil {
				continue
			}
			switch c.Kind() {
			case "(", ")", ",":
			default:
				base = c
			}
			if base != nil {
				break
			}
		}
		if base == nil {
			break
		}
		if base.Kind() == "method_invocation" && streamMethodName(base, w.src) == "setFlag" {
			node = base
			continue
		}
		switch base.Kind() {
		case "field_access":
			flagField = fieldNameFromLHS(base, w.src)
		case "identifier":
			nm := textOf(base, w.src)
			if ff, a := w.localAlias[nm]; a {
				flagField = ff
			} else {
				flagField = nm
			}
		default:
			flagField = "flags"
		}
		break
	}
	if len(bools) == 0 {
		return nil, "", false
	}
	if flagField == "" {
		flagField = "flags"
	}

	for i, j := 0, len(bools)-1; i < j; i, j = i+1, j-1 {
		bools[i], bools[j] = bools[j], bools[i]
	}
	return bools, flagField, true
}

func (w *walker) tryFlagAssignFromAlias(stmt *sitter.Node) bool {
	assign := stmt
	if stmt.Kind() == "expression_statement" {
		assign = stmt.Child(0)
	}
	if assign == nil || assign.Kind() != "assignment_expression" {
		return false
	}
	lhs := assign.Child(0)
	rhs := assign.Child(2)
	field := fieldNameFromLHS(lhs, w.src)
	if field == "" || rhs == nil || rhs.Kind() != "identifier" {
		return false
	}
	if !(field == "flags" || strings.HasPrefix(field, "flags")) {
		return false
	}
	aliasName := textOf(rhs, w.src)
	if _, isAlias := w.localAlias[aliasName]; !isAlias {
		return false
	}

	w.localAlias[aliasName] = field
	w.flags.declareFlagField(field)
	w.emit(field, "#")
	return true
}

func (w *walker) tryBoolFromMaskLocal(stmt *sitter.Node) bool {
	assign := stmt
	if stmt.Kind() == "expression_statement" {
		assign = stmt.Child(0)
	}
	if assign == nil || assign.Kind() != "assignment_expression" {
		return false
	}
	lhs := assign.Child(0)
	rhs := assign.Child(2)
	field := fieldNameFromLHS(lhs, w.src)
	if field == "" || rhs == nil {
		return false
	}

	var source *sitter.Node
	switch rhs.Kind() {
	case "identifier":
		source = rhs
	case "binary_expression":
		c0, c1, c2 := rhs.Child(0), rhs.Child(1), rhs.Child(2)
		if c0 == nil || c0.Kind() != "identifier" || textOf(c1, w.src) != "!=" || textOf(c2, w.src) != "0" {
			return false
		}
		source = c0
	default:
		return false
	}
	ff, bit, ok := w.flags.lookupBool(textOf(source, w.src))
	if !ok {
		return false
	}
	w.flags.mapBool(field, ff, bit)
	w.pendingBool = append(w.pendingBool, pendingBool{
		name: field, flagField: ff, bit: bit,
	})
	return true
}

func (w *walker) parseHasFlagCall(call *sitter.Node) (string, int, bool) {
	if call == nil || call.Kind() != "method_invocation" {
		return "", 0, false
	}
	if streamMethodName(call, w.src) != "hasFlag" {
		return "", 0, false
	}
	args := childByKind(call, "argument_list")
	if args == nil {
		return "", 0, false
	}

	flagsNode := args.Child(1)
	maskNode := args.Child(3)
	if flagsNode == nil || maskNode == nil {
		return "", 0, false
	}
	bit, ok := bitOfMask(textOf(maskNode, w.src))
	if !ok {
		return "", 0, false
	}
	var field string
	switch flagsNode.Kind() {
	case "field_access":
		field = fieldNameFromLHS(flagsNode, w.src)
	case "identifier":
		field = textOf(flagsNode, w.src)
	default:
		field = "flags"
	}
	if field == "" {
		field = "flags"
	}
	return w.resolveFlagField(field), bit, true
}

func (w *walker) tryHasFlagAssign(stmt *sitter.Node, fs flagState) bool {
	assign := stmt.Child(0)
	if assign == nil || assign.Kind() != "assignment_expression" {
		return false
	}
	lhs := assign.Child(0)
	rhs := assign.Child(2)
	field := fieldNameFromLHS(lhs, w.src)
	if field == "" || rhs == nil || rhs.Kind() != "method_invocation" {
		return false
	}
	flagField, bit, ok := w.parseHasFlagCall(rhs)
	if !ok {
		return false
	}
	w.ensureFlagField(flagField)
	base := "true"
	if t, okT := fieldType(w.astClass, field); okT && t != "Bool" {
		base = t
	}
	w.emit(field, flagPrefix(flagField, bit, base))
	return true
}

func (w *walker) tryHasFlagTernary(stmt *sitter.Node) bool {
	assign := stmt.Child(0)
	if assign == nil || assign.Kind() != "assignment_expression" {
		return false
	}
	field := fieldNameFromLHS(assign.Child(0), w.src)
	tern := assign.Child(2)
	if field == "" || tern == nil || tern.Kind() != "ternary_expression" {
		return false
	}
	flagField, bit, ok := w.parseHasFlagCall(tern.Child(0))
	if !ok {
		return false
	}
	thenE := tern.Child(2)
	if thenE == nil || thenE.Kind() != "method_invocation" {
		return false
	}
	base, ok := parseStreamPrimitive(streamMethodName(thenE, w.src))
	if !ok {
		return false
	}
	w.ensureFlagField(flagField)
	w.emit(field, flagPrefix(flagField, bit, base))
	return true
}

func (w *walker) tryHoistedFieldAssign(stmt *sitter.Node, fs flagState) bool {
	assign := stmt.Child(0)
	if assign == nil || assign.Kind() != "assignment_expression" {
		return false
	}
	lhs := assign.Child(0)
	rhs := assign.Child(2)
	if rhs == nil || rhs.Kind() != "identifier" {
		return false
	}
	field := fieldNameFromLHS(lhs, w.src)
	if field == "" || !w.localValue[textOf(rhs, w.src)] {
		return false
	}
	base, ok := fieldType(w.astClass, field)
	if !ok {
		return false
	}
	w.emit(field, w.typed(field, base, fs))
	return true
}

func (w *walker) classifyLoopElement(stmt *sitter.Node) (string, string, bool) {
	expr := stmt.Child(0)
	if expr == nil || expr.Kind() != "method_invocation" {
		return "", "", false
	}
	name := streamMethodName(expr, w.src)
	obj := expr.Child(0)
	switch name {
	case "add":
		field := w.fieldFromThisChain(obj)
		if field == "" {
			return "", "", false
		}

		if base, ok := fieldType(w.astClass, field); ok {
			return field, base, true
		}

		if prim, ok := w.primitiveOfAddArg(expr); ok {
			return field, "Vector<" + prim + ">", true
		}
	case "serializeToStream":

		field := w.fieldFromThisChain(obj)
		if field == "" {
			return "", "", false
		}
		if base, ok := fieldType(w.astClass, field); ok {
			return field, base, true
		}
		return "", "", false
	case "serialize", "serializeInt", "serializeLong", "serializeString", "serializeBytes", "serializeByteArray", "serializeDouble", "serializeBool":

		field := w.fieldFromArg(expr)
		if field == "" {
			return "", "", false
		}
		if base, ok := fieldType(w.astClass, field); ok {
			return field, base, true
		}
		return "", "", false
	default:

		if prim, ok := parseStreamPrimitive(name); ok {
			field := w.fieldFromArg(expr)
			if field == "" {
				return "", "", false
			}
			if base, ok2 := fieldType(w.astClass, field); ok2 {
				return field, base, true
			}
			return field, "Vector<" + prim + ">", true
		}
	}
	return "", "", false
}

func (w *walker) fieldFromThisChain(obj *sitter.Node) string {
	if obj != nil && obj.Kind() == "identifier" {
		return w.localField[textOf(obj, w.src)]
	}
	return fieldNameFromLHS(obj, w.src)
}

func (w *walker) primitiveOfAddArg(call *sitter.Node) (string, bool) {
	args := childByKind(call, "argument_list")
	if args == nil {
		return "", false
	}
	var found string
	var ok bool
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n == nil || ok {
			return
		}
		if n.Kind() == "method_invocation" {
			if prim, p := parseStreamPrimitive(streamMethodName(n, w.src)); p {
				found, ok = prim, true
				return
			}
		}
		for i := uint(0); i < n.ChildCount(); i++ {
			walk(n.Child(i))
		}
	}
	walk(args)
	return found, ok
}

func (w *walker) fieldFromArg(call *sitter.Node) string {
	args := childByKind(call, "argument_list")
	if args == nil {
		return ""
	}
	var field string
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n == nil || field != "" {
			return
		}
		if n.Kind() == "field_access" {
			if name := fieldNameFromLHS(n, w.src); name != "" && name != "get" {
				if _, ok := fieldVar(w.astClass, name); ok {
					field = name
					return
				}
			}
		}
		for i := uint(0); i < n.ChildCount(); i++ {
			walk(n.Child(i))
		}
	}
	walk(args)
	return field
}

func (w *walker) handleIf(stmt *sitter.Node, loopNesting int, fs flagState) {
	cond := childByKind(stmt, "parenthesized_expression")
	body := childByKind(stmt, "block")
	if cond == nil {
		return
	}
	inner := cond.Child(1)
	newFS, isMagic := w.flagStateFromCond(inner)

	if isMagic {
		w.walkAllBlocks(stmt, loopNesting, fs)
		return
	}
	if newFS.active {
		w.ensureFlagField(newFS.flagField)
		if newFS.boolGuarded {
			w.flags.markBoolGuard(newFS.flagField, newFS.bit)
		}
		w.walkStatements(body, loopNesting, newFS)
		return
	}

	w.walkStatements(body, loopNesting, fs)
}

func (w *walker) walkAllBlocks(stmt *sitter.Node, loopNesting int, fs flagState) {
	for i := uint(0); i < stmt.ChildCount(); i++ {
		c := stmt.Child(i)
		if c == nil {
			continue
		}
		if c.Kind() == "block" {
			w.walkStatements(c, loopNesting, fs)
		}
		if c.Kind() == "if_statement" {
			w.handleIf(c, loopNesting, fs)
		}
	}
}

func (w *walker) flagStateFromCond(inner *sitter.Node) (flagState, bool) {
	if inner == nil {
		return flagState{}, false
	}
	text := textOf(inner, w.src)

	if contains481674261(text) {
		return flagState{}, true
	}
	switch inner.Kind() {
	case "binary_expression":
		if ff, bit, ok := parseMaskNotZero(inner, w.src); ok {
			return flagState{active: true, flagField: w.resolveFlagField(ff), bit: bit, boolGuarded: false}, false
		}

		c0, c1, c2 := inner.Child(0), inner.Child(1), inner.Child(2)
		if c0 != nil && c0.Kind() == "identifier" && textOf(c1, w.src) == "!=" && textOf(c2, w.src) == "0" {
			if ff, bit, ok := w.flags.lookupBool(textOf(c0, w.src)); ok {
				return flagState{active: true, flagField: ff, bit: bit, boolGuarded: false}, false
			}
		}
	case "identifier":
		if ff, bit, ok := w.flags.lookupBool(textOf(inner, w.src)); ok {
			return flagState{active: true, flagField: ff, bit: bit, boolGuarded: true}, false
		}
	case "method_invocation":
		if flagField, bit, ok := w.parseHasFlagCall(inner); ok {
			return flagState{active: true, flagField: flagField, bit: bit, boolGuarded: false}, false
		}
	}
	return flagState{}, false
}

func contains481674261(s string) bool {
	return indexOf(s, "481674261") >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
