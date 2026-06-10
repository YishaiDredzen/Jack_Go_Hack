package main

import (
	"fmt"
	"log"
	"my_compiler/parser"
	"my_compiler/tokenizer"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Pass a path as a command-line argument, or default to current directory
	inputPath := "."
	if len(os.Args) > 1 {
		inputPath = os.Args[1]
	}

	info, err := os.Stat(inputPath)
	if err != nil {
		log.Fatalf("Cannot access path: %v", err)
	}

	var jackFiles []string

	if info.IsDir() {
		// Collect all .jack files in the directory
		entries, err := os.ReadDir(inputPath)
		if err != nil {
			log.Fatalf("Cannot read directory: %v", err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".jack") {
				jackFiles = append(jackFiles, filepath.Join(inputPath, e.Name()))
			}
		}
		if len(jackFiles) == 0 {
			log.Fatalf("No .jack files found in: %s", inputPath)
		}
	} else {
		// Single file
		jackFiles = []string{inputPath}
	}

	for _, filePath := range jackFiles {
		processFile(filePath)
	}
}

func processFile(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file %s: %v", filePath, err)
	}

	// 1. Lexical analysis
	tokens := tokenizer.Tokenize(string(content))

	// 2. Syntax parsing
	p := parser.NewParser(tokens)
	astTree := p.ParseClass()

	// 3. Convert to XML
	xmlOutput := astTree.ToXML(0)

	// 4. Write output to an 'output' folder next to the .jack file
	//    so we never overwrite the expected .xml files from nand2tetris
	dir := filepath.Dir(filePath)
	outputDir := filepath.Join(dir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(filePath), ".jack") + ".xml"
	outputFile := filepath.Join(outputDir, baseName)

	err = os.WriteFile(outputFile, []byte(xmlOutput), 0644)
	if err != nil {
		log.Fatalf("Failed to write XML output: %v", err)
	}

	fmt.Printf("✓ %s\n  → %s\n", filePath, outputFile)
}