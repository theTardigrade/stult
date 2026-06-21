package main

import "fmt"

type bytecodeVMErrorHandler struct {
	CatchIP        int
	StackSize      int
	IteratorCount  int
	CurrentDotMap  *Map
	DotMapStackLen int
}

func (vm *BytecodeVM) tryStart(catchIP int) error {
	if vm.chunk == nil {
		return fmt.Errorf("no active bytecode chunk")
	}

	if catchIP < 0 || catchIP > len(vm.chunk.Instructions) {
		return fmt.Errorf("try catch target %d out of bounds", catchIP)
	}

	vm.errorHandlers = append(vm.errorHandlers, bytecodeVMErrorHandler{
		CatchIP:        catchIP,
		StackSize:      len(vm.stack),
		IteratorCount:  len(vm.iterators),
		CurrentDotMap:  vm.currentDotMap,
		DotMapStackLen: len(vm.dotMapStack),
	})

	return nil
}

func (vm *BytecodeVM) tryEnd() error {
	if len(vm.errorHandlers) == 0 {
		return fmt.Errorf("no active try handler to end")
	}

	vm.errorHandlers = vm.errorHandlers[:len(vm.errorHandlers)-1]
	return nil
}

func (vm *BytecodeVM) handleRuntimeError(err error) bool {
	if len(vm.errorHandlers) == 0 {
		return false
	}

	last := len(vm.errorHandlers) - 1
	handler := vm.errorHandlers[last]
	vm.errorHandlers = vm.errorHandlers[:last]

	if handler.StackSize < 0 || handler.StackSize > len(vm.stack) {
		handler.StackSize = 0
	}

	if handler.IteratorCount < 0 || handler.IteratorCount > len(vm.iterators) {
		handler.IteratorCount = 0
	}

	if handler.DotMapStackLen < 0 || handler.DotMapStackLen > len(vm.dotMapStack) {
		handler.DotMapStackLen = 0
	}

	vm.stack = vm.stack[:handler.StackSize]
	vm.iterators = vm.iterators[:handler.IteratorCount]
	vm.currentDotMap = handler.CurrentDotMap
	vm.dotMapStack = vm.dotMapStack[:handler.DotMapStackLen]
	vm.pushValue(NewStringValue(err.Error()))
	vm.ip = handler.CatchIP

	return true
}
