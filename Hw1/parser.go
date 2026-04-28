package main

import "strings"

type Parser struct {
	lines []string
	index int
}

func NewParser(lines []string) *Parser {
	return &Parser{lines: lines, index: -1}
}

func (p *Parser) HasMoreCommands() bool {
	return p.index+1 < len(p.lines)
}

func (p *Parser) Advance() {
	p.index++
}

func (p *Parser) Current() string {
	line := strings.TrimSpace(p.lines[p.index])

	if i := strings.Index(line, "//"); i != -1 {
		line = line[:i]
	}

	return strings.TrimSpace(line)
}

func (p *Parser) CommandType() string {
	fields := strings.Fields(p.Current())
	if len(fields) == 0 {
		return ""
	}

	switch fields[0] {
	case "push":
		return "C_PUSH"
	case "pop":
		return "C_POP"
	case "label":
		return "C_LABEL"
	case "goto":
		return "C_GOTO"
	case "if-goto":
		return "C_IF"
	case "function":
		return "C_FUNCTION"
	case "call":
		return "C_CALL"
	case "return":
		return "C_RETURN"
	default:
		return "C_ARITHMETIC"
	}
}

func (p *Parser) Arg1() string {
	fields := strings.Fields(p.Current())
	cmdType := p.CommandType()

	// C_RETURN should not call Arg1 according to the spec,
	// but if it does, we should handle it.
	if cmdType == "C_ARITHMETIC" {
		return fields[0]
	}
	if len(fields) > 1 {
		return fields[1]
	}
	return ""
}

func (p *Parser) Arg2() string {
	fields := strings.Fields(p.Current())
	cmdType := p.CommandType()

	// Only these types have a second argument (usually an index or nVars/nArgs)
	if cmdType == "C_PUSH" || cmdType == "C_POP" || cmdType == "C_FUNCTION" || cmdType == "C_CALL" {
		if len(fields) > 2 {
			return fields[2]
		}
	}
	return ""
}
