package main

import (
	"fmt"
	"os"
	"strings"
)

type CodeWriter struct {
	out        *os.File
	labelCount int
	fileName   string
}

func NewCodeWriter(out *os.File, input string) *CodeWriter {
	parts := strings.Split(input, "\\")
	name := parts[len(parts)-1]
	name = strings.Replace(name, ".vm", "", 1)

	return &CodeWriter{out: out, fileName: name}
}

func (cw *CodeWriter) WriteArithmetic(cmd string) {
	fmt.Fprintf(cw.out, "// %s\n", cmd)

	switch cmd {

	case "add":
		cw.binary("M=M+D")
	case "sub":
		cw.binary("M=M-D")
	case "and":
		cw.binary("M=M&D")
	case "or":
		cw.binary("M=M|D")

	case "neg":
		cw.unary("M=-M")
	case "not":
		cw.unary("M=!M")

	case "eq", "gt", "lt":
		cw.compare(cmd)
	}
}

func (cw *CodeWriter) binary(op string) {
	fmt.Fprintln(cw.out, "@SP")
	fmt.Fprintln(cw.out, "AM=M-1")
	fmt.Fprintln(cw.out, "D=M")
	fmt.Fprintln(cw.out, "A=A-1")
	fmt.Fprintln(cw.out, op)
}

func (cw *CodeWriter) unary(op string) {
	fmt.Fprintln(cw.out, "@SP")
	fmt.Fprintln(cw.out, "A=M-1")
	fmt.Fprintln(cw.out, op)
}

func (cw *CodeWriter) compare(cmd string) {
	id := cw.labelCount
	cw.labelCount++

	trueLabel := fmt.Sprintf("TRUE_%d", id)
	endLabel := fmt.Sprintf("END_%d", id)

	fmt.Fprintln(cw.out, "@SP")
	fmt.Fprintln(cw.out, "AM=M-1")
	fmt.Fprintln(cw.out, "D=M")
	fmt.Fprintln(cw.out, "A=A-1")
	fmt.Fprintln(cw.out, "D=M-D")

	fmt.Fprintf(cw.out, "@%s\n", trueLabel)

	switch cmd {
	case "eq":
		fmt.Fprintln(cw.out, "D;JEQ")
	case "gt":
		fmt.Fprintln(cw.out, "D;JGT")
	case "lt":
		fmt.Fprintln(cw.out, "D;JLT")
	}

	fmt.Fprintln(cw.out, "@SP")
	fmt.Fprintln(cw.out, "A=M-1")
	fmt.Fprintln(cw.out, "M=0")
	fmt.Fprintf(cw.out, "@%s\n0;JMP\n", endLabel)

	fmt.Fprintf(cw.out, "(%s)\n", trueLabel)
	fmt.Fprintln(cw.out, "@SP")
	fmt.Fprintln(cw.out, "A=M-1")
	fmt.Fprintln(cw.out, "M=-1")

	fmt.Fprintf(cw.out, "(%s)\n", endLabel)
}

func (cw *CodeWriter) WritePushPop(cmd, segment, index string) {
	fmt.Fprintf(cw.out, "// %s %s %s\n", cmd, segment, index)

	if cmd == "C_PUSH" {

		switch segment {

		case "constant":
			fmt.Fprintf(cw.out, "@%s\nD=A\n", index)

		case "local":
			cw.pushFromSegment("LCL", index)

		case "argument":
			cw.pushFromSegment("ARG", index)

		case "this":
			cw.pushFromSegment("THIS", index)

		case "that":
			cw.pushFromSegment("THAT", index)

		case "temp":
			fmt.Fprintf(cw.out, "@%d\nD=M\n", 5+atoi(index))

		case "pointer":
			if index == "0" {
				fmt.Fprintln(cw.out, "@THIS\nD=M")
			} else {
				fmt.Fprintln(cw.out, "@THAT\nD=M")
			}

		case "static":
			fmt.Fprintf(cw.out, "@%s.%s\nD=M\n", cw.fileName, index)
		}

		cw.pushD()
	}

	if cmd == "C_POP" {

		switch segment {

		case "local":
			cw.popToSegment("LCL", index)

		case "argument":
			cw.popToSegment("ARG", index)

		case "this":
			cw.popToSegment("THIS", index)

		case "that":
			cw.popToSegment("THAT", index)

		case "temp":
			fmt.Fprintln(cw.out, "@SP")
			fmt.Fprintln(cw.out, "AM=M-1")
			fmt.Fprintln(cw.out, "D=M")
			fmt.Fprintf(cw.out, "@%d\nM=D\n", 5+atoi(index))

		case "pointer":
			fmt.Fprintln(cw.out, "@SP")
			fmt.Fprintln(cw.out, "AM=M-1")
			fmt.Fprintln(cw.out, "D=M")
			if index == "0" {
				fmt.Fprintln(cw.out, "@THIS\nM=D")
			} else {
				fmt.Fprintln(cw.out, "@THAT\nM=D")
			}

		case "static":
			fmt.Fprintln(cw.out, "@SP")
			fmt.Fprintln(cw.out, "AM=M-1")
			fmt.Fprintln(cw.out, "D=M")
			fmt.Fprintf(cw.out, "@%s.%s\nM=D\n", cw.fileName, index)
		}
	}
}

func (cw *CodeWriter) pushD() {
	fmt.Fprint(cw.out, "@SP\nA=M\nM=D\n@SP\nM=M+1\n")
}

func (cw *CodeWriter) pushFromSegment(base, index string) {
	fmt.Fprintf(cw.out, "@%s\nD=A\n@%s\nA=M+D\nD=M\n", index, base)
}

func (cw *CodeWriter) popToSegment(base, index string) {
	fmt.Fprintf(cw.out, "@%s\nD=A\n@%s\nD=M+D\n@R13\nM=D\n", index, base)
	fmt.Fprint(cw.out, "@SP\nAM=M-1\nD=M\n@R13\nA=M\nM=D\n")
}

func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
