package main

type RuntimeContext struct {
	Args         []string
	BundleAssets *BundleAssetStore
}

func NewRuntimeContext(args []string) *RuntimeContext {
	return &RuntimeContext{
		Args:         append([]string{}, args...),
		BundleAssets: NewEmptyBundleAssetStore(),
	}
}

type RuntimeMode int

const (
	RuntimeModeBytecode RuntimeMode = iota
	RuntimeModeInterpreter
)
