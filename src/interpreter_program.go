package main

import "fmt"

func (i *Interpreter) EvalProgram(program *Program) error {
	for _, stmt := range program.Statements {
		if _, err := i.evalStatement(stmt); err != nil {
			if flow, ok := asControlFlow(err); ok {
				switch flow.Kind {
				case controlFlowBreak:
					return fmt.Errorf("break used outside loop")
				case controlFlowReturn:
					return fmt.Errorf("return used outside function")
				}
			}

			return err
		}
	}

	return nil
}
