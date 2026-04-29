package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <path-to-vm-file-or-directory>")
		return
	}

	inputPath := os.Args[1]
	var vmFiles []string
	var outputPath string

	// 1. Determine if input is a file or a directory
	info, err := os.Stat(inputPath)
	if err != nil {
		fmt.Printf("Error reading path: %v\n", err)
		return
	}

	if info.IsDir() {
		// Output file named after the directory (e.g., /Main/Main.asm)
		baseName := filepath.Base(inputPath)
		outputPath = filepath.Join(inputPath, baseName+".asm")

		// Get all .vm files in the directory
		entries, _ := os.ReadDir(inputPath)
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".vm") {
				vmFiles = append(vmFiles, filepath.Join(inputPath, entry.Name()))
			}
		}

		// 2. Initialize CodeWriter once for the entire output
		outFile, _ := os.Create(outputPath)
		defer outFile.Close()

		writer := NewCodeWriter(outFile)

		// 3. Only write bootstrap if Sys.vm is present
		hasSys := false
		for _, entry := range entries {
			if entry.Name() == "Sys.vm" {
				hasSys = true
				break
			}
		}
		if hasSys {
			writer.WriteInit()
		}

		// 4. Process each VM file
		for _, file := range vmFiles {
			writer.SetFileName(file)
			processFile(file, writer)
		}

	} else {
		// Single file mode — no bootstrap
		outputPath = strings.Replace(inputPath, ".vm", ".asm", 1)
		vmFiles = append(vmFiles, inputPath)

		outFile, _ := os.Create(outputPath)
		defer outFile.Close()

		writer := NewCodeWriter(outFile)

		for _, file := range vmFiles {
			writer.SetFileName(file)
			processFile(file, writer)
		}
	}
}

func processFile(filePath string, writer *CodeWriter) {
	inFile, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer inFile.Close()

	var lines []string
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	parser := NewParser(lines)
	for parser.HasMoreCommands() {
		parser.Advance()
		if parser.Current() == "" {
			continue
		}

		switch parser.CommandType() {
		case "C_ARITHMETIC":
			writer.WriteArithmetic(parser.Arg1())
		case "C_PUSH", "C_POP":
			writer.WritePushPop(parser.CommandType(), parser.Arg1(), parser.Arg2())
		case "C_LABEL":
			writer.WriteLabel(parser.Arg1())
		case "C_GOTO":
			writer.WriteGoto(parser.Arg1())
		case "C_IF":
			writer.WriteIf(parser.Arg1())
		case "C_FUNCTION":
			writer.WriteFunction(parser.Arg1(), atoi(parser.Arg2()))
		case "C_CALL":
			writer.WriteCall(parser.Arg1(), atoi(parser.Arg2()))
		case "C_RETURN":
			writer.WriteReturn()
		}
	}
}
