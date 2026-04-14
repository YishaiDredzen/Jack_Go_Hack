package main

import (
	"bufio"
	"os"
	"strings"
)

func main() {
	//if len(os.Args) < 2 {
	//	println("Usage: go run . test.vm")
	//	return
	//}

	input := os.Args[1]
	output := strings.Replace(input, ".vm", ".asm", 1)

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
	writer := NewCodeWriter(outFile, input)

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
