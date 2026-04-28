package main

import (
	"fmt"
	"os"
	"strings"
)

type CodeWriter struct {
	out             *os.File
	labelCount      int
	fileName        string
	currentFunction string
	callCount       int
}

func NewCodeWriter(out *os.File) *CodeWriter {
	return &CodeWriter{out: out}
}

func (cw *CodeWriter) SetFileName(name string) {
	// This extracts "Main" from "path/to/Main.vm"
	parts := strings.Split(name, "/")
	if len(parts) == 1 {
		parts = strings.Split(name, "\\")
	}
	filename := parts[len(parts)-1]
	cw.fileName = strings.Replace(filename, ".vm", "", 1)
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
	case "add#":
		cw.binary("M=M+D")
		cw.unary("M=-M")
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

// Stage 2 (8 online)
func (cw *CodeWriter) WriteInit() {
	fmt.Fprint(cw.out, "@256\nD=A\n@SP\nM=D\n")
	cw.WriteCall("Sys.init", 0)
}

func (cw *CodeWriter) WriteLabel(label string) {
	full := cw.currentFunction + "$" + label
	fmt.Fprintf(cw.out, "(%s)\n", full)
}

func (cw *CodeWriter) WriteGoto(label string) {
	full := cw.currentFunction + "$" + label
	fmt.Fprintf(cw.out, "@%s\n0;JMP\n", full)
}

func (cw *CodeWriter) WriteIf(label string) {
	full := cw.currentFunction + "$" + label

	fmt.Fprint(cw.out, "@SP\nAM=M-1\nD=M\n")
	fmt.Fprintf(cw.out, "@%s\nD;JNE\n", full)
}

func (cw *CodeWriter) WriteFunction(name string, nVars int) {
	cw.currentFunction = name

	fmt.Fprintf(cw.out, "(%s)\n", name)

	for i := 0; i < nVars; i++ {
		fmt.Fprint(cw.out, "@0\nD=A\n")
		cw.pushD()
	}
}

func (cw *CodeWriter) WriteCall(name string, nArgs int) {
	returnLabel := fmt.Sprintf("%s$ret.%d", name, cw.callCount)
	cw.callCount++

	// push return address
	fmt.Fprintf(cw.out, "@%s\nD=A\n", returnLabel)
	cw.pushD()

	// push LCL, ARG, THIS, THAT
	for _, seg := range []string{"LCL", "ARG", "THIS", "THAT"} {
		fmt.Fprintf(cw.out, "@%s\nD=M\n", seg)
		cw.pushD()
	}

	// ARG = SP - nArgs - 5
	fmt.Fprint(cw.out, "@SP\nD=M\n")
	fmt.Fprintf(cw.out, "@%d\nD=D-A\n", nArgs)
	fmt.Fprint(cw.out, "@5\nD=D-A\n@ARG\nM=D\n")

	// LCL = SP
	fmt.Fprint(cw.out, "@SP\nD=M\n@LCL\nM=D\n")

	// goto function
	fmt.Fprintf(cw.out, "@%s\n0;JMP\n", name)

	// return label
	fmt.Fprintf(cw.out, "(%s)\n", returnLabel)
}

func (cw *CodeWriter) WriteReturn() {

	// FRAME = LCL
	fmt.Fprint(cw.out, "@LCL\nD=M\n@R13\nM=D\n")

	// RET = *(FRAME-5)
	fmt.Fprint(cw.out, "@5\nA=D-A\nD=M\n@R14\nM=D\n")

	// *ARG = pop()
	fmt.Fprint(cw.out, "@SP\nAM=M-1\nD=M\n@ARG\nA=M\nM=D\n")

	// SP = ARG + 1
	fmt.Fprint(cw.out, "@ARG\nD=M+1\n@SP\nM=D\n")

	// restore THAT, THIS, ARG, LCL
	restore := func(offset int, seg string) {
		fmt.Fprintf(cw.out, "@R13\nD=M\n@%d\nA=D-A\nD=M\n@%s\nM=D\n", offset, seg)
	}

	restore(1, "THAT")
	restore(2, "THIS")
	restore(3, "ARG")
	restore(4, "LCL")

	// goto RET
	fmt.Fprint(cw.out, "@R14\nA=M\n0;JMP\n")
}
