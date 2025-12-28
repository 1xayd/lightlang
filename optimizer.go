package main

import (
	"lightlang/builtins"
	"math"
	"strconv"
)

type Optimizer struct {
	Instructions []Instruction
	Constants    []Constant
	SymbolTable  *SymbolTable
}

type globalInfo struct {
	name  string
	usage int
}

func NewOptimizer(instructions []Instruction, constants []Constant, sym *SymbolTable) *Optimizer {
	return &Optimizer{
		Instructions: instructions,
		Constants:    constants,
		SymbolTable:  sym,
	}
}

func (o *Optimizer) Optimize() ([]Instruction, []Constant) {
	for {
		originalLen := len(o.Instructions)

		o.doConstantFolding()

		o.doNameScraping()

		o.doCleanup()

		if len(o.Instructions) == originalLen {
			break
		}
	}

	o.doGarbageCollection()

	return o.Instructions, o.Constants
}

func (o *Optimizer) doNameScraping() {
	builtinlist := make(map[string]bool)
	for name := range builtins.Builtins {
		builtinlist[name] = true
	}
	globalUsage := make(map[string]int)
	localUsage := make(map[int]int)
	constantUsage := make(map[int]int)

	localNameMap := make(map[int]string)

	for _, inst := range o.Instructions {
		switch inst.Op {
		case OpGetGlobal, OpSetGlobal:
			if name, ok := inst.Arg.(string); ok {
				globalUsage[name]++
			}
		case OpGetLocal, OpSetLocal:
			if idx, ok := inst.Arg.(float64); ok {
				localIdx := int(idx)
				localUsage[localIdx]++
			}
		case OpConstant:
			if idx, ok := inst.Arg.(float64); ok {
				constIdx := int(idx)
				if constIdx >= 0 && constIdx < len(o.Constants) {
					constantUsage[constIdx]++
				}
			}
		case OpCall:
			if target, ok := inst.Arg.(string); ok && target != "" {
				globalUsage[target]++
			}
		}
	}

	localCounter := 1
	globalNameMap := make(map[string]string)
	globalCounter := 1

	var globalList []globalInfo
	for name, usage := range globalUsage {
		if builtinlist[name] {
			continue
		}
		globalList = append(globalList, globalInfo{name: name, usage: usage})
	}

	for i := 0; i < len(globalList); i++ {
		for j := i + 1; j < len(globalList); j++ {
			if globalList[j].usage > globalList[i].usage {
				globalList[i], globalList[j] = globalList[j], globalList[i]
			}
		}
	}

	for _, info := range globalList {
		newName := "g" + strconv.Itoa(globalCounter)
		globalNameMap[info.name] = newName
		globalCounter++
	}

	for localIdx := range localUsage {
		newName := "l" + strconv.Itoa(localCounter)
		localNameMap[localIdx] = newName
		localCounter++
	}

	for i := range o.Instructions {
		switch o.Instructions[i].Op {
		case OpGetGlobal, OpSetGlobal:
			if name, ok := o.Instructions[i].Arg.(string); ok {
				if newName, exists := globalNameMap[name]; exists {
					o.Instructions[i].Arg = newName
				}
			}
		case OpCall:
			if target, ok := o.Instructions[i].Arg.(string); ok {
				if newName, exists := globalNameMap[target]; exists {
					o.Instructions[i].Arg = newName
				}
			}
		}
	}

	if o.SymbolTable != nil {
		for oldName, newName := range globalNameMap {
			if typ, exists := o.SymbolTable.Globals[oldName]; exists {
				o.SymbolTable.Globals[newName] = typ
				delete(o.SymbolTable.Globals, oldName)
			}
		}

		if len(localNameMap) > 0 {
			oldNameToIdx := make(map[string]int)
			for name, idx := range o.SymbolTable.Locals {
				oldNameToIdx[name] = idx
			}

			newLocals := make(map[string]int)
			for oldName, idx := range oldNameToIdx {
				if newName, exists := localNameMap[idx]; exists {
					newLocals[newName] = idx
				} else {
					newLocals[oldName] = idx
				}
			}
			o.SymbolTable.Locals = newLocals
		}
	}
}

func (o *Optimizer) doConstantFolding() {
	constantValues := make(map[string]interface{})
	isConstant := make(map[string]bool)

	constantUsed := make([]bool, len(o.Constants))

	for i := 0; i < len(o.Instructions); i++ {
		inst := o.Instructions[i]

		if inst.Op == OpConstant {
			if idx, ok := inst.Arg.(float64); ok {
				constIdx := int(idx)
				if constIdx >= 0 && constIdx < len(o.Constants) {
					constantUsed[constIdx] = true
				}
			}
		}

		if inst.Op == OpSetGlobal {
			if name, ok := inst.Arg.(string); ok && i > 0 {
				prev := o.Instructions[i-1]
				if prev.Op == OpConstant {
					if idx, ok := prev.Arg.(float64); ok {
						constIdx := int(idx)
						if constIdx >= 0 && constIdx < len(o.Constants) {
							constantValues[name] = o.Constants[constIdx].Value
							isConstant[name] = true
						}
					}
				}
			}
		}
	}

	for i := 0; i < len(o.Instructions); i++ {
		if i+2 < len(o.Instructions) {
			if o.Instructions[i].Op == OpConstant &&
				o.Instructions[i+1].Op == OpConstant &&
				isArithmeticOp(o.Instructions[i+2].Op) {

				idx1, ok1 := o.Instructions[i].Arg.(float64)
				idx2, ok2 := o.Instructions[i+1].Arg.(float64)

				if ok1 && ok2 {
					constIdx1 := int(idx1)
					constIdx2 := int(idx2)

					if constIdx1 >= 0 && constIdx1 < len(o.Constants) &&
						constIdx2 >= 0 && constIdx2 < len(o.Constants) {

						val1 := o.Constants[constIdx1].Value
						val2 := o.Constants[constIdx2].Value

						result, ok := performArithmetic(val1, val2, o.Instructions[i+2].Op)
						if ok {
							constIdx := len(o.Constants)
							o.Constants = append(o.Constants, Constant{
								Value: result,
								Type:  getTypeString(result),
							})

							o.Instructions[i] = Instruction{
								Op:   OpConstant,
								Arg:  float64(constIdx),
								Line: o.Instructions[i].Line,
							}

							o.Instructions = append(o.Instructions[:i+1], o.Instructions[i+3:]...)
							i--
						}
					}
				}
			}
		}
	}
}

func (o *Optimizer) doCleanup() {
	globalUsage := make(map[string]int)
	localUsage := make(map[int]int)

	for _, inst := range o.Instructions {
		switch inst.Op {
		case OpGetGlobal:
			if name, ok := inst.Arg.(string); ok {
				globalUsage[name]++
			}
		case OpSetGlobal:
			if name, ok := inst.Arg.(string); ok {
				if _, exists := globalUsage[name]; !exists {
					globalUsage[name] = 0
				}
			}
		case OpGetLocal:
			if idx, ok := inst.Arg.(float64); ok {
				localIdx := int(idx)
				localUsage[localIdx]++
			}
		case OpSetLocal:
			if idx, ok := inst.Arg.(float64); ok {
				localIdx := int(idx)
				if _, exists := localUsage[localIdx]; !exists {
					localUsage[localIdx] = 0
				}
			}
		case OpMakeFunc:
			if inst.Arg != nil {
				if idx, ok := inst.Arg.(float64); ok {
					if int(idx) < len(o.Constants) {
						if functionConst, ok := o.Constants[int(idx)].Value.(float64); ok {
							ip := int(functionConst)
							if ip >= 0 && ip < len(o.Instructions) {
								for j := ip; j < len(o.Instructions); j++ {
									if o.Instructions[j].Op == OpReturn {
										break
									}
								}
							}
						}
					}
				}
			}
		case OpCall:
			if target, ok := inst.Arg.(string); ok {
				if target != "" {
					globalUsage[target]++
				}
			}
		}
	}

	toKeep := make([]bool, len(o.Instructions))
	keepCount := 0

	for i := 0; i < len(o.Instructions); i++ {
		inst := o.Instructions[i]
		keep := true

		switch inst.Op {
		case OpSetGlobal:
			if name, ok := inst.Arg.(string); ok {
				if count, exists := globalUsage[name]; exists {
					isFuncDef := false
					if i+1 < len(o.Instructions) && o.Instructions[i+1].Op == OpMakeFunc {
						isFuncDef = true
					}

					if count == 0 && !isFuncDef {
						keep = false
						if i > 0 && o.Instructions[i-1].Op == OpConstant {
							toKeep[i-1] = false
						}
					}
				}
			}

		case OpSetLocal:
			if idx, ok := inst.Arg.(float64); ok {
				localIdx := int(idx)
				if count, exists := localUsage[localIdx]; exists && count == 0 {
					keep = false
					if i > 0 && o.Instructions[i-1].Op == OpConstant {
						toKeep[i-1] = false
					}
				}
			}

		case OpMakeFunc:
			keep = true

		case OpConstant:
			if idx, ok := inst.Arg.(float64); ok {
				constIdx := int(idx)
				if constIdx >= 0 && constIdx < len(o.Constants) {
					keep = true
				}
			}

		default:
			keep = true
		}

		toKeep[i] = keep
		if keep {
			keepCount++
		}
	}

	newInstructions := make([]Instruction, 0, keepCount)
	for i, keep := range toKeep {
		if keep {
			newInstructions = append(newInstructions, o.Instructions[i])
		}
	}

	o.Instructions = newInstructions
}

func (o *Optimizer) doGarbageCollection() {
	constantUsed := make([]bool, len(o.Constants))

	for _, inst := range o.Instructions {
		if inst.Op == OpConstant {
			if idx, ok := inst.Arg.(float64); ok {
				constIdx := int(idx)
				if constIdx >= 0 && constIdx < len(o.Constants) {
					constantUsed[constIdx] = true
				}
			}
		}
		if inst.Op == OpMakeFunc {
			if idx, ok := inst.Arg.(float64); ok {
				constIdx := int(idx)
				if constIdx >= 0 && constIdx < len(o.Constants) {
					constantUsed[constIdx] = true
				}
			}
		}
	}

	oldToNew := make([]int, len(o.Constants))
	newConstants := make([]Constant, 0)

	for i, used := range constantUsed {
		if used {
			oldToNew[i] = len(newConstants)
			newConstants = append(newConstants, o.Constants[i])
		} else {
			oldToNew[i] = -1
		}
	}

	for i := range o.Instructions {
		if o.Instructions[i].Op == OpConstant {
			if idx, ok := o.Instructions[i].Arg.(float64); ok {
				oldIdx := int(idx)
				if oldIdx >= 0 && oldIdx < len(oldToNew) && oldToNew[oldIdx] != -1 {
					o.Instructions[i].Arg = float64(oldToNew[oldIdx])
				} else {
					o.Instructions[i].Arg = float64(0)
				}
			}
		}
		if o.Instructions[i].Op == OpMakeFunc {
			if idx, ok := o.Instructions[i].Arg.(float64); ok {
				oldIdx := int(idx)
				if oldIdx >= 0 && oldIdx < len(oldToNew) && oldToNew[oldIdx] != -1 {
					o.Instructions[i].Arg = float64(oldToNew[oldIdx])
				}
			}
		}
	}

	o.Constants = newConstants
}

func isArithmeticOp(op OpCode) bool {
	return op == OpAdd || op == OpSub || op == OpMul || op == OpDiv
}

func performArithmetic(a, b interface{}, op OpCode) (interface{}, bool) {
	var fa, fb float64

	switch v := a.(type) {
	case float64:
		fa = v
	case int:
		fa = float64(v)
	case int64:
		fa = float64(v)
	default:
		return nil, false
	}

	switch v := b.(type) {
	case float64:
		fb = v
	case int:
		fb = float64(v)
	case int64:
		fb = float64(v)
	default:
		return nil, false
	}

	var result float64
	switch op {
	case OpAdd:
		result = fa + fb
	case OpSub:
		result = fa - fb
	case OpMul:
		result = fa * fb
	case OpDiv:
		if fb == 0 {
			return nil, false
		}
		result = fa / fb
	default:
		return nil, false
	}

	if math.Trunc(result) == result && result >= -1<<53 && result < 1<<53 {
		return int(result), true
	}

	return result, true
}

func getTypeString(val interface{}) string {
	switch val.(type) {
	case float64:
		return "number"
	case int, int64:
		return "number"
	case string:
		return "string"
	case bool:
		return "bool"
	case nil:
		return "nil"
	default:
		return "any"
	}
}

func OptimizeBytecode(instructions []Instruction, constants []Constant, sym *SymbolTable) ([]Instruction, []Constant) {
	optimizer := NewOptimizer(instructions, constants, sym)
	return optimizer.Optimize()
}
