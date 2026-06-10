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

// 1. ParseClass: entry point for a .jack file
func (p *Parser) ParseClass() *Node {
	classNode := &Node{Type: "class"}

	// Consume 'class'
	if p.current() != nil {
		classNode.Children = append(classNode.Children, &Node{Type: "keyword", Value: p.current().Value})
		p.advance()
	}

	// Consume className
	if p.current() != nil {
		classNode.Children = append(classNode.Children, &Node{Type: "identifier", Value: p.current().Value})
		p.advance()
	}

	// Consume '{'
	if p.current() != nil {
		classNode.Children = append(classNode.Children, &Node{Type: "symbol", Value: p.current().Value})
		p.advance()
	}

	// Parse class variable declarations (static / field) loop
	for p.current() != nil && (p.current().Value == "static" || p.current().Value == "field") {
		classNode.Children = append(classNode.Children, p.CompileClassVarDec())
	}

	// Parse subroutine declarations (constructor / function / method) loop
	for p.current() != nil && (p.current().Value == "constructor" || p.current().Value == "function" || p.current().Value == "method") {
		classNode.Children = append(classNode.Children, p.CompileSubroutine())
	}

	// Consume '}'
	if p.current() != nil && p.current().Value == "}" {
		classNode.Children = append(classNode.Children, &Node{Type: "symbol", Value: p.current().Value})
		p.advance()
	}

	return classNode
}

// 2. CompileClassVarDec: Parses lines like "field int score;"
func (p *Parser) CompileClassVarDec() *Node {
	node := &Node{Type: "classVarDec"}

	// 'static' or 'field'
	node.Children = append(node.Children, &Node{Type: "keyword", Value: p.current().Value})
	p.advance()

	// Type (int, char, boolean, or class name)
	node.Children = append(node.Children, &Node{Type: p.current().TokenType, Value: p.current().Value})
	p.advance()

	// Variable Name
	node.Children = append(node.Children, &Node{Type: "identifier", Value: p.current().Value})
	p.advance()

	// Handle extra variables on the same line separated by commas (e.g., field int x, y, z;)
	for p.current() != nil && p.current().Value == "," {
		node.Children = append(node.Children, &Node{Type: "symbol", Value: ","})
		p.advance()
		node.Children = append(node.Children, &Node{Type: "identifier", Value: p.current().Value})
		p.advance()
	}

	// ';'
	node.Children = append(node.Children, &Node{Type: "symbol", Value: ";"})
	p.advance()

	return node
}

// 3. CompileSubroutine: Parses methods and functions safely
func (p *Parser) CompileSubroutine() *Node {
	node := &Node{Type: "subroutineDec"}

	// 'constructor', 'function', or 'method'
	node.Children = append(node.Children, &Node{Type: "keyword", Value: p.current().Value})
	p.advance()

	// Return type ('void' or data type)
	node.Children = append(node.Children, &Node{Type: p.current().TokenType, Value: p.current().Value})
	p.advance()

	// Subroutine name
	node.Children = append(node.Children, &Node{Type: "identifier", Value: p.current().Value})
	p.advance()

	// '('
	node.Children = append(node.Children, &Node{Type: "symbol", Value: "("})
	p.advance()

	// Parameter List
	node.Children = append(node.Children, p.CompileParameterList())

	// ')'
	node.Children = append(node.Children, &Node{Type: "symbol", Value: ")"})
	p.advance()

	// --- Parse the Subroutine Body safely ---
	subBody := &Node{Type: "subroutineBody"}
	
	// Consume '{'
	if p.current() != nil {
		subBody.Children = append(subBody.Children, &Node{Type: "symbol", Value: p.current().Value})
		p.advance()
	}

	// Skip internal tokens inside function body until we find the closing '}'
	for p.current() != nil && p.current().Value != "}" {
		p.advance()
	}

	// Consume '}'
	if p.current() != nil && p.current().Value == "}" {
		subBody.Children = append(subBody.Children, &Node{Type: "symbol", Value: "}"})
		p.advance()
	}

	node.Children = append(node.Children, subBody)
	return node
}

// 4. CompileParameterList: Parses argument lists
func (p *Parser) CompileParameterList() *Node {
	node := &Node{Type: "parameterList"}

	// If empty parameter list, return node immediately
	if p.current() != nil && p.current().Value == ")" {
		return node
	}

	// First type & arg
	node.Children = append(node.Children, &Node{Type: p.current().TokenType, Value: p.current().Value})
	p.advance()
	node.Children = append(node.Children, &Node{Type: "identifier", Value: p.current().Value})
	p.advance()

	// Loop for remaining commas and arguments
	for p.current() != nil && p.current().Value == "," {
		node.Children = append(node.Children, &Node{Type: "symbol", Value: ","})
		p.advance()
		node.Children = append(node.Children, &Node{Type: p.current().TokenType, Value: p.current().Value})
		p.advance()
		node.Children = append(node.Children, &Node{Type: "identifier", Value: p.current().Value})
		p.advance()
	}

	return node
}
