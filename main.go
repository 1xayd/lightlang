package main

import (
	"fmt"
	"os"
	"strings"
)

func buildCommand(source string, output string) {
	content, err := os.ReadFile(source)
	if err != nil {
		fmt.Printf("Error reading source file: %v\n", err)
		return
	}

	nodes, err := Parse(string(content))
	if err != nil {
		fmt.Printf("Parse Error: %v\n", err)
		return
	}

	builder := NewBuilder()
	for _, node := range nodes {
		if err := node.TypeCheck(builder.SymbolTable); err != nil {
			fmt.Printf("Type Error: %v\n", err)
			return
		}
		node.Emit(builder)
	}
	builder.Emit(OpHalt, nil)

	instructions, constants := builder.Bytecode()
	instructions, constants = OptimizeBytecode(instructions, constants, builder.SymbolTable)

	err = SaveBytecode(output, instructions, constants)
	if err != nil {
		fmt.Printf("Error writing bytecode file: %v\n", err)
		return
	}

	fmt.Printf("Successfully built '%s' -> '%s'\n", source, output)
}

func runFile(target string) {
	vm := NewVM()

	if strings.HasSuffix(target, ".ll") {
		content, err := os.ReadFile(target)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}

		nodes, err := Parse(string(content))
		if err != nil {
			fmt.Printf("Parse Error: %v\n", err)
			return
		}

		builder := NewBuilder()
		for _, node := range nodes {
			if err := node.TypeCheck(builder.SymbolTable); err != nil {
				fmt.Printf("Type Error: %v\n", err)
				return
			}
			node.Emit(builder)
		}
		builder.Emit(OpHalt, nil)

		builder.Instructions, builder.Constants = OptimizeBytecode(builder.Instructions, builder.Constants, builder.SymbolTable)

		vm.Instructions, vm.Constants = builder.Instructions, builder.Constants

	} else {
		err := vm.loadBytecode(target)
		if err != nil {
			fmt.Printf("Error loading bytecode: %v\n", err)
			return
		}
	}

	if err := vm.Run(""); err != nil {
		fmt.Printf("Runtime Error: %v\n", err)
	}
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	if len(os.Args) == 2 {
		arg := os.Args[1]
		if arg == "help" || arg == "--help" || arg == "-h" {
			printHelp()
			return
		}

		runFile(arg)
		return
	}

	command := os.Args[1]

	switch command {
	case "build":
		if len(os.Args) < 3 {
			fmt.Println("Nope, do it like this: lightlang build <source.ll>")
			return
		}
		source := os.Args[2]
		output := strings.TrimSuffix(source, ".ll") + ".llbytecode"
		if len(os.Args) >= 4 {
			output = os.Args[3]
		}
		buildCommand(source, output)

	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Nope, do it like this: lightlang run <file.ll|file.llbytecode>")
			return
		}
		target := os.Args[2]
		runFile(target)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
	}
}

func printHelp() {
	fmt.Println("lightlang is a lightweight language implemented in go; portable and simple;")
	fmt.Println("lightlang build <file.ll>	Build bytecode from source")
	fmt.Println("lightlang run <file.ll> or <file.llbytecode>	Run source file directly or bytecode")
	fmt.Println("lightlang <file.ll|file.llbytecode>	Run file directly")
}
