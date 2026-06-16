package parser

import (
	"fmt"
	"strings"
)

type Node struct {
	Type     string  // e.g., "whileStatement", "expression", "keyword", "symbol"
	Value    string  // Only filled for terminal elements (e.g., "let", "count")
	Children []*Node // Array of nested elements
}

// terminal token types — these are always single-line leaf nodes
var terminalTypes = map[string]bool{
	"keyword":         true,
	"symbol":          true,
	"integerConstant": true,
	"stringConstant":  true,
	"identifier":      true,
}

// ToXML recursively creates the structured XML format required for Lab 4
func (n *Node) ToXML(indent int) string {
	indentStr := strings.Repeat("  ", indent) // 2 spaces per indentation level

	// Terminal nodes are always single-line: <keyword> let </keyword>
	if terminalTypes[n.Type] {
		return fmt.Sprintf("%s<%s> %s </%s>\n", indentStr, n.Type, escapeXML(n.Value), n.Type)
	}

	// Non-terminal containers always use the two-line open/close format,
	// even when empty — nand2tetris TextComparer requires this for
	// parameterList, expressionList, statements, etc.
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s<%s>\n", indentStr, n.Type))
	for _, child := range n.Children {
		sb.WriteString(child.ToXML(indent + 1))
	}
	sb.WriteString(fmt.Sprintf("%s</%s>\n", indentStr, n.Type))

	return sb.String()
}

func escapeXML(val string) string {
	switch val {
	case "<":
		return "&lt;"
	case ">":
		return "&gt;"
	case "\"":
		return "&quot;"
	case "&":
		return "&amp;"
	default:
		return val
	}
}