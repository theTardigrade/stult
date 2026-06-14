package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

const bytecodeBundleModeFilename = ".stult-bytecode-bundle"
const bytecodeBundleModeFileContents = "bytecode\n"
const bytecodeBundleDir = ".stult-bytecode"
const bytecodeBundleRunMapFilename = ".stult-bytecode/run-map.json"
const bytecodeBundleFileExtension = ".stultbc.json"

type bundledBytecodeRunMap struct {
	RunFiles map[string]string `json:"run_files"`
}

type bundledBytecodeChunk struct {
	Name         string
	Instructions []BytecodeInstruction
	Constants    []bundledBytecodeValue
	Locals       []BytecodeLocal
	Upvalues     []BytecodeUpvalue
	Functions    []bundledBytecodeFunction
	SourceSpans  []BytecodeSourceSpan
}

type bundledBytecodeFunction struct {
	Name              string
	Parameters        []BytecodeParameter
	VariadicParameter *BytecodeParameter
	Upvalues          []BytecodeUpvalue
	Chunk             *bundledBytecodeChunk
}

type bundledBytecodeValue struct {
	Kind            ValueKind
	Number          string
	Bool            bool
	String          string
	StringImmutable bool
}

func embeddedBundleWantsBytecode(files fs.FS) bool {
	_, err := fs.ReadFile(files, bytecodeBundleModeFilename)

	return err == nil
}

func bytecodeBundleRunFileKey(filename string) string {
	if filepath.IsAbs(filename) {
		return filepath.ToSlash(filepath.Clean(filename))
	}

	return cleanFSPath(filename)
}

func bytecodeBundlePathForSource(filename string) string {
	fsPath := cleanFSPath(filename)

	return path.Join(bytecodeBundleDir, fsPath+bytecodeBundleFileExtension)
}

func bytecodeBundlePathForAbsoluteSource(filename string) string {
	cleaned := filepath.Clean(filename)
	hash := sha256.Sum256([]byte(cleaned))

	return path.Join(
		bytecodeBundleDir,
		"__absolute",
		hex.EncodeToString(hash[:])+bytecodeBundleFileExtension,
	)
}

func encodeBundledBytecodeRunMap(runFiles map[string]string) ([]byte, error) {
	return json.MarshalIndent(bundledBytecodeRunMap{
		RunFiles: runFiles,
	}, "", "\t")
}

func readBundledBytecodeRunMap(files fs.FS) (map[string]string, error) {
	bytes, err := fs.ReadFile(files, bytecodeBundleRunMapFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}

		return nil, fmt.Errorf("Could not read bundled bytecode run map: %w", err)
	}

	var runMap bundledBytecodeRunMap

	if err := json.Unmarshal(bytes, &runMap); err != nil {
		return nil, fmt.Errorf("Could not decode bundled bytecode run map: %w", err)
	}

	if runMap.RunFiles == nil {
		return map[string]string{}, nil
	}

	return runMap.RunFiles, nil
}

func compileSourceBytesToBundledBytecode(sourceBytes []byte, displayName string) ([]byte, error) {
	source := string(sourceBytes)

	lexer := NewLexer(source)
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		return nil, formatParserErrors(displayName, source, parser.Errors())
	}

	chunk, err := CompileBytecode(program, displayName)
	if err != nil {
		return nil, fmt.Errorf("Could not compile bytecode for %q: %w", displayName, err)
	}

	bytes, err := encodeBundledBytecodeChunk(chunk)
	if err != nil {
		return nil, fmt.Errorf("Could not encode bytecode for %q: %w", displayName, err)
	}

	return bytes, nil
}

func encodeBundledBytecodeChunk(chunk *BytecodeChunk) ([]byte, error) {
	bundled, err := bundledBytecodeChunkFromChunk(chunk)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(bundled, "", "\t")
}

func decodeBundledBytecodeChunk(bytes []byte) (*BytecodeChunk, error) {
	var bundled bundledBytecodeChunk

	if err := json.Unmarshal(bytes, &bundled); err != nil {
		return nil, err
	}

	return bundled.toBytecodeChunk()
}

func bundledBytecodeChunkFromChunk(chunk *BytecodeChunk) (*bundledBytecodeChunk, error) {
	if chunk == nil {
		return nil, fmt.Errorf("cannot bundle nil bytecode chunk")
	}

	constants := make([]bundledBytecodeValue, 0, len(chunk.Constants))
	for index, constant := range chunk.Constants {
		value, err := bundledBytecodeValueFromValue(constant)
		if err != nil {
			return nil, fmt.Errorf("constant %d: %w", index, err)
		}

		constants = append(constants, value)
	}

	functions := make([]bundledBytecodeFunction, 0, len(chunk.Functions))
	for index, function := range chunk.Functions {
		bundledFunction, err := bundledBytecodeFunctionFromFunction(function)
		if err != nil {
			return nil, fmt.Errorf("function %d: %w", index, err)
		}

		functions = append(functions, bundledFunction)
	}

	return &bundledBytecodeChunk{
		Name:         chunk.Name,
		Instructions: append([]BytecodeInstruction{}, chunk.Instructions...),
		Constants:    constants,
		Locals:       append([]BytecodeLocal{}, chunk.Locals...),
		Upvalues:     append([]BytecodeUpvalue{}, chunk.Upvalues...),
		Functions:    functions,
		SourceSpans:  append([]BytecodeSourceSpan{}, chunk.SourceSpans...),
	}, nil
}

func bundledBytecodeFunctionFromFunction(
	function BytecodeFunction,
) (bundledBytecodeFunction, error) {
	chunk, err := bundledBytecodeChunkFromChunk(function.Chunk)
	if err != nil {
		return bundledBytecodeFunction{}, err
	}

	return bundledBytecodeFunction{
		Name:              function.Name,
		Parameters:        append([]BytecodeParameter{}, function.Parameters...),
		VariadicParameter: cloneBytecodeParameterPointer(function.VariadicParameter),
		Upvalues:          append([]BytecodeUpvalue{}, function.Upvalues...),
		Chunk:             chunk,
	}, nil
}

func cloneBytecodeParameterPointer(parameter *BytecodeParameter) *BytecodeParameter {
	if parameter == nil {
		return nil
	}

	clone := *parameter

	return &clone
}

func bundledBytecodeValueFromValue(value Value) (bundledBytecodeValue, error) {
	value = resolveSpecializedValue(value)

	switch value.Kind {
	case ValueVoid:
		return bundledBytecodeValue{Kind: ValueVoid}, nil

	case ValueNumber:
		if value.Number == nil {
			return bundledBytecodeValue{}, fmt.Errorf("number constant is invalid")
		}

		return bundledBytecodeValue{
			Kind:   ValueNumber,
			Number: numberToBigFloat(value.Number).Text('g', -1),
		}, nil

	case ValueBool:
		return bundledBytecodeValue{
			Kind: ValueBool,
			Bool: value.Bool,
		}, nil

	case ValueString:
		if value.Text == nil {
			return bundledBytecodeValue{}, fmt.Errorf("string constant is invalid")
		}

		return bundledBytecodeValue{
			Kind:            ValueString,
			String:          value.Text.String(),
			StringImmutable: value.Text.IsImmutable,
		}, nil

	default:
		return bundledBytecodeValue{}, fmt.Errorf(
			"unsupported bytecode constant kind %d",
			value.Kind,
		)
	}
}

func (chunk bundledBytecodeChunk) toBytecodeChunk() (*BytecodeChunk, error) {
	constants := make([]Value, 0, len(chunk.Constants))
	for index, constant := range chunk.Constants {
		value, err := constant.toValue()
		if err != nil {
			return nil, fmt.Errorf("constant %d: %w", index, err)
		}

		constants = append(constants, value)
	}

	functions := make([]BytecodeFunction, 0, len(chunk.Functions))
	for index, function := range chunk.Functions {
		bytecodeFunction, err := function.toBytecodeFunction()
		if err != nil {
			return nil, fmt.Errorf("function %d: %w", index, err)
		}

		functions = append(functions, bytecodeFunction)
	}

	return &BytecodeChunk{
		Name:         chunk.Name,
		Instructions: append([]BytecodeInstruction{}, chunk.Instructions...),
		Constants:    constants,
		Locals:       append([]BytecodeLocal{}, chunk.Locals...),
		Upvalues:     append([]BytecodeUpvalue{}, chunk.Upvalues...),
		Functions:    functions,
		SourceSpans:  append([]BytecodeSourceSpan{}, chunk.SourceSpans...),
	}, nil
}

func (function bundledBytecodeFunction) toBytecodeFunction() (BytecodeFunction, error) {
	if function.Chunk == nil {
		return BytecodeFunction{}, fmt.Errorf("function chunk is missing")
	}

	chunk, err := function.Chunk.toBytecodeChunk()
	if err != nil {
		return BytecodeFunction{}, err
	}

	return BytecodeFunction{
		Name:              function.Name,
		Parameters:        append([]BytecodeParameter{}, function.Parameters...),
		VariadicParameter: cloneBytecodeParameterPointer(function.VariadicParameter),
		Upvalues:          append([]BytecodeUpvalue{}, function.Upvalues...),
		Chunk:             chunk,
	}, nil
}

func (value bundledBytecodeValue) toValue() (Value, error) {
	switch value.Kind {
	case ValueVoid:
		return NewVoidValue(), nil

	case ValueNumber:
		number, err := NewNumberValueFromString(value.Number)
		if err != nil {
			return Value{}, err
		}

		return number, nil

	case ValueBool:
		return NewBoolValue(value.Bool), nil

	case ValueString:
		return NewStringValueWithImmutability(value.String, value.StringImmutable), nil

	default:
		return Value{}, fmt.Errorf("unsupported bytecode constant kind %d", value.Kind)
	}
}

func runBundledBytecodeManifestFromFS(
	vm *BytecodeVM,
	files fs.FS,
	runFiles []string,
) error {
	runMap, err := readBundledBytecodeRunMap(files)
	if err != nil {
		return err
	}

	for _, filename := range runFiles {
		if err := runBundledBytecodeSourceFromFS(vm, files, filename, filename, runMap); err != nil {
			return err
		}
	}

	return nil
}

func runBundledBytecodeSourceFromFS(
	vm *BytecodeVM,
	files fs.FS,
	filename string,
	displayName string,
	runMap map[string]string,
) error {
	bytecodePath := bytecodeBundlePathForRunFile(filename, runMap)

	bytecodeBytes, err := fs.ReadFile(files, bytecodePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(
				"bytecode bundle is missing compiled bytecode for %q at %q",
				displayName,
				bytecodePath,
			)
		}

		return fmt.Errorf("Could not read bundled bytecode %q: %w", bytecodePath, err)
	}

	chunk, err := decodeBundledBytecodeChunk(bytecodeBytes)
	if err != nil {
		return fmt.Errorf("Could not decode bundled bytecode %q: %w", bytecodePath, err)
	}

	if _, err := vm.Run(chunk); err != nil {
		return fmt.Errorf("Bytecode runtime error in %q: %w", displayName, err)
	}

	return nil
}

func bytecodeBundlePathForRunFile(filename string, runMap map[string]string) string {
	key := bytecodeBundleRunFileKey(filename)

	if bytecodePath, ok := runMap[key]; ok {
		return cleanFSPath(bytecodePath)
	}

	if filepath.IsAbs(filename) {
		return bytecodeBundlePathForAbsoluteSource(filename)
	}

	return bytecodeBundlePathForSource(filename)
}
