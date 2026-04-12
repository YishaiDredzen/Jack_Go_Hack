package main

import (
	"bufio"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		println("Usage: go run . file.vm")
		return
	}

	input := os.Args[1]
	output := input[:len(input)-3] + "asm"

	inFile, _ := os.Open(input)
	defer inFile.Close()

	outFile, _ := os.Create(output)
	defer outFile.Close()

	var lines []string
	scanner := bufio.NewScanner(inFile)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	parser := NewParser(lines)
	writer := NewCodeWriter(outFile)

	for parser.HasMoreCommands() {
		parser.Advance()
		cmd := parser.Current()

		if cmd == "" {
			continue
		}

		switch parser.CommandType() {

		case "C_ARITHMETIC":
			writer.WriteArithmetic(parser.Arg1())

		case "C_PUSH", "C_POP":
			writer.WritePushPop(
				parser.CommandType(),
				parser.Arg1(),
				parser.Arg2(),
			)
		}
	}
}
