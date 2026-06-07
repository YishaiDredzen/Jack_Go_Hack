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

// ParseClass is the entry point for parsing an entire file
func (p *Parser) ParseClass() *Node {
	classNode := &Node{Type: "class"}

	// Consume 'class' keyword
	if p.current() != nil && p.current().Value == "class" {
		classNode.Children = append(classNode.Children, &Node{Type: "keyword", Value: "class"})
		p.advance()
	}

	// Consume className (identifier)
	if p.current() != nil && p.current().TokenType == "identifier" {
		classNode.Children = append(classNode.Children, &Node{Type: "identifier", Value: p.current().Value})
		p.advance()
	}

	// Consume '{' symbol
	if p.current() != nil && p.current().Value == "{" {
		classNode.Children = append(classNode.Children, &Node{Type: "symbol", Value: "{"})
		p.advance()
	}

	// Here you would write loops to compile class variables and subroutines recursively!

	// Consume '}' symbol
	if p.current() != nil && p.current().Value == "}" {
		classNode.Children = append(classNode.Children, &Node{Type: "symbol", Value: "}"})
		p.advance()
	}

	return classNode
}
