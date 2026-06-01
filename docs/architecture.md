# Architecture

This document describes how Stult is implemented.

It is intended for maintainers and contributors who need to understand the Go codebase, runtime pipeline, bytecode virtual machine (VM), interpreter fallback, manifest handling, bundled executables and test strategy.

For language-level usage and syntax, start with [`../README.md`](../README.md).

## Contents

- [Overview](#overview)
- [Execution model](#execution-model)
  - [Default bytecode runtime](#default-bytecode-runtime)
  - [Interpreter runtime](#interpreter-runtime)
- [Source layout](#source-layout)
- [Command-line entrypoint](#command-line-entrypoint)
- [Lexing](#lexing)
- [Parsing and AST](#parsing-and-ast)
- [Bytecode compiler](#bytecode-compiler)
  - [Chunks](#chunks)
  - [Constants](#constants)
  - [Locals, globals and upvalues](#locals-globals-and-upvalues)
  - [Control flow](#control-flow)
  - [Dynamic loops](#dynamic-loops)
  - [Source spans and disassembly](#source-spans-and-disassembly)
- [Bytecode virtual machine](#bytecode-virtual-machine)
  - [VM state](#vm-state)
  - [Globals and standard library setup](#globals-and-standard-library-setup)
  - [Stack entries](#stack-entries)
  - [Locals and cells](#locals-and-cells)
  - [Functions and closures](#functions-and-closures)
  - [Collection iteration](#collection-iteration)
  - [Operators](#operators)
  - [Runtime errors](#runtime-errors)
- [Tree-walk interpreter](#tree-walk-interpreter)
- [Runtime context](#runtime-context)
- [Values and bindings](#values-and-bindings)
- [Standard library](#standard-library)
- [Manifests](#manifests)
- [Bundled executables](#bundled-executables)
  - [Embedded bundle startup](#embedded-bundle-startup)
  - [Bundle building](#bundle-building)
  - [Bytecode bundles](#bytecode-bundles)
  - [Source/interpreter bundles](#sourceinterpreter-bundles)
- [Errors](#errors)
- [Tests](#tests)
- [Release builds](#release-builds)
- [Maintainer notes](#maintainer-notes)

## Overview

Stult is implemented in Go.

The user-facing runtime is bytecode-first:

1. Source text is read from a file, manifest entry, eval string or embedded bundle.
2. The lexer converts source text into tokens.
3. The parser converts tokens into an AST.
4. The bytecode compiler lowers the AST into a `BytecodeChunk`.
5. The bytecode VM executes the chunk.

The tree-walk interpreter is still kept in the codebase. It is useful as a reference runtime, compatibility fallback and test oracle.

The main invariant is that bytecode execution and interpreter execution should preserve the same Stult language semantics.

## Execution model

### Default bytecode runtime

The default command path is:

```text
stult run ...
```

with bytecode as the default runtime mode.

The default source execution pipeline is:

```text
source
  -> Lexer
  -> Parser
  -> AST
  -> CompileBytecode
  -> BytecodeVM.Run
```

For manifest projects, one VM instance is reused across the files listed by the manifest. This preserves shared global runtime state across files.

For eval strings, the display name is `"<eval>"`.

For bytecode dumping, the source is compiled and formatted, but not executed:

```text
source
  -> Lexer
  -> Parser
  -> AST
  -> CompileBytecode
  -> FormatBytecode
```

`stult dump` is bytecode-only.

### Interpreter runtime

The interpreter can be selected explicitly:

```text
stult run --interpreter ...
```

The interpreter path is:

```text
source
  -> Lexer
  -> Parser
  -> AST
  -> Interpreter.EvalProgram
```

For manifest projects, one interpreter instance is reused across the manifest files. This preserves the same shared-state model as the bytecode runtime.

The interpreter remains important because it is the source-of-truth implementation to compare against when changing bytecode compiler or VM behavior.

## Source layout

The main implementation lives in `src/`.

```text
main.go                    CLI entrypoint and source/manifest execution helpers
main_flags.go              CLI mode parsing and usage text

lexer.go                   source text to tokens
token.go                   token definitions and source locations

parser*.go                 tokens to AST
ast.go                     AST node definitions

bytecode.go                bytecode opcodes, chunks, compiler and disassembler
bytecode_vm.go             bytecode VM execution loop and runtime behavior
bytecode_bundle.go         bytecode encoding/decoding for bundled executables

interpreter*.go            tree-walk interpreter
environment.go             lexical environments and bindings
control.go                 internal break/return control flow

value*.go                  runtime value types, formatting, comparison and operators

runtime_context.go         process/runtime data passed to builtins
std*.go                    standard-library maps and builtin functions

manifest.go                manifest loading and normalization

bundle.go                  embedded bundle detection/loading
bundle_archive.go          bundle archive creation
bundle_build.go            build command
bundle_footer.go           bundle footer format

examples_test.go           example parity tests
```

Outside `src/`:

```text
examples/                  public example Stult programs
docs/                      documentation
util/build_helper.go       local and release build helper
.github/workflows/         release workflow
```

Some bytecode-related code currently lives in larger files rather than many small packages. That is deliberate for now: the implementation is still changing and keeping the compiler/disassembler together reduces churn while the representation stabilizes.

## Command-line entrypoint

`main.go` starts by checking whether the current executable contains an embedded bundle:

```text
run()
  -> runEmbeddedBundleIfPresent()
```

If an embedded bundle is present, the executable behaves as the bundled program and does not continue into the development CLI.

If no bundle is present, the CLI dispatches explicit subcommands:

```text
stult run ...
stult dump ...
stult build ...
```

The `run` command parses the runtime mode:

```text
--bytecode       default
--interpreter    explicit tree-walk interpreter mode
```

Then it resolves the target:

```text
no target        search upward for a manifest
file.stult       run one source file
directory        find manifest from that directory
manifest file    run that manifest directly
-e / --eval      run a source string
```

Program arguments after the source, manifest or directory target are stored in `RuntimeContext.Args` and exposed to Stult code as `STD["SYSTEM"]["ARGS"]`.

The `dump` command does not use runtime modes. It compiles to bytecode and prints a disassembly.

The `build` command creates a standalone executable by appending a bundle archive and footer to the current runner executable.

## Lexing

The lexer is responsible for converting source text into tokens.

It tracks line and column information so parser and runtime errors can refer back to source locations.

The lexer recognizes:

```text
identifiers
numbers
strings
boolean literals
void
operators
assignment operators
compound assignment operators
delimiters
line comments
bounded comments
```

Identifier mutability is detected lexically:

```text
uppercase-only names         immutable
names containing lowercase   mutable
```

The parser and runtime later use that token-level mutability information when creating bindings, locals and parameters.

Comments are discarded from the token stream. Source positions are preserved.

## Parsing and AST

The parser builds the abstract syntax tree used by both runtime implementations.

The parser does not evaluate code and does not compile to bytecode directly. Its job is to produce a structured representation of the source program.

Important AST categories include:

```text
statements
expressions
assignments
compound assignments
conditionals
dynamic loops
function literals
function calls
array literals
map literals
range segments
index expressions
outer-name expressions
```

Because the bytecode compiler and interpreter share the AST, language syntax changes must usually be handled in both runtime paths.

## Bytecode compiler

The bytecode compiler lowers the AST into a `BytecodeChunk`.

A chunk is the executable bytecode unit used by the VM. Top-level source files compile to top-level chunks. Function literals compile to nested function chunks.

The compiler is responsible for preserving Stult semantics while lowering higher-level AST constructs into explicit VM instructions.

### Chunks

A `BytecodeChunk` contains:

```text
name
instructions
constants
locals
upvalues
functions
source spans
```

The chunk is both the execution unit and the disassembler unit.

Nested function chunks are stored in the parent chunk's function table and referenced by bytecode instructions.

### Constants

The compiler stores constants in a per-chunk constant table and reuses existing entries when the same constant appears more than once.

Typical constants include:

```text
numbers
strings
names
```

Name constants are used by load/store operations and are also useful for disassembly comments.

### Locals, globals and upvalues

The compiler tracks lexical scopes.

At top level, plain assignment creates or updates globals.

Inside functions or nested scopes, plain assignment usually maps to locals in the current scope.

Outer access has distinct semantics:

```text
name       read through normal lookup rules
@name      read nearest outer binding, skipping current scope
@name : v  write nearest mutable outer binding
```

The compiler handles this through local scope tracking, parent compiler links and upvalue registration.

The important invariant is that bytecode must preserve the interpreter's `@` behavior exactly.

Top-level block scopes are a special case because they can have an outer context even without a parent function compiler. The compiler therefore needs to know whether it has an outer context when deciding whether an outer-name operation should fall back to a global operation.

### Control flow

Conditionals, loops, break and early return are lowered to jumps and returns.

The compiler emits placeholder jump operands and patches them once target instruction indexes are known.

Logical `&` and `|` are compiled with short-circuit behavior rather than eager binary evaluation.

Early return from functions compiles to a return path that exits the current function chunk.

A bare `^` inside loops compiles to a loop break.

### Dynamic loops

Stult uses one loop syntax for condition loops and collection loops:

```text
((condition_or_collection)) { ... }
```

The bytecode compiler emits runtime-dispatched loop code.

At runtime, the VM decides whether the loop expression is a collection. If it is a collection, the loop uses iterator opcodes. Otherwise, it behaves as a condition loop.

Collection loops support up to four parameters:

```text
value
key
collection
position
```

The compiler allocates locals for those parameters and resets loop-scope locals on each iteration.

### Source spans and disassembly

A source span is the source file, line and column range that produced a bytecode instruction.

Compiler instructions can carry source-span metadata as `BytecodeSourceSpan`.

Source spans are used for runtime error messages and by the disassembler.

`FormatBytecode` produces a human-readable view of a chunk, including instructions, operands, locals, functions and source-map information.

This output is intended for compiler/VM debugging. It is not a stable serialized bytecode format.

## Bytecode virtual machine

The bytecode VM executes `BytecodeChunk` values.

The VM is stack-based. Instructions read and write stack entries, locals, globals, upvalues and iterator state.

The VM is intentionally close to the language semantics rather than a low-level native-code compiler.

### VM state

The VM owns:

```text
current chunk
instruction pointer
operand stack
globals
locals
upvalues
iterator stack
runtime context
local-index caches
reset-local caches
```

A single VM instance can run multiple chunks, which is important for manifest execution.

### Globals and standard library setup

VM globals are initialized with `STD`.

`STD` is created from the shared runtime context:

```text
RuntimeContext
  -> NewStdMap(runtime)
  -> STD global binding
```

This makes command-line arguments and other process context available to standard-library builtins without passing the whole interpreter or VM around.

### Stack entries

VM stack entries wrap values.

Some stack entries carry extra metadata, for example to mark range segments while building arrays.

Most instructions push and pop `Value` instances. Some helper methods resolve specialized values before applying operators or truthiness checks.

### Locals and cells

Locals are stored as VM cells.

A local cell tracks:

```text
value
initialized status
immutability
```

Immutability must be enforced for uppercase-only names and immutable parameters.

The compiler emits `RESET_LOCALS` instructions for scoped blocks so locals from a previous iteration or block execution do not leak into later executions of the same local slot.

The VM caches local indexes by chunk and reset depth to reduce repeated lookup work.

### Functions and closures

Function literals compile to bytecode function metadata and nested chunks.

At runtime, function values need:

```text
function chunk
parameters
optional variadic parameter
captured upvalues
```

The VM saves and restores execution state when running bytecode functions. This avoids constructing a separate VM for every function call while still isolating chunk, stack, locals, upvalues and iterator state for the call.

The interpreter has its own function representation that stores the AST body and defining environment.

### Collection iteration

The VM has explicit iterator state for collection loops.

Array, string and map iteration share the same loop machinery but produce different keys and values:

```text
array   key is numeric index, value is element
string  key is numeric index, value is one-character string
map     key is string key, value is map entry value
```

Map iteration is deterministic only to the degree enforced by the runtime helper used to find the next key. Any change to map iteration order must be checked against examples and interpreter parity.

### Operators

The VM implements common operators directly in opcode-specific helper functions.

This avoids dispatching every arithmetic or comparison operation back through a generic AST operator evaluator.

Operator semantics must still match the interpreter exactly.

When changing operator behavior, check both:

```text
value_operator.go / value_compare.go
bytecode_vm.go
```

and run `go test ./...`, which includes the example-code tests that compare interpreter and bytecode behavior.

### Runtime errors

VM runtime errors should include:

```text
source display name
line
column
instruction index
human-readable runtime message
```

The instruction index is useful for VM debugging.

The source location is useful for Stult users.

Runtime error text is part of user-visible behavior, so changes should be made carefully.

## Tree-walk interpreter

The interpreter evaluates the AST directly.

It uses chained `Environment` values for lexical scope.

The interpreter remains useful because it is simpler to reason about than the bytecode compiler plus VM. When bytecode behavior differs from interpreter behavior, the interpreter should usually be treated as the reference unless the interpreter is known to be wrong.

The interpreter path is selected with:

```text
stult run --interpreter ...
```

It is also used by source/interpreter bundles.

## Runtime context

`RuntimeContext` carries process-level runtime data that should be shared without depending on a concrete interpreter or VM type.

Currently, it stores program arguments.

Builtins receive:

```go
func(runtime *RuntimeContext, args []Value) (Value, error)
```

This lets both runtime implementations call the same builtins.

The standard library should not depend on interpreter internals or bytecode VM internals.

## Values and bindings

Runtime values are represented by `Value`.

Important value kinds include:

```text
void
number
bool
string
array
map
function
builtin function
```

Specialized or mutable values may be resolved before formatting, comparison or operator application.

Bindings wrap values with mutability metadata.

The language-level binding mutability rule is:

```text
uppercase-only identifiers are immutable
identifiers containing lowercase are mutable
```

The implementation must enforce this rule consistently for:

```text
global bindings
local cells
function parameters
collection entries where applicable
outer writes
```

Binding immutability and collection freezing are separate concepts.

Binding immutability controls whether a name or map entry can be rebound. Collection freezing controls whether the contents of an existing array, map or string can be internally modified.

Arrays, maps and strings carry collection-level immutability flags. The standard-library function `STD["TYPE"]["COLLECTION"]["FREEZE"]` sets those flags deeply and in place for arrays, maps and strings, then returns the same collection value. `STD["TYPE"]["COLLECTION"]["IS_FROZEN"]` reports whether a collection value currently has its collection-level immutability flag set.

Mutation helpers for arrays, maps and strings must check collection-level immutability before changing collection contents. User-facing errors for frozen collection mutation should describe the collection as frozen, rather than describing the binding as immutable.

## Standard library

The standard library is exposed through the immutable global binding `STD`.

`STD` is a map of maps:

```text
STD["IO"]
STD["SYSTEM"]
STD["FILE"]
STD["PATH"]
STD["TIME"]
STD["MATH"]
STD["TYPE"]
STD["DATA"]
```

Builtins are Go functions wrapped as Stult callable values.

Standard-library functions should return Stult values and errors, not print internal Go details.

Because both runtime modes use the same standard library, changes to builtins usually affect both bytecode and interpreter behavior.

Collection helpers live under:

```text
STD["TYPE"]["COLLECTION"]
```

These helpers operate on arrays, maps and strings where appropriate.

`STD["TYPE"]["COLLECTION"]["FREEZE"]` deeply freezes arrays, maps and strings in place and returns the frozen collection. This means aliases to nested collections observe the frozen state too.

`STD["TYPE"]["COLLECTION"]["IS_FROZEN"]` returns a boolean for arrays, maps and strings. It returns false for non-collection values.

## Manifests

A manifest lists source files to run in order.

Manifest loading normalizes STULTON and JSON manifests into one Go representation with `RunFiles`.

Manifest execution preserves one shared runtime state across all listed files.

For bytecode mode, that means one VM instance runs the compiled chunks in order.

For interpreter mode, that means one interpreter instance evaluates parsed programs in order.

Manifest order is therefore semantically important.

A file listed earlier can define bindings used by later files.

## Bundled executables

Bundled executables are implemented by appending an archive and footer to a normal Stult executable.

At runtime, `runEmbeddedBundleIfPresent` checks the current executable before normal CLI dispatch. If a bundle is found, the executable runs the embedded program directly.

### Embedded bundle startup

Startup order is:

```text
main
  -> run
  -> runEmbeddedBundleIfPresent
      -> openEmbeddedBundle
      -> runEmbeddedBundle
```

If no embedded bundle exists, normal CLI dispatch continues.

If an embedded bundle exists, CLI dispatch does not run. The executable behaves as the bundled application.

### Bundle building

The build command reads the currently running executable, strips any existing embedded bundle if necessary, appends a new bundle archive and writes a footer containing the bundle size/magic data.

The build command refuses to overwrite the executable currently running the build.

The default build mode is bytecode.

### Bytecode bundles

A bytecode bundle embeds:

```text
manifest
bytecode marker
bytecode run map
encoded bytecode chunks
```

It does not need original `.stult` source files at runtime.

For manifest entries, the run map connects the original manifest path to the embedded bytecode path. This is especially important for absolute manifest run paths, because an absolute filesystem path cannot be used directly as a path inside the embedded archive.

At runtime, a bytecode bundle decodes each embedded bytecode chunk and runs it on one shared VM.

### Source/interpreter bundles

A source/interpreter bundle embeds:

```text
manifest
.stult source files
```

At runtime, it loads source from the embedded archive and evaluates it through the tree-walk interpreter.

Absolute manifest entries are not loaded from the embedded archive in the same way as relative entries. They may be read from the target machine filesystem, so source/interpreter bundles should prefer relative manifest paths.

## Errors

Stult has three main error phases:

```text
lexing/parsing
bytecode compilation
runtime execution
```

Parser errors should include the source display name and the original source text context.

Bytecode compiler errors should identify the construct or source span that cannot be compiled.

Bytecode runtime errors should include source location and instruction index.

Interpreter runtime errors should include source context where possible.

Do not replace user-facing Stult error messages with raw Go panics.

## Tests

The Go test suite includes example-based integration tests.

The important parity tests run examples through both runtime modes:

```text
interpreter
bytecode VM
```

For deterministic examples, stdout and stderr should match exactly.

For nondeterministic examples, such as time-driven animation, the test can still run both modes but skip exact stdout comparison.

The purpose of these tests is not just coverage. The examples are public documentation, so they are also compatibility fixtures.

When changing compiler, VM, interpreter, standard library, manifests or bundling, run:

```bash
go test ./...
```

## Release builds

Release builds are handled by GitHub Actions.

The release workflow:

```text
checks out the repository
sets up Go
runs tests
builds dist executables with util/build_helper.go
generates checksums
uploads dist files to the GitHub release
```

Local dist builds use:

```bash
go run ./util/build_helper.go dist
```

Local single-platform builds use:

```bash
go run ./util/build_helper.go local
```

## Maintainer notes

When changing syntax, update all affected stages:

```text
lexer
parser
AST
interpreter
bytecode compiler
bytecode VM
examples
docs
tests
```

When changing runtime semantics, check both runtime implementations.

When changing standard-library behavior, remember that builtins are shared by both the interpreter and bytecode VM through `RuntimeContext`.

When changing collection mutability, keep binding immutability, map-entry immutability and frozen collection state separate. The public language distinction is that immutable bindings cannot be rebound, while frozen collections cannot be internally modified.

When changing manifest behavior, check:

```text
normal bytecode runs
normal interpreter runs
bytecode bundles
source/interpreter bundles
example parity tests
```

When changing bundle behavior, test both:

```text
stult build ...
stult build --interpreter ...
```

and run the generated executables directly.

When changing the bytecode format or disassembler, update `stult dump` output expectations accordingly.