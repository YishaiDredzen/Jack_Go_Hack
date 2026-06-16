package codegen

import (
	"fmt"
	"strconv"

	"my_compiler/parser"
)

type CodeGenerator struct {
	symbols   *SymbolTable
	writer    *VMWriter
	className string
	labelCount int // for generating unique if/while labels
}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{
		symbols: NewSymbolTable(),
		writer:  NewVMWriter(),
	}
}

// Generate is the entry point — pass in the root class node
func (cg *CodeGenerator) Generate(classNode *parser.Node) string {
	cg.compileClass(classNode)
	return cg.writer.String()
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

// childrenOfType returns all direct children whose Type matches
func childrenOfType(node *parser.Node, typ string) []*parser.Node {
	var result []*parser.Node
	for _, c := range node.Children {
		if c.Type == typ {
			result = append(result, c)
		}
	}
	return result
}

// firstChildOfType returns the first direct child matching typ
func firstChildOfType(node *parser.Node, typ string) *parser.Node {
	for _, c := range node.Children {
		if c.Type == typ {
			return c
		}
	}
	return nil
}

// childValue returns the Value of a direct child at position i
func childValue(node *parser.Node, i int) string {
	if i < len(node.Children) {
		return node.Children[i].Value
	}
	return ""
}

func (cg *CodeGenerator) uniqueLabel(prefix string) string {
	label := fmt.Sprintf("%s_%s_%d", cg.className, prefix, cg.labelCount)
	cg.labelCount++
	return label
}

// ─────────────────────────────────────────────
// 1. Class
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileClass(node *parser.Node) {
	// children: 'class', className, '{', [classVarDec*], [subroutineDec*], '}'
	cg.className = node.Children[1].Value

	for _, child := range node.Children {
		switch child.Type {
		case "classVarDec":
			cg.compileClassVarDec(child)
		case "subroutineDec":
			cg.compileSubroutine(child)
		}
	}
}

// ─────────────────────────────────────────────
// 2. Class Variable Declarations
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileClassVarDec(node *parser.Node) {
	// children: ('static'|'field'), type, varName (, varName)*  ';'
	kindStr := node.Children[0].Value
	typ := node.Children[1].Value

	var kind Kind
	if kindStr == "static" {
		kind = KindStatic
	} else {
		kind = KindField
	}

	// Every identifier child after the type is a variable name
	for i := 2; i < len(node.Children); i++ {
		if node.Children[i].Type == "identifier" {
			cg.symbols.Define(node.Children[i].Value, typ, kind)
		}
	}
}

// ─────────────────────────────────────────────
// 3. Subroutine
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileSubroutine(node *parser.Node) {
	// children: ('constructor'|'function'|'method'), returnType, name, '(', parameterList, ')', subroutineBody
	cg.symbols.StartSubroutine()

	subroutineKind := node.Children[0].Value // constructor | function | method
	name := node.Children[2].Value

	// Methods receive 'this' as argument 0
	if subroutineKind == "method" {
		cg.symbols.Define("this", cg.className, KindArg)
	}

	// Parameter list
	paramList := firstChildOfType(node, "parameterList")
	if paramList != nil {
		cg.compileParameterList(paramList)
	}

	// Subroutine body
	body := firstChildOfType(node, "subroutineBody")
	if body == nil {
		return
	}

	// Count local vars first so we can emit the function declaration
	nLocals := 0
	for _, child := range body.Children {
		if child.Type == "varDec" {
			for _, c := range child.Children {
				if c.Type == "identifier" {
					nLocals++
				}
			}
		}
	}

	funcName := fmt.Sprintf("%s.%s", cg.className, name)
	cg.writer.WriteFunction(funcName, nLocals)

	switch subroutineKind {
	case "constructor":
		// Allocate memory for the object: push nFields, call Memory.alloc 1, pop pointer 0
		nFields := cg.symbols.VarCount(KindField)
		cg.writer.WritePush("constant", nFields)
		cg.writer.WriteCall("Memory.alloc", 1)
		cg.writer.WritePop("pointer", 0)

	case "method":
		// argument 0 is 'this' — anchor THIS to it
		cg.writer.WritePush("argument", 0)
		cg.writer.WritePop("pointer", 0)

	case "function":
		// Nothing special needed
	}

	// Compile the body (varDecs already counted, now compile statements)
	cg.compileSubroutineBody(body)
}

// ─────────────────────────────────────────────
// 4. Parameter List
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileParameterList(node *parser.Node) {
	// children: type varName (, type varName)*
	i := 0
	for i < len(node.Children) {
		if node.Children[i].Value == "," {
			i++
			continue
		}
		// type varName pair
		if i+1 < len(node.Children) {
			typ := node.Children[i].Value
			name := node.Children[i+1].Value
			cg.symbols.Define(name, typ, KindArg)
			i += 2
		} else {
			break
		}
	}
}

// ─────────────────────────────────────────────
// 5. Subroutine Body
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileSubroutineBody(node *parser.Node) {
	for _, child := range node.Children {
		switch child.Type {
		case "varDec":
			cg.compileVarDec(child)
		case "statements":
			cg.compileStatements(child)
		}
	}
}

// ─────────────────────────────────────────────
// 6. Variable Declaration
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileVarDec(node *parser.Node) {
	// children: 'var', type, varName (, varName)* ';'
	typ := node.Children[1].Value
	for i := 2; i < len(node.Children); i++ {
		if node.Children[i].Type == "identifier" {
			cg.symbols.Define(node.Children[i].Value, typ, KindVar)
		}
	}
}

// ─────────────────────────────────────────────
// 7. Statements
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileStatements(node *parser.Node) {
	for _, child := range node.Children {
		switch child.Type {
		case "letStatement":
			cg.compileLet(child)
		case "ifStatement":
			cg.compileIf(child)
		case "whileStatement":
			cg.compileWhile(child)
		case "doStatement":
			cg.compileDo(child)
		case "returnStatement":
			cg.compileReturn(child)
		}
	}
}

// ─────────────────────────────────────────────
// 8. let
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileLet(node *parser.Node) {
	// children: 'let', varName, ('[' expression ']')?, '=', expression, ';'
	varName := node.Children[1].Value
	sym := cg.symbols.Lookup(varName)

	isArray := false
	exprIndex := 3 // default: index of the RHS expression

	// Detect optional array indexing
	if len(node.Children) > 3 && node.Children[2].Value == "[" {
		isArray = true
		// Push base address of array
		cg.writer.WritePush(kindToSegment(sym.Kind), sym.Index)
		// Compile index expression (between '[' and ']')
		cg.compileExpression(firstChildOfTypeFrom(node, "expression", 0))
		// Add: base + index → address of target element
		cg.writer.WriteArithmetic("add")
		// The RHS expression is after ']', which is child index 5
		exprIndex = 5
	}

	// Compile RHS expression
	cg.compileExpression(firstChildOfTypeFrom(node, "expression", exprIndex))

	if isArray {
		// Store result in temp 0, set THAT to the computed address, then pop
		cg.writer.WritePop("temp", 0)
		cg.writer.WritePop("pointer", 1)
		cg.writer.WritePush("temp", 0)
		cg.writer.WritePop("that", 0)
	} else {
		cg.writer.WritePop(kindToSegment(sym.Kind), sym.Index)
	}
}

// firstChildOfTypeFrom finds the Nth expression child starting at child index `from`
func firstChildOfTypeFrom(node *parser.Node, typ string, from int) *parser.Node {
	for i := from; i < len(node.Children); i++ {
		if node.Children[i].Type == typ {
			return node.Children[i]
		}
	}
	return nil
}

// ─────────────────────────────────────────────
// 9. if
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileIf(node *parser.Node) {
	// children: 'if', '(', expression, ')', '{', statements, '}' [, 'else', '{', statements, '}']
	labelTrue := cg.uniqueLabel("IF_TRUE")
	labelFalse := cg.uniqueLabel("IF_FALSE")
	labelEnd := cg.uniqueLabel("IF_END")

	// Compile condition
	expr := firstChildOfType(node, "expression")
	cg.compileExpression(expr)
	cg.writer.WriteIfGoto(labelTrue)
	cg.writer.WriteGoto(labelFalse)

	// True branch
	cg.writer.WriteLabel(labelTrue)
	statementsNodes := childrenOfType(node, "statements")
	cg.compileStatements(statementsNodes[0])

	hasElse := len(statementsNodes) > 1

	if hasElse {
		cg.writer.WriteGoto(labelEnd)
	}

	// False / else branch
	cg.writer.WriteLabel(labelFalse)
	if hasElse {
		cg.compileStatements(statementsNodes[1])
		cg.writer.WriteLabel(labelEnd)
	}
}

// ─────────────────────────────────────────────
// 10. while
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileWhile(node *parser.Node) {
	// children: 'while', '(', expression, ')', '{', statements, '}'
	labelStart := cg.uniqueLabel("WHILE_START")
	labelEnd := cg.uniqueLabel("WHILE_END")

	cg.writer.WriteLabel(labelStart)
	expr := firstChildOfType(node, "expression")
	cg.compileExpression(expr)
	cg.writer.WriteArithmetic("not")
	cg.writer.WriteIfGoto(labelEnd)

	stmts := firstChildOfType(node, "statements")
	cg.compileStatements(stmts)

	cg.writer.WriteGoto(labelStart)
	cg.writer.WriteLabel(labelEnd)
}

// ─────────────────────────────────────────────
// 11. do
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileDo(node *parser.Node) {
	// children: 'do', <subroutine call tokens...>, ';'
	// The call tokens start at index 1
	cg.compileSubroutineCall(node, 1)
	// Discard the return value (void method)
	cg.writer.WritePop("temp", 0)
}

// ─────────────────────────────────────────────
// 12. return
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileReturn(node *parser.Node) {
	// children: 'return', [expression], ';'
	expr := firstChildOfType(node, "expression")
	if expr != nil {
		cg.compileExpression(expr)
	} else {
		// void: push dummy 0
		cg.writer.WritePush("constant", 0)
	}
	cg.writer.WriteReturn()
}

// ─────────────────────────────────────────────
// 13. Subroutine Call
// ─────────────────────────────────────────────

// compileSubroutineCall handles: name(...) or name.name(...)
// startIdx is where the call tokens begin in node.Children
func (cg *CodeGenerator) compileSubroutineCall(node *parser.Node, startIdx int) {
	children := node.Children[startIdx:]

	// children layout:
	//   name '(' expressionList ')'              → simple call
	//   name '.' name '(' expressionList ')'     → qualified call

	name1 := children[0].Value
	nArgs := 0
	var callName string

	if children[1].Value == "." {
		// Qualified: could be ClassName.method() or varName.method()
		methodName := children[2].Value
		sym := cg.symbols.Lookup(name1)

		if sym != nil {
			// It's a variable — push the object, call its type's method
			cg.writer.WritePush(kindToSegment(sym.Kind), sym.Index)
			nArgs = 1 // 'this' is argument 0
			callName = fmt.Sprintf("%s.%s", sym.Type, methodName)
		} else {
			// It's a class name — static/function call
			callName = fmt.Sprintf("%s.%s", name1, methodName)
		}

		// expressionList is at index 4 (name '.' name '(' expressionList ')')
		exprList := children[4]
		nArgs += cg.compileExpressionList(exprList)
	} else {
		// Unqualified: implicit method call on current object
		// Push 'this' as argument 0
		cg.writer.WritePush("pointer", 0)
		nArgs = 1
		callName = fmt.Sprintf("%s.%s", cg.className, name1)

		// expressionList is at index 2 (name '(' expressionList ')')
		exprList := children[2]
		nArgs += cg.compileExpressionList(exprList)
	}

	cg.writer.WriteCall(callName, nArgs)
}

// ─────────────────────────────────────────────
// 14. Expression
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileExpression(node *parser.Node) {
	if node == nil {
		return
	}
	// children: term (op term)*
	i := 0
	for i < len(node.Children) {
		child := node.Children[i]
		if child.Type == "term" {
			cg.compileTerm(child)
			i++
		} else {
			// it's an operator symbol
			op := child.Value
			i++
			if i < len(node.Children) {
				cg.compileTerm(node.Children[i])
				i++
			}
			cg.writeOp(op)
		}
	}
}

func (cg *CodeGenerator) writeOp(op string) {
	switch op {
	case "+":
		cg.writer.WriteArithmetic("add")
	case "-":
		cg.writer.WriteArithmetic("sub")
	case "*":
		cg.writer.WriteCall("Math.multiply", 2)
	case "/":
		cg.writer.WriteCall("Math.divide", 2)
	case "&":
		cg.writer.WriteArithmetic("and")
	case "|":
		cg.writer.WriteArithmetic("or")
	case "<":
		cg.writer.WriteArithmetic("lt")
	case ">":
		cg.writer.WriteArithmetic("gt")
	case "=":
		cg.writer.WriteArithmetic("eq")
	}
}

// ─────────────────────────────────────────────
// 15. Term
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileTerm(node *parser.Node) {
	if node == nil || len(node.Children) == 0 {
		return
	}

	first := node.Children[0]

	switch first.Type {
	case "integerConstant":
		val, _ := strconv.Atoi(first.Value)
		cg.writer.WritePush("constant", val)

	case "stringConstant":
		str := first.Value
		// Allocate string object, then append each character
		cg.writer.WritePush("constant", len(str))
		cg.writer.WriteCall("String.new", 1)
		for _, ch := range str {
			cg.writer.WritePush("constant", int(ch))
			cg.writer.WriteCall("String.appendChar", 2)
		}

	case "keyword":
		switch first.Value {
		case "true":
			cg.writer.WritePush("constant", 0)
			cg.writer.WriteArithmetic("not") // ~0 = -1 = true
		case "false", "null":
			cg.writer.WritePush("constant", 0)
		case "this":
			cg.writer.WritePush("pointer", 0)
		}

	case "symbol":
		switch first.Value {
		case "(":
			// '(' expression ')'
			cg.compileExpression(node.Children[1])
		case "-":
			// unary minus
			cg.compileTerm(node.Children[1])
			cg.writer.WriteArithmetic("neg")
		case "~":
			// bitwise NOT
			cg.compileTerm(node.Children[1])
			cg.writer.WriteArithmetic("not")
		}

	case "identifier":
		// Could be: varName, varName[expr], subroutineCall
		if len(node.Children) == 1 {
			// Plain variable
			sym := cg.symbols.Lookup(first.Value)
			if sym != nil {
				cg.writer.WritePush(kindToSegment(sym.Kind), sym.Index)
			}
		} else if len(node.Children) >= 3 && node.Children[1].Value == "[" {
			// varName '[' expression ']'
			sym := cg.symbols.Lookup(first.Value)
			cg.writer.WritePush(kindToSegment(sym.Kind), sym.Index)
			cg.compileExpression(node.Children[2])
			cg.writer.WriteArithmetic("add")
			cg.writer.WritePop("pointer", 1)
			cg.writer.WritePush("that", 0)
		} else {
			// Subroutine call
			cg.compileSubroutineCall(node, 0)
		}
	}
}

// ─────────────────────────────────────────────
// 16. Expression List
// ─────────────────────────────────────────────

func (cg *CodeGenerator) compileExpressionList(node *parser.Node) int {
	if node == nil {
		return 0
	}
	count := 0
	for _, child := range node.Children {
		if child.Type == "expression" {
			cg.compileExpression(child)
			count++
		}
	}
	return count
}
