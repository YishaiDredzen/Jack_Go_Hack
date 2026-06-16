package tokenizer

import (
	"regexp"
	"strings"
)

var keywords = []string{
	"class", "constructor", "function", "method", "field", "static", "var",
	"int", "char", "boolean", "void", "true", "false", "null", "this",
	"let", "do", "if", "else", "while", "return",
}

var symbols = []string{
	"{", "}", "(", ")", "[", "]", ".", ",", ";", "+", "-", "*", "/", "&", "|", "<", ">", "=", "~",
}

type Token struct {
	TokenType string // "keyword", "symbol", "integerConstant", "stringConstant", "identifier"
	Value     string
}

func Tokenize(source string) []*Token {
	source = removeComments(source)

	// Combine regexes into one extraction pattern
	// Crucial: we match strings, symbols, and integers first so they don't break on keywords
	pattern := `"[^"\n]+"` + `|` + 
		`[{}()\[\].,;+\-*/&|<>=~]` + `|` + 
		`\d+` + `|` + 
		`[a-zA-Z_]\w*`

	re := regexp.MustCompile(pattern)
	rawTokens := re.FindAllString(source, -1)

	var tokens []*Token
	for _, raw := range rawTokens {
		tokens = append(tokens, &Token{
			TokenType: detectType(raw),
			Value:     strings.Trim(raw, `"`), // Remove literal quotes around stringConstants
		})
	}
	return tokens
}

func detectType(val string) string {
	if regexp.MustCompile(`^"[^"\n]+"$`).MatchString(val) {
		return "stringConstant"
	}
	if regexp.MustCompile(`^\d+$`).MatchString(val) {
		return "integerConstant"
	}
	for _, s := range symbols {
		if s == val {
			return "symbol"
		}
	}
	for _, k := range keywords {
		if k == val {
			return "keyword"
		}
	}
	return "identifier"
}

func removeComments(str string) string {
	// Remove single-line comments //
	str = regexp.MustCompile(`//.*`).ReplaceAllString(str, "")
	// Remove multi-line comments /* ... */
	str = regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(str, "")
	return str
}