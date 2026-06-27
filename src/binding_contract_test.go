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
value : {:}
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

func TestBindingContractsRejectSameKindRedeclarationOnExistingBinding(t *testing.T) {
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

func TestBindingContractsRejectAnyRedeclarationOnExistingBinding(t *testing.T) {
	source := `value<*> : 1
value<*> : "one"`

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

func TestBindingContractsRejectMapEntryContractSyntax(t *testing.T) {
	sources := []string{
		`settings : {
	.port<.> : 8080
}`,
		`settings : {
	"mode"<STD.TYPE.STRING> : "dev"
}`,
	}

	for _, source := range sources {
		for _, mode := range bindingContractRunModes() {
			t.Run(mode.Name, func(t *testing.T) {
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

func TestBindingContractsAllowNamedScalarContracts(t *testing.T) {
	source := `value<STD.TYPE.NUMBER> : 11
value : 12
STD.IO.OUTPUT.WRITE_LINE(value)`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			stdout, stderr, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err != nil {
				t.Fatalf("program failed: %v", err)
			}

			if stdout != "12\n" {
				t.Fatalf("unexpected stdout: %q", stdout)
			}

			if stderr != "" {
				t.Fatalf("unexpected stderr: %q", stderr)
			}
		})
	}
}

func TestBindingContractsAllowNamedVoidContract(t *testing.T) {
	source := `value<STD.TYPE.VOID> : _
value : _
STD.IO.OUTPUT.WRITE_LINE(STD.TYPE.IS_VOID(value))
STD.IO.OUTPUT.WRITE_LINE(STD.TYPE.IS_VOID(STD.TYPE.VOID.NEW()))`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			stdout, stderr, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err != nil {
				t.Fatalf("program failed: %v", err)
			}

			if stdout != "+\n+\n" {
				t.Fatalf("unexpected stdout: %q", stdout)
			}

			if stderr != "" {
				t.Fatalf("unexpected stderr: %q", stderr)
			}
		})
	}
}

func TestBindingContractsRejectNamedVoidContractMismatch(t *testing.T) {
	source := `value<STD.TYPE.VOID> : _
value : 0`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			_, _, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err == nil {
				t.Fatal("expected program to fail")
			}

			assertErrorContainsForBindingContractTest(t, err, "expects void value, got number value")
		})
	}
}

func TestBindingContractsRejectNamedScalarContractMismatch(t *testing.T) {
	source := `value<STD.TYPE.NUMBER> : 11
value : "eleven"`

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

func TestBindingContractsRejectInvalidInitialNamedScalarValue(t *testing.T) {
	source := `value<STD.TYPE.NUMBER> : "eleven"`

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

func TestBindingContractsEnforceArrayElementContractsThroughAliases(t *testing.T) {
	source := `values<STD.TYPE.ARRAY<STD.TYPE.STRING>> : {"one", "two"}
alias : values
alias[2] : 3`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			_, _, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err == nil {
				t.Fatal("expected program to fail")
			}

			assertErrorContainsForBindingContractTest(t, err, "expects string value, got number value")
		})
	}
}

func TestBindingContractsEnforceMapValueContractsThroughAliases(t *testing.T) {
	source := `flags<STD.TYPE.MAP<STD.TYPE.BOOL>> : {
	"dev": -
	"prod": +
}
alias : flags
alias["temp"] : "not available"`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			_, _, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err == nil {
				t.Fatal("expected program to fail")
			}

			assertErrorContainsForBindingContractTest(t, err, "expects bool value, got string value")
		})
	}
}

func TestBindingContractsAllowNestedAnyContracts(t *testing.T) {
	source := `values<STD.TYPE.ARRAY<STD.TYPE.ARRAY<*>>> : {
	{1, 2, 3}
	{"four", 5, 6}
	{_}
}
values[0][0] : "one"
values[3] : {+, "mixed", 99}
STD.IO.OUTPUT.WRITE_LINE(STD.TYPE.IS_ARRAY(values[3]))`

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

func TestBindingContractsInferMapValueSameKindContracts(t *testing.T) {
	source := `values<STD.TYPE.MAP<.>> : {
	"a": 1
	"b": 2
	"c": 3
}
values["d"] : "four"`

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

func TestBindingContractsInferArrayElementSameKindContractsFromFirstAppend(t *testing.T) {
	source := `values<STD.TYPE.ARRAY<.>> : {}
values[0] : "first"
values[1] : 2`

	for _, mode := range bindingContractRunModes() {
		t.Run(mode.Name, func(t *testing.T) {
			_, _, err := captureStdoutAndStderrForTest(t, func() error {
				return mode.Run(source)
			})

			if err == nil {
				t.Fatal("expected program to fail")
			}

			assertErrorContainsForBindingContractTest(t, err, "expects string value, got number value")
		})
	}
}

func TestBindingContractsEnforceMultidimensionalNamedContracts(t *testing.T) {
	source := `value<STD.TYPE.ARRAY<STD.TYPE.ARRAY<STD.TYPE.NUMBER>>> : {
	{1, 2, 3}
	{4, 5, 6}
	{7, 8, 9}
}
row : value[0]
row[0] : "one"`

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
