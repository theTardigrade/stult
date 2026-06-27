package main

import (
	"strings"
	"testing"
)

type bindingContractRunMode struct {
	Name string
	Run  func(source string) error
}

func TestBindingContractsKeepSameKind(t *testing.T) {
	source := `value<.> : 1
value : 2
STD.IO.OUTPUT.WRITE_LINE(value)`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			stdout, stderr, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err != nil {
				t.Fatalf("program failed: %v", err)
			}

			if stdout != "2\n" {
				t.Fatalf("unexpected stdout: %q", stdout)
			}

			if stderr != "" {
				t.Fatalf("unexpected stderr: %q", stderr)
			}
		})
	}
}

func TestBindingContractsRejectKindChange(t *testing.T) {
	source := `value<.> : 1
value : "one"`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			_, _, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err == nil {
				t.Fatal("expected program to fail")
			}

			assertErrorContainsForBindingContractTest(t, err, "expects number value, got string value")
		})
	}
}

func TestBindingContractsAllowExplicitAnyContract(t *testing.T) {
	source := `value<*> : 1
value : "one"
value<*> : {:}
STD.IO.OUTPUT.WRITE_LINE(STD.TYPE.IS_MAP(value))`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			stdout, stderr, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err != nil {
				t.Fatalf("program failed: %v", err)
			}

			if stdout != "+\n" {
				t.Fatalf("unexpected stdout: %q", stdout)
			}

			if stderr != "" {
				t.Fatalf("unexpected stderr: %q", stderr)
			}
		})
	}
}

func TestBindingContractsAreAllowedOnImmutableBindings(t *testing.T) {
	source := `NAME<.> : "Example"
LIMIT<*> : 10
STD.IO.OUTPUT.WRITE_LINE(NAME, ":", LIMIT)`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			stdout, stderr, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err != nil {
				t.Fatalf("program failed: %v", err)
			}

			if stdout != "Example:10\n" {
				t.Fatalf("unexpected stdout: %q", stdout)
			}

			if stderr != "" {
				t.Fatalf("unexpected stderr: %q", stderr)
			}
		})
	}
}

func TestBindingContractsRejectRedeclarationOnExistingSameKindBinding(t *testing.T) {
	source := `value<.> : 1
value<.> : 2`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			_, _, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err == nil {
				t.Fatal("expected program to fail")
			}

			assertErrorContainsForBindingContractTest(t, err, "can only be declared when the binding is created")
		})
	}
}

func TestBindingContractsOnMapEntriesRejectKindChange(t *testing.T) {
	source := `settings : {
	.port<.> : 8080
	"mode"<.> : "dev"
}
settings.port : "8080"`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			_, _, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err == nil {
				t.Fatal("expected program to fail")
			}

			assertErrorContainsForBindingContractTest(t, err, "expects number value, got string value")
		})
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
