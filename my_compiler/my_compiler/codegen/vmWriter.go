package codegen

import (
	"fmt"
	"strings"
)

type VMWriter struct {
	output strings.Builder
}

func NewVMWriter() *VMWriter {
	return &VMWriter{}
}

func (w *VMWriter) WritePush(segment string, index int) {
	w.write(fmt.Sprintf("push %s %d", segment, index))
}

func (w *VMWriter) WritePop(segment string, index int) {
	w.write(fmt.Sprintf("pop %s %d", segment, index))
}

func (w *VMWriter) WriteArithmetic(command string) {
	w.write(command)
}

func (w *VMWriter) WriteLabel(label string) {
	w.write(fmt.Sprintf("label %s", label))
}

func (w *VMWriter) WriteGoto(label string) {
	w.write(fmt.Sprintf("goto %s", label))
}

func (w *VMWriter) WriteIfGoto(label string) {
	w.write(fmt.Sprintf("if-goto %s", label))
}

func (w *VMWriter) WriteCall(name string, nArgs int) {
	w.write(fmt.Sprintf("call %s %d", name, nArgs))
}

func (w *VMWriter) WriteFunction(name string, nLocals int) {
	w.write(fmt.Sprintf("function %s %d", name, nLocals))
}

func (w *VMWriter) WriteReturn() {
	w.write("return")
}

func (w *VMWriter) String() string {
	return w.output.String()
}

func (w *VMWriter) write(line string) {
	w.output.WriteString(line + "\n")
}
