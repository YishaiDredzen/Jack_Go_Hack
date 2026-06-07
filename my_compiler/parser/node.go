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

// ToXML recursively creates the structured XML format required for Lab 4
func (n *Node) ToXML(indent int) string {
	indentStr := strings.Repeat("  ", indent) // 2 spaces per indentation level

	// If it has no children, it's a leaf/terminal node
	if len(n.Children) == 0 {
		return fmt.Sprintf("%s<%s> %s </%s>\n", indentStr, n.Type, escapeXML(n.Value), n.Type)
	}

	// If it has children, it's a non-terminal container
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
