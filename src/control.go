package main

type controlFlowKind int

const (
	controlFlowBreak controlFlowKind = iota
	controlFlowReturn
)

type controlFlow struct {
	Kind  controlFlowKind
	Value Value
}

func (flow *controlFlow) Error() string {
	switch flow.Kind {
	case controlFlowBreak:
		return "break outside loop"
	case controlFlowReturn:
		return "return outside function"
	default:
		return "unknown control flow"
	}
}

func asControlFlow(err error) (*controlFlow, bool) {
	flow, ok := err.(*controlFlow)
	return flow, ok
}
