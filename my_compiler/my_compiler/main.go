package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"my_compiler/parser"
	"my_compiler/tokenizer"
)

func main() {
	// Replace this path string with your actual Jack source file inside VS Code
	filePath := "Main.jack"

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}

	// 1. Run Lexical analysis (Lab 4 Part 1)
	tokens := tokenizer.Tokenize(string(content))

	// 2. Run Syntax parsing (Lab 4 Part 2)
	p := parser.NewParser(tokens)
	astTree := p.ParseClass()

	// 3. Convert AST properties directly into clean XML tags
	xmlOutput := astTree.ToXML(0)

	// 4. Save file out for testing
	outputFile := "MainOutput.xml"
	err = ioutil.WriteFile(outputFile, []byte(xmlOutput), 0644)
	if err != nil {
		log.Fatalf("Failed to write XML output: %v", err)
	}

	fmt.Printf("Success! Syntax analysis XML saved to %s\n", outputFile)
}
