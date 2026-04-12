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
	default:
		return "C_ARITHMETIC"
	}
}

func (p *Parser) Arg1() string {
	fields := strings.Fields(p.Current())

	if p.CommandType() == "C_ARITHMETIC" {
		return fields[0]
	}
	return fields[1]
}

func (p *Parser) Arg2() string {
	fields := strings.Fields(p.Current())
	return fields[2]
}
