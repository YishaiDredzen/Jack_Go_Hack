package parser

import (
	"my_compiler/tokenizer"
)

type Parser struct {
	tokens []*tokenizer.Token
	pos    int
}

func NewParser(tokens []*tokenizer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) current() *tokenizer.Token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	p.pos++
}

// peek returns the token at pos+offset without advancing
func (p *Parser) peek(offset int) *tokenizer.Token {
	i := p.pos + offset
	if i >= len(p.tokens) {
		return nil
	}
	return p.tokens[i]
}

// consumeToken appends the current token as a child node and advances
func (p *Parser) consumeToken(parent *Node) {
	if p.current() == nil {
		return
	}
	parent.Children = append(parent.Children, &Node{
		Type:  p.current().TokenType,
		Value: p.current().Value,
	})
	p.advance()
}

// ─────────────────────────────────────────────
// 1. ParseClass
// ─────────────────────────────────────────────

func (p *Parser) ParseClass() *Node {
	classNode := &Node{Type: "class"}

	p.consumeToken(classNode) // 'class'
	p.consumeToken(classNode) // className
	p.consumeToken(classNode) // '{'

	for p.current() != nil && (p.current().Value == "static" || p.current().Value == "field") {
		classNode.Children = append(classNode.Children, p.CompileClassVarDec())
	}

	for p.current() != nil && (p.current().Value == "constructor" || p.current().Value == "function" || p.current().Value == "method") {
		classNode.Children = append(classNode.Children, p.CompileSubroutine())
	}

	p.consumeToken(classNode) // '}'

	return classNode
}

// ─────────────────────────────────────────────
// 2. CompileClassVarDec
// ─────────────────────────────────────────────

func (p *Parser) CompileClassVarDec() *Node {
	node := &Node{Type: "classVarDec"}

	p.consumeToken(node) // 'static' or 'field'
	p.consumeToken(node) // type
	p.consumeToken(node) // varName

	for p.current() != nil && p.current().Value == "," {
		p.consumeToken(node) // ','
		p.consumeToken(node) // varName
	}

	p.consumeToken(node) // ';'
	return node
}

// ─────────────────────────────────────────────
// 3. CompileSubroutine
// ─────────────────────────────────────────────

func (p *Parser) CompileSubroutine() *Node {
	node := &Node{Type: "subroutineDec"}

	p.consumeToken(node) // 'constructor' | 'function' | 'method'
	p.consumeToken(node) // return type
	p.consumeToken(node) // subroutine name
	p.consumeToken(node) // '('
	node.Children = append(node.Children, p.CompileParameterList())
	p.consumeToken(node) // ')'
	node.Children = append(node.Children, p.CompileSubroutineBody())

	return node
}

// ─────────────────────────────────────────────
// 4. CompileParameterList
// ─────────────────────────────────────────────

func (p *Parser) CompileParameterList() *Node {
	node := &Node{Type: "parameterList"}

	if p.current() == nil || p.current().Value == ")" {
		return node
	}

	p.consumeToken(node) // type
	p.consumeToken(node) // varName

	for p.current() != nil && p.current().Value == "," {
		p.consumeToken(node) // ','
		p.consumeToken(node) // type
		p.consumeToken(node) // varName
	}

	return node
}

// ─────────────────────────────────────────────
// 5. CompileSubroutineBody
// ─────────────────────────────────────────────

func (p *Parser) CompileSubroutineBody() *Node {
	node := &Node{Type: "subroutineBody"}

	p.consumeToken(node) // '{'

	// Zero or more varDec
	for p.current() != nil && p.current().Value == "var" {
		node.Children = append(node.Children, p.CompileVarDec())
	}

	// statements
	node.Children = append(node.Children, p.CompileStatements())

	p.consumeToken(node) // '}'
	return node
}

// ─────────────────────────────────────────────
// 6. CompileVarDec
// ─────────────────────────────────────────────

func (p *Parser) CompileVarDec() *Node {
	node := &Node{Type: "varDec"}

	p.consumeToken(node) // 'var'
	p.consumeToken(node) // type
	p.consumeToken(node) // varName

	for p.current() != nil && p.current().Value == "," {
		p.consumeToken(node) // ','
		p.consumeToken(node) // varName
	}

	p.consumeToken(node) // ';'
	return node
}

// ─────────────────────────────────────────────
// 7. CompileStatements
// ─────────────────────────────────────────────

func (p *Parser) CompileStatements() *Node {
	node := &Node{Type: "statements"}

	for p.current() != nil {
		switch p.current().Value {
		case "let":
			node.Children = append(node.Children, p.CompileLetStatement())
		case "if":
			node.Children = append(node.Children, p.CompileIfStatement())
		case "while":
			node.Children = append(node.Children, p.CompileWhileStatement())
		case "do":
			node.Children = append(node.Children, p.CompileDoStatement())
		case "return":
			node.Children = append(node.Children, p.CompileReturnStatement())
		default:
			// No more statements
			return node
		}
	}

	return node
}

// ─────────────────────────────────────────────
// 8. CompileLetStatement
// ─────────────────────────────────────────────

func (p *Parser) CompileLetStatement() *Node {
	node := &Node{Type: "letStatement"}

	p.consumeToken(node) // 'let'
	p.consumeToken(node) // varName

	// Optional array index: '[' expression ']'
	if p.current() != nil && p.current().Value == "[" {
		p.consumeToken(node) // '['
		node.Children = append(node.Children, p.CompileExpression())
		p.consumeToken(node) // ']'
	}

	p.consumeToken(node) // '='
	node.Children = append(node.Children, p.CompileExpression())
	p.consumeToken(node) // ';'

	return node
}

// ─────────────────────────────────────────────
// 9. CompileIfStatement
// ─────────────────────────────────────────────

func (p *Parser) CompileIfStatement() *Node {
	node := &Node{Type: "ifStatement"}

	p.consumeToken(node) // 'if'
	p.consumeToken(node) // '('
	node.Children = append(node.Children, p.CompileExpression())
	p.consumeToken(node) // ')'
	p.consumeToken(node) // '{'
	node.Children = append(node.Children, p.CompileStatements())
	p.consumeToken(node) // '}'

	// Optional 'else'
	if p.current() != nil && p.current().Value == "else" {
		p.consumeToken(node) // 'else'
		p.consumeToken(node) // '{'
		node.Children = append(node.Children, p.CompileStatements())
		p.consumeToken(node) // '}'
	}

	return node
}

// ─────────────────────────────────────────────
// 10. CompileWhileStatement
// ─────────────────────────────────────────────

func (p *Parser) CompileWhileStatement() *Node {
	node := &Node{Type: "whileStatement"}

	p.consumeToken(node) // 'while'
	p.consumeToken(node) // '('
	node.Children = append(node.Children, p.CompileExpression())
	p.consumeToken(node) // ')'
	p.consumeToken(node) // '{'
	node.Children = append(node.Children, p.CompileStatements())
	p.consumeToken(node) // '}'

	return node
}

// ─────────────────────────────────────────────
// 11. CompileDoStatement
// ─────────────────────────────────────────────

func (p *Parser) CompileDoStatement() *Node {
	node := &Node{Type: "doStatement"}

	p.consumeToken(node) // 'do'
	p.compileSubroutineCall(node)
	p.consumeToken(node) // ';'

	return node
}

// compileSubroutineCall handles: name(...) or name.name(...)
// It is called with the identifier already NOT consumed yet.
func (p *Parser) compileSubroutineCall(parent *Node) {
	p.consumeToken(parent) // subroutineName or className/varName

	if p.current() != nil && p.current().Value == "." {
		p.consumeToken(parent) // '.'
		p.consumeToken(parent) // subroutineName
	}

	p.consumeToken(parent) // '('
	parent.Children = append(parent.Children, p.CompileExpressionList())
	p.consumeToken(parent) // ')'
}

// ─────────────────────────────────────────────
// 12. CompileReturnStatement
// ─────────────────────────────────────────────

func (p *Parser) CompileReturnStatement() *Node {
	node := &Node{Type: "returnStatement"}

	p.consumeToken(node) // 'return'

	// Optional return expression
	if p.current() != nil && p.current().Value != ";" {
		node.Children = append(node.Children, p.CompileExpression())
	}

	p.consumeToken(node) // ';'
	return node
}

// ─────────────────────────────────────────────
// 13. CompileExpression
// ─────────────────────────────────────────────

var opSet = map[string]bool{
	"+": true, "-": true, "*": true, "/": true,
	"&": true, "|": true, "<": true, ">": true, "=": true,
}

func (p *Parser) CompileExpression() *Node {
	node := &Node{Type: "expression"}

	node.Children = append(node.Children, p.CompileTerm())

	// (op term)*
	for p.current() != nil && opSet[p.current().Value] {
		p.consumeToken(node) // op
		node.Children = append(node.Children, p.CompileTerm())
	}

	return node
}

// ─────────────────────────────────────────────
// 14. CompileTerm
// ─────────────────────────────────────────────

func (p *Parser) CompileTerm() *Node {
	node := &Node{Type: "term"}

	if p.current() == nil {
		return node
	}

	tok := p.current()

	switch {
	// integerConstant
	case tok.TokenType == "integerConstant":
		p.consumeToken(node)

	// stringConstant
	case tok.TokenType == "stringConstant":
		p.consumeToken(node)

	// keywordConstant: true, false, null, this
	case tok.Value == "true" || tok.Value == "false" || tok.Value == "null" || tok.Value == "this":
		p.consumeToken(node)

	// '(' expression ')'
	case tok.Value == "(":
		p.consumeToken(node) // '('
		node.Children = append(node.Children, p.CompileExpression())
		p.consumeToken(node) // ')'

	// unaryOp term: '-' or '~'
	case tok.Value == "-" || tok.Value == "~":
		p.consumeToken(node) // unaryOp
		node.Children = append(node.Children, p.CompileTerm())

	// identifier: could be varName, varName[expr], subroutineCall
	case tok.TokenType == "identifier":
		next := p.peek(1)
		if next != nil && next.Value == "[" {
			// varName '[' expression ']'
			p.consumeToken(node) // varName
			p.consumeToken(node) // '['
			node.Children = append(node.Children, p.CompileExpression())
			p.consumeToken(node) // ']'
		} else if next != nil && (next.Value == "(" || next.Value == ".") {
			// subroutineCall
			p.compileSubroutineCall(node)
		} else {
			// plain varName
			p.consumeToken(node)
		}
	}

	return node
}

// ─────────────────────────────────────────────
// 15. CompileExpressionList
// ─────────────────────────────────────────────

func (p *Parser) CompileExpressionList() *Node {
	node := &Node{Type: "expressionList"}

	if p.current() == nil || p.current().Value == ")" {
		return node
	}

	node.Children = append(node.Children, p.CompileExpression())

	for p.current() != nil && p.current().Value == "," {
		p.consumeToken(node) // ','
		node.Children = append(node.Children, p.CompileExpression())
	}

	return node
}