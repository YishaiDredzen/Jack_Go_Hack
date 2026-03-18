package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Global variables
var (
	currentVMFile  string
	logicalCounter int
	writer         *bufio.Writer
)

// ─── Arithmetic Helper Functions ─────────────────────────────────────────────
func handleArithmetic(cmd string) {
	fmt.Fprintf(writer, "command: %s\n", cmd)
}
func handleAdd() { handleArithmetic("add") }
func handleSub() { handleArithmetic("sub") }
func handleNeg() { handleArithmetic("neg") }

// ─── Logic Helper Functions ───────────────────────────────────────────────────

func handleLogical(cmd string) {
	fmt.Fprintf(writer, "command: %s\n", cmd)
	fmt.Fprintf(writer, "counter: %d\n", logicalCounter)
	logicalCounter++
}

func handleEq() { handleLogical("eq") }
func handleGt() { handleLogical("gt") }
func handleLt() { handleLogical("lt") }

// ─── Memory Access Helper Functions ──────────────────────────────────────────

func handleMemory(cmd, segment string, index int) {
	fmt.Fprintf(writer, "command: %s segment: %s index: %d\n", cmd, segment, index)
}

func handlePush(segment string, index int) { handleMemory("push", segment, index) }
func handlePop(segment string, index int)  { handleMemory("pop", segment, index) }

// ─── Line Parser ──────────────────────────────────────────────────────────────

func parseLine(line string) {
	// Remove inline comments and trim whitespace
	if idx := strings.Index(line, "//"); idx != -1 {
		line = line[:idx]
	}
	line = strings.TrimSpace(line)

	if line == "" {
		return
	}

	parts := strings.Fields(line)
	command := parts[0]

	switch command {
	// Arithmetic
	case "add":
		handleAdd()
	case "sub":
		handleSub()
	case "neg":
		handleNeg()

	// Logic
	case "eq":
		handleEq()
	case "gt":
		handleGt()
	case "lt":
		handleLt()

	// Memory Access (3-word commands)
	case "push":
		index, _ := strconv.Atoi(parts[2])
		handlePush(parts[1], index)
	case "pop":
		index, _ := strconv.Atoi(parts[2])
		handlePop(parts[1], index)

	default:
		fmt.Println("Unknown command:", command)
	}
}

// ─── Process a Single .vm File ────────────────────────────────────────────────

func processVMFile(vmFilePath string) error {
	// Store current file name without .vm extension
	currentVMFile = strings.TrimSuffix(filepath.Base(vmFilePath), ".vm")

	// Reset logical counter for each new file
	logicalCounter = 1

	file, err := os.Open(vmFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read and parse line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parseLine(scanner.Text())
	}

	// g. Print to screen at the end of each input file
	fmt.Printf("End of input file: %s.vm\n", currentVMFile)

	return scanner.Err()
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	// a. Receive path from user
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter path to VM files: ")
	scanner.Scan()
	dirPath := strings.TrimSpace(scanner.Text())

	// b. Output file named after the folder
	folderName := filepath.Base(dirPath)
	outputFilePath := filepath.Join(dirPath, folderName+".asm")

	// c. Open single output file for writing
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	writer = bufio.NewWriter(outputFile)

	// d. Find all .vm files (no assumed count)
	vmFiles, err := filepath.Glob(filepath.Join(dirPath, "*.vm"))
	if err != nil || len(vmFiles) == 0 {
		fmt.Println("No .vm files found in:", dirPath)
		os.Exit(1)
	}

	// e. Process each .vm file
	for _, vmFile := range vmFiles {
		if err := processVMFile(vmFile); err != nil {
			fmt.Println("Error processing file:", err)
		}
	}

	writer.Flush()

	// h. Print to screen at the end of the program
	fmt.Printf("Output file is ready: %s.asm\n", folderName)
}
