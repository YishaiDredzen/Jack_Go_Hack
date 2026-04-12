package main

import (
	"fmt"
	"os"
)

type CodeWriter struct {
	out *os.File
}

func NewCodeWriter(out *os.File) *CodeWriter {
	return &CodeWriter{out: out}
}

func (cw *CodeWriter) WriteArithmetic(cmd string) {
	fmt.Fprintf(cw.out, "// %s\n", cmd)

	switch cmd {

	case "add":
		fmt.Fprintln(cw.out, "@SP")
		fmt.Fprintln(cw.out, "AM=M-1")
		fmt.Fprintln(cw.out, "D=M")
		fmt.Fprintln(cw.out, "A=A-1")
		fmt.Fprintln(cw.out, "M=M+D")
	}
}

func (cw *CodeWriter) WritePushPop(cmd, segment, index string) {
	fmt.Fprintf(cw.out, "// %s %s %s\n", cmd, segment, index)

	if cmd == "C_PUSH" && segment == "constant" {
		fmt.Fprintf(cw.out, "@%s\nD=A\n", index)
		fmt.Fprintln(cw.out, "@SP")
		fmt.Fprintln(cw.out, "A=M")
		fmt.Fprintln(cw.out, "M=D")
		fmt.Fprintln(cw.out, "@SP")
		fmt.Fprintln(cw.out, "M=M+1")
	}
}
