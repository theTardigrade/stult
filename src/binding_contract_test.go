package main

import (
	"strings"
	"testing"
)

type bindingContractRunMode struct {
	Name string
	Run  func(source string) error
}

func TestBindingContractsRejectMapEntryContractSyntax(t *testing.T) {
	sources := map[string]string{
		"leading dot key": `settings : {
	.port<.> : 8080
}`,
		"string key": `settings : {
	"mode"<STD.TYPE.STRING> : "dev"
}`,
	}

	for name, source := range sources {
		for _, mode := range bindingContractRunModes() {
			t.Run(name+"/"+mode.Name, func(t *testing.T) {
				_, _, err := captureStdoutAndStderrForTest(t, func() error {
					return mode.Run(source)
				})

				if err == nil {
					t.Fatal("expected program to fail")
				}

				assertErrorContainsForBindingContractTest(t, err, "map entry contracts are not supported")
			})
		}
	}
}

func TestBindingContractsMustTouchBindingNames(t *testing.T) {
	source := `value <.> : 1`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			_, _, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err == nil {
				t.Fatal("expected program to fail")
			}
		})
	}
}

func TestBindingContractsRejectCompoundAssignmentSyntax(t *testing.T) {
	source := `value<.> :+ 1`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			_, _, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err == nil {
				t.Fatal("expected program to fail")
			}

			assertErrorContainsForBindingContractTest(t, err, "binding contracts can only be declared on plain assignment")
		})
	}
}

func bindingContractRunModes() []bindingContractRunMode {
	return []bindingContractRunMode{
		{
			Name: "interpreter",
			Run: func(source string) error {
				return runSourceString(NewInterpreter(), source, "<binding-contract-test>")
			},
		},
		{
			Name: "bytecode",
			Run: func(source string) error {
				return runSourceStringWithBytecodeVM(NewBytecodeVM(nil), source, "<binding-contract-test>")
			},
		},
	}
}

func assertErrorContainsForBindingContractTest(t *testing.T, err error, expected string) {
	t.Helper()

	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain %q, got: %v", expected, err)
	}
}
