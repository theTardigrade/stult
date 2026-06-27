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
  - [Try-catch](#try-catch)
  - [Range-backed loop optimisation](#range-backed-loop-optimisation)
  - [Source spans and disassembly](#source-spans-and-disassembly)
- [Bytecode virtual machine](#bytecode-virtual-machine)
  - [VM state](#vm-state)
  - [Globals and standard library setup](#globals-and-standard-library-setup)
  - [Stack entries](#stack-entries)
  - [Locals and cells](#locals-and-cells)
  - [Functions and closures](#functions-and-closures)
  - [Collection iteration](#collection-iteration)
  - [Function loops](#function-loops)
  - [Try handlers](#try-handlers)
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

1. Source text is read from a file, standard input, manifest entry, eval string or embedded bundle.
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

For standard-input source, the display name is `"<stdin>"`. For eval strings, the display name is `"<eval>"`.

For bytecode dumping, the source is compiled and formatted, but not executed:

```text
source
  -> Lexer
  -> Parser
  -> AST
  -> CompileBytecode
  -> FormatBytecode
  -> stdout or -o/--output file
```

`stult dump` is bytecode-only. By default it prints the formatted disassembly to standard output. If `-o` or `--output` is provided, it writes the same disassembly to that file instead and creates parent directories as needed.

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

bytecode.go                bytecode opcodes, chunks and core bytecode types
bytecode_compile*.go       bytecode compiler
bytecode_disassemble.go    bytecode formatting and disassembly
bytecode_vm*.go            bytecode VM execution and runtime behaviour

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
bundle_bytecode.go         bytecode encoding/decoding for bundled executables
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
-                read and run source from standard input
-e / --eval      run a source string
```

Program arguments after the source, `-` stdin target, manifest or directory target are stored in `RuntimeContext.Args` and exposed to Stult code as `STD["SYSTEM"]["ARGS"]`.

The `dump` command does not use runtime modes. It compiles to bytecode and prints a disassembly, or writes that disassembly to `-o`/`--output` when an output file is provided.

The `build` command creates a standalone executable by appending a bundle archive and footer to the current runner executable. It accepts `--output` and the short alias `-o` to choose the generated executable path.

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

A single `.` token is used for dot access. Number and range lexing still handle dot-prefixed numbers, apostrophe-separated numbers, percentage-suffixed numbers and range operators specially, so `.5`, `1'000`, `1.5`, `50%`, `..` and `...` remain distinct from dot access.

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
conditional expressions
match expressions
try-catch statements
dynamic loops
function literals
optional function parameters
function calls
array literals
map literals
frozen collection literals
range segments
index expressions
dot access
leading-dot receiver expressions
outer-name expressions
```

Function calls are parsed as postfix expressions when an opening parenthesis appears on the same line as the callee expression. The idiomatic form keeps the callee and `(` touching, but horizontal whitespace before the `(` is accepted. A newline before the `(` keeps its normal statement-separating meaning and does not start a call. Call arguments may use suffix spread syntax, such as `values...`, which the parser records on the call argument rather than as a standalone expression. The interpreter and bytecode VM both evaluate call arguments left to right and expand spread arrays into positional arguments at call time.

Index expressions are parsed as postfix expressions when an opening square bracket appears on the same line as the indexed expression. The idiomatic form keeps the indexed expression and `[` touching, but horizontal whitespace before the `[` is accepted. A newline before the `[` keeps its normal statement-separating meaning and does not start an index expression.

Range indexing reuses the range-segment grammar inside an index expression. The parser represents `object[start..end:step]` separately from ordinary indexing so interpreter and bytecode paths can return new array or string slice values while keeping map indexing unchanged. Range index expressions are expressions only; they are not valid assignment targets.

Dot access is parsed as syntax sugar for string-key indexing. The parser lowers `object.key` to the same AST shape as `object["key"]`, preserving the identifier spelling as the string key. This keeps interpreter and bytecode behaviour aligned with ordinary map indexing.

Leading-dot field access, such as `.name`, is also lowered to an index expression, but its object is a `LeadingDotReceiverExpression`. At runtime that receiver resolves to the nearest active or captured map literal. This means leading-dot access can be used while evaluating map entry values, and functions created inside maps capture that map for later leading-dot access. If no such map exists, or if the nearest map does not contain the requested key, evaluation raises a runtime error.

Map literals accept ordinary string keys and leading-dot map keys. A leading-dot map key, such as `.name : value`, is parsed as the string key `"name"` while preserving the usual map-entry mutability rules derived from the resulting key string. Bare identifier-shaped map keys remain invalid.

Frozen collection literals are written by prefixing an array, map or string literal with `~`. The parser records this on the literal AST node, and the interpreter or bytecode VM creates the collection normally before applying the shared shallow-freeze helper. This keeps frozen literal syntax independent from standard-library lookup while matching the default shallow behaviour of `STD["TYPE"]["COLLECTION"]["FREEZE"]`.

Conditional expressions are represented as `ConditionalExpression` AST nodes. They require a parenthesised condition followed by a same-line `:` branch list, written idiomatically as `(condition):(when_true|when_false)`. Horizontal whitespace may appear around the `:`. In multiline branch lists, the `|` branch separator must appear at the end of the true-branch line. Unlike dot access, conditional expressions are not lowered to an existing AST shape because only one branch may be evaluated.

Match expressions are represented as `MatchExpression` AST nodes. They require a parenthesised subject followed by a same-line `:` arm list, written idiomatically as `(subject):{ "case": result _: fallback }`. Horizontal whitespace may appear around the `:`. Match arms store scalar literal patterns separately from their result expressions. The default `_` arm is stored separately so explicit arms can be checked before the default arm, even when `_` appears earlier in source.

The expression parser treats a trailing binary operator as an explicit line-continuation marker. This applies to arithmetic, comparison, equality and logical operators. A newline before a binary operator does not continue the previous expression; in delimited expression contexts it is rejected, and in list-like contexts it starts a new item or statement according to the surrounding grammar.

Function literal parameters are represented with parameter metadata rather than plain identifier tokens. Ordinary parameters can be required or optional. Optional parameters are written with `?` in source and receive void when omitted at call time.

Variadic parameters are still represented separately because they have different semantics: they receive an array of remaining arguments, or an empty array when no remaining arguments exist.

Try-catch statements are represented as `TryCatchStatement` AST nodes. The parser accepts `?{ ... }` as an error-suppression try block and `?{ ... }|{ ... }` as a try block with a catch block. Horizontal whitespace may appear between `?` and `{`, and around the block separator `|`, as long as the tokens stay on the same line.

Because the bytecode compiler and interpreter share the AST, language syntax changes must usually be handled in both runtime paths.

## Bytecode compiler

The bytecode compiler lowers the AST into a `BytecodeChunk`.

A chunk is the executable bytecode unit used by the VM. Top-level source files compile to top-level chunks. Function literals compile to nested function chunks.

The compiler is responsible for preserving Stult semantics while lowering higher-level AST constructs into explicit VM instructions.

When compiling function literals, ordinary parameters are lowered into bytecode parameter metadata. This metadata includes the parameter name, binding immutability and whether the parameter is optional. The variadic parameter is stored separately because it binds an array rather than void when omitted.

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

Conditionals, conditional expressions, match expressions, loops, break and early return are lowered to jumps and returns.

The compiler emits placeholder jump operands and patches them once target instruction indices are known.

Logical `&`, logical `|`, conditional expressions and match expressions are compiled with control flow rather than eager evaluation.

For a conditional expression, the compiler emits the condition, jumps to the false branch when needed, compiles exactly one selected branch at runtime and leaves that branch value on the stack.

For a match expression, the compiler evaluates the subject once, stores it in a compiler-generated local slot, compares it with each explicit arm pattern in source order, and jumps to the selected result expression. If no explicit arm matches, the compiler emits the default expression when one exists, or void when no default exists.

Early return from functions compiles to a return path that exits the current function chunk.

A bare `^` inside loops compiles to a loop break.

Try-catch statements compile to explicit try-handler setup and teardown instructions. The compiler emits `TRY_START` before the try body, `TRY_END` when the try body completes normally, and a jump around the catch body. Break and early-return code paths emit any needed `TRY_END` instructions so control flow does not leave stale handlers behind.

### Dynamic loops

Stult uses one loop syntax for condition loops, collection loops and function loops:

```text
((condition_or_collection_or_function)) { ... }
```

The bytecode compiler emits runtime-dispatched loop code.

At runtime, the VM decides whether the loop expression is a collection or a user-defined function. If it is a collection, the loop uses iterator opcodes for collection iteration. If it is a user-defined function, the loop uses the same iterator machinery to call the function repeatedly until it returns void. Otherwise, it behaves as a condition loop.

Collection loops support up to four parameters:

```text
value
key
collection
position
```

Function loops support up to two parameters:

```text
value
position
```

The compiler allocates locals for loop parameters and resets loop-scope locals on each iteration.

The `ITERATOR_INIT` instruction carries the number of loop parameters as its operand. The VM uses that count when validating whether the runtime loop source is a collection or a function.

### Try-catch

Try-catch is a statement-level control-flow construct. The compiler lowers it with an active error-handler range rather than by wrapping whole chunks.

The emitted shape is:

```text
TRY_START catch_start
try body
TRY_END
JUMP after_catch
catch_start:
  optional error-message binding
  catch body
after_catch:
```

The catch parameter is optional. When present and not `_`, the compiler creates a local binding for the runtime error message string at the beginning of the catch block.

The compiler tracks try depth so that break and early return can close active handlers before leaving the protected region. Those control-flow paths must not be treated as catchable errors.

### Range-backed loop optimisation

A loop source can be written as a range-only array literal:

```text
(({1..1000})) { ... }
```

The ordinary meaning of `{1..1000}` is still an array literal. However, when that literal is used directly as the loop source, contains exactly one range segment and the loop body has fewer than three loop parameters, the compiler can avoid materialising the array and emit direct range-iterator setup instead.

The optimisation is not valid when the loop body has three or four parameters, because the third collection-loop parameter exposes the collection object itself. In that case, the compiler must preserve ordinary array semantics by emitting the normal range-building and array-building instructions before iterator initialisation.

Range bounds, steps and current values must use Stult's exact integer number representation rather than Go host-sized integers. This keeps the optimised path aligned with ordinary range materialisation for values larger than `int64`.

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
try-handler stack
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

The VM caches local indices by chunk and reset depth to reduce repeated lookup work.

### Functions and closures

Function literals compile to bytecode function metadata and nested chunks.

At runtime, user-defined functions remain ordinary `ValueFunction` values in both runtime modes. A `Function` value may be backed by either interpreter data or bytecode data. Interpreter-backed functions store AST parameters, body, return expressions and a defining environment. Bytecode-backed functions store a `BytecodeFunction` pointer and captured VM upvalue cells.

Bytecode-backed function values need:

```text
function chunk
parameters
optional variadic parameter
captured upvalues
```

The VM saves and restores execution state when running bytecode functions. This avoids constructing a separate VM for every function call while still isolating chunk, stack, locals, upvalues and iterator state for the call.

Function-call argument binding uses each function's arity metadata. Required ordinary parameters must receive arguments. Optional ordinary parameters receive the supplied argument when present, or void when omitted. Variadic parameters receive an array containing the remaining arguments, or an empty array when no remaining arguments exist.

The helper that checks whether a user-defined function can accept a given argument count must work for both interpreter-backed functions and bytecode-backed functions. Function loops rely on that helper when deciding whether to call a generator with the zero-based index or with no arguments.

### Collection iteration

The VM has explicit iterator state for collection loops.

Array, string and map iteration share the same loop machinery but produce different keys and values:

```text
array   key is numeric index, value is element
string  key is numeric index, value is one-character string
map     key is string key, value is map entry value
```

Map iteration is deterministic only to the degree enforced by the runtime helper used to find the next key. Any change to map iteration order must be checked against examples and interpreter parity.

The iterator stack can also hold direct range-iterator state for the single-range loop optimisation. This state stores the current value, end value, step and inclusivity using exact Stult number data, so direct range loops do not impose an `int64` limit. Array iteration uses the array method layer and Stult-number indices so overflow-backed arrays and direct range loops share the same arbitrary-size index model.

### Function loops

Function loops reuse the VM iterator stack. A function loop's iterator source is a user-defined `ValueFunction` rather than a collection.

For each iteration, the VM computes the next zero-based position and decides how to call the generator function:

```text
can accept one argument  call generator(position)
can accept zero args     call generator()
otherwise                runtime error
```

If the generator returns void, the iterator jumps to the loop exit target. Otherwise, the returned value becomes the loop value for that iteration and the position becomes the optional second loop-body parameter.

Function loops allow zero, one or two loop parameters. They deliberately do not provide the collection parameter used by collection loops, because the source is active code rather than passive collection data. If a bytecode instruction tries to store a collection parameter for a function-loop iterator, the VM reports a runtime error.

Builtin functions are not currently function-loop sources. They remain `ValueBuiltinFunction` values rather than `ValueFunction` values, and the loop-source dispatch accepts only user-defined functions.

### Try handlers

The VM keeps a stack of active try handlers. Each handler records the catch instruction pointer and enough VM state to restore execution when a runtime error is caught. This includes the operand stack size and iterator-stack depth at the point where the try block began.

When an instruction returns a runtime error and at least one try handler is active, the VM handles the error by:

1. popping the nearest handler,
2. restoring the protected stack and iterator state,
3. pushing the error message as a string for the catch block and
4. jumping to the catch instruction pointer.

When no try handler is active, runtime errors keep the ordinary uncaught-error behaviour.

`TRY_END` removes the nearest handler when a try body completes normally or when compiler-emitted control flow deliberately leaves the protected region.

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

Runtime error text is part of user-visible behavior, so changes should be made carefully. Try-catch exposes the caught runtime error message as a Stult string, so changes to error wording can also affect programs that inspect caught errors.

## Tree-walk interpreter

The interpreter evaluates the AST directly.

It uses chained `Environment` values for lexical scope.

The interpreter remains useful because it is simpler to reason about than the bytecode compiler plus VM. When bytecode behavior differs from interpreter behavior, the interpreter should usually be treated as the reference unless the interpreter is known to be wrong.

Conditional expressions are evaluated directly by the interpreter. The interpreter evaluates the condition first, checks that it is a boolean and then evaluates only the selected branch expression.

Match expressions are also evaluated directly by the interpreter. The interpreter evaluates the subject once, compares it with explicit scalar-literal patterns using the same equality semantics as `=`, evaluates only the selected result expression, and falls back to the default arm or void when no explicit arm matches.

Try-catch statements are evaluated directly by the interpreter. The interpreter runs the try body in a child frame. If it receives an ordinary runtime error, it runs the catch body in a fresh child frame and optionally binds the error message string to the catch parameter. Break and early-return control values pass through unchanged.

Function loops are evaluated directly by the interpreter. The interpreter evaluates the loop source once to decide whether the loop is boolean, collection-based or function-based. For a function loop, it repeatedly calls the user-defined generator with the zero-based index when the function can accept one argument, or with no arguments when it can accept zero. A void return stops the loop normally.

The interpreter also has the same direct range-loop fast path as the bytecode runtime: a single range-only array literal used directly as a loop source may be streamed when the body has fewer than three parameters. If the body requests the collection parameter, the interpreter materialises the array to preserve observable behaviour.

The interpreter path is selected with:

```text
stult run --interpreter ...
```

It is also used by source/interpreter bundles.

Interpreter function calls use the same arity model as the bytecode VM. Required parameters must be supplied, optional parameters are filled with void when omitted, and variadic parameters collect remaining arguments into an array.

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

`ValueFunction` is the user-defined function kind for both interpreter-backed and bytecode-backed functions. It is distinct from `ValueBuiltinFunction`, which wraps Go standard-library functions. This distinction matters for function loops, which currently accept user-defined functions as generator sources but not builtin functions.

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

Bindings may also carry optional runtime type contracts. `<.>` records the runtime value kind assigned when the binding is created and rejects later assignments of a different kind. `<*>` explicitly requests the default dynamic behaviour. Named contracts such as `STD.TYPE.NUMBER`, `STD.TYPE.ARRAY<STD.TYPE.STRING>` and `STD.TYPE.MAP<STD.TYPE.BOOL>` are parsed as contract descriptors rather than ordinary expressions. They are enforced when the binding or map entry is created and on later reassignment.

Collection contracts also attach to the collection value itself. An array value created or assigned through `STD.TYPE.ARRAY<contract>` stores the element contract on the `Array`, and a map value created or assigned through `STD.TYPE.MAP<contract>` stores the value contract on the `Map`. Mutation paths such as array index assignment, map index assignment and helper functions that call `Array.Set` or `Map.Set` must check those collection-level contracts. This prevents aliases from bypassing the contract.

Contract markers are declaration syntax: they may appear only when a binding or map entry is created, not on later reassignment. Because bytecode execution needs to distinguish explicit `<*>` from an uncontracted assignment, the compiler emits separate store opcodes for explicit-any contracts even though the resulting binding contract is the same as the default any-kind contract. Named contracts are stored directly on bytecode instructions so the VM can create and validate the same runtime contract as the interpreter.

### Number values

Stult has one user-visible number type.

Internally, number values may be represented as small integers, arbitrary-size integers or scaled decimals. This is an implementation detail: Stult programs still see a single number type.

For scaled decimals, Stult stores a signed whole-number coefficient plus a decimal scale. The value is the coefficient divided by ten to the power of that scale. Apostrophe digit separators are removed before number values are parsed, so `1'000` enters the runtime as the same number value as `1000`. Percentage-suffixed number literals are normalised by increasing the decimal scale by two, so `50%` enters the runtime as the same number value as `0.5`. The suffix is handled while reading and parsing the number literal; it does not introduce a separate percentage operator.

Whole-number values are theoretically unbounded, subject to available memory and processing time. Digits after the decimal point are bounded and rounded to a maximum number of decimal places.

The maximum decimal-place limit controls the number of digits after the decimal point, not the total number of digits in the number. Ordinary display uses fewer decimal places by default, but standard-library formatting helpers can request more, up to `STD["TYPE"]["NUMBER"]["DECIMAL_PLACES_MAX"]`.

Exact arithmetic and comparison should use the integer or scaled-decimal representation directly where possible. This includes ordinary arithmetic, exact integer operations, fixed decimal formatting and exact serialisation.

Approximate mathematical operations, such as square roots, non-integer powers, interpolation constants and trigonometric functions, may use high-precision floating-point working values internally.

### Array and string values

Arrays are dense ordered collections. The public semantics remain simple: an index below the current length replaces or reads an existing element, an index equal to the current length appends, and an index greater than the current length is out of bounds.

The reference implementation stores arrays in two segments:

```text
Ordinary    fast slice-backed storage for indices below 1 << 24
Overflow    map-backed continuation chunks for indices at or above 1 << 24
Length      Stult number length for arrays that may exceed ordinary storage
```

Overflow chunks are keyed by decimal string chunk indices. This keeps ordinary arrays fast while allowing extremely large dense arrays to continue beyond the host slice index size. Overflow storage is still dense: it is a continuation of the ordinary segment, not a sparse map of arbitrary indices.

Array helper methods such as `Len`, `Get`, `Set`, `Append`, `Clear` and `ForEach` are the source of truth for whole-array behaviour. Code that intentionally works only with the ordinary segment may use the ordinary slice directly, but formatting, cloning, freezing, iteration, standard-library traversal and serialisation should use the method layer so overflow elements are preserved.

`STD["TYPE"]["ARRAY"]["SORT"]` also uses the array method layer to copy elements before sorting. The sort returns a new mutable array and is stable; values that do not have a natural same-kind ordering, such as arrays, maps and functions, compare equal within their kind so their relative order is preserved.

Strings remain contiguous Unicode code-point slices in the reference implementation. They are mutable unless frozen, but unlike arrays they are still bounded by the host representation and by operations that convert to or from Go strings. Making strings chunked or streaming would require a larger representation and standard-library refactor.

### Immutability versus freezing

Binding immutability and collection freezing are separate concepts.

Binding immutability controls whether a name or map entry can be rebound. Collection freezing controls whether the contents of an existing array, map or string can be internally modified.

Arrays, maps and strings carry collection-level frozen flags. In the Go implementation those collection fields are named `IsFrozen` on `Array`, `Map` and `String`, while binding and map-entry rebinding metadata continues to use `IsImmutable`. This keeps the implementation distinction aligned with the language distinction: immutable bindings cannot be rebound, and frozen collections cannot be internally mutated. Frozen collection literals and `STD["TYPE"]["COLLECTION"]["FREEZE"]` without the optional deep flag use the shared shallow-freeze helper for arrays, maps and strings. `FREEZE(collection, +)` uses the deep-freeze helper to recursively freeze nested collection graphs. `FREEZE` modifies an existing collection in place and returns the same collection value; frozen literals create a new literal value and then freeze that value before it is exposed to user code. `STD["TYPE"]["COLLECTION"]["IS_FROZEN"]` reports whether a collection value currently has its collection-level frozen flag set.

`STD["TYPE"]["COLLECTION"]["CLONE"]` shallow-clones by default, returning a new mutable top-level array, map or string while reusing nested values. `CLONE(value, +)` deeply clones collection graphs. The deep clone operation is cycle-safe and alias-preserving: if the original graph contains multiple references to the same nested collection, the cloned graph contains multiple references to the same cloned nested collection. Cycles in the original graph become cycles in the cloned graph. Cloned arrays, maps and strings are mutable even when the original collections are frozen. Map-entry binding mutability is preserved. Numbers are cloned defensively, while booleans, void, functions and builtin functions are reused.

`STD["TYPE"]["MAP"]["MERGE"]` performs a shallow two-map merge by default, returning a new mutable map. It accepts frozen input maps, lets the right map override the left map, reuses entry values and copies the winning entry binding metadata so map-entry mutability is preserved. `MERGE(left, right, +)` recursively merges map values when both sides of the same key are maps. Arrays, strings and all non-map values are replaced rather than merged. Recursive merge results are new mutable maps; other values are reused.

Mutation helpers for arrays, maps and strings must check collection-level freezing before changing collection contents. User-facing errors for frozen collection mutation should describe the collection as frozen, rather than describing the binding as immutable.

## Standard library

The standard library is exposed through the immutable global binding `STD`.

`STD` is a map of maps:

```text
STD["ERROR"]
STD["DATA"]
STD["FILE"]
STD["IO"]
STD["MATH"]
STD["SYSTEM"]
STD["TIME"]
STD["TYPE"]
```

Builtins are Go functions wrapped as Stult callable values.

Standard-library functions should return Stult values and errors, not print internal Go details.

Because both runtime modes use the same standard library, changes to builtins usually affect both bytecode and interpreter behavior.

## Manifests

A manifest lists source files to run in order.

Manifest loading normalizes STULTON and JSON manifests into one Go representation with `RunFiles`. STULTON manifests use the uppercase `RUN` field, while JSON manifests use the lowercase `run` field.

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

Try-catch catches runtime execution errors only. Lexing, parsing, bytecode compilation and command-line setup errors happen before program execution and are not catchable by Stult code.

Do not replace user-facing Stult error messages with raw Go panics.

## Tests

The Go test suite includes public example-test programs and ordinary public examples.

Regression-oriented example tests live under:

```text
examples/tests/
```

Ordinary public examples live under:

```text
examples/
examples/projects/
```

Runnable `.stult` files and manifest projects are run through both runtime modes:

```text
interpreter
bytecode VM
```

A test passes only when both runtime modes complete successfully and produce the same stdout and stderr.

If both runtime modes fail with matching errors, the test still fails. Matching failure is not success for these example programs.

The purpose of these tests is not just coverage. The files under `examples/tests/` are public regression fixtures for language behaviour, parser behaviour, standard-library behaviour and interpreter/bytecode parity. For example, the function-loop fixture checks indexed generators, zero-argument generators, optional generator parameters, ignored loop-body parameters and ordinary break behaviour. The range-loop optimisation fixture checks direct range streaming, very large integer bounds, descending and stepped ranges, and the fallback path where the loop body can observe the materialised collection.

The ordinary examples are also documentation fixtures, so the test suite runs the non-interactive examples that users are likely to copy from the README and examples documentation. Examples that intentionally animate or sleep for a long time are kept in an explicit skip list with a reason.

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

Released-version notes and changelog entries are maintained in [`versions.md`](versions.md).

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

Some syntax can deliberately reuse existing AST and runtime paths. Dot access is one example: `object.key` is lowered to an index expression with a string key, so ordinary indexing, assignment and compound-assignment behaviour should remain the source of truth. Leading-dot field access also reuses index-expression assignment behaviour, but its receiver is runtime-resolved from the nearest captured map, so interpreter and bytecode map-capture handling must stay aligned.

Other syntax needs its own AST shape even when it looks compact. Conditional expressions are one example: `(condition):(when_true|when_false)` must remain lazy, so it should be handled as control flow in both the interpreter and bytecode compiler rather than as a call-like expression.

Match expressions are another example: `(subject):{ ... }` must evaluate the subject once, evaluate only the selected result expression, and treat `_` as a fallback after explicit patterns fail. It should therefore be handled as its own AST and compiler path rather than lowered to a map or function call.

Try-catch has both syntax and runtime implications. Parser changes should preserve the same-line `?{` opener and the same-line `}|{` separator, while allowing horizontal spacing in both places. Runtime changes should keep break and early return separate from catchable errors. In bytecode, any control-flow path that leaves a protected try region must also emit matching `TRY_END` instructions.

Function loops are different: they deliberately reuse the existing `LoopStatement` AST shape. The parser does not need a new syntax node because `((source)) { ... }` is already parsed as a loop. The interpreter and VM decide at runtime whether the source is a boolean, a collection or a user-defined function.

The range-loop optimisation is also deliberately not new syntax. It recognises a specific existing AST shape in loop-source position and must never change observable behaviour. In particular, the collection parameter forces materialisation, and exact integer range semantics must not be narrowed to host integer widths.

When changing function parameter or call-argument syntax, keep parser validation, interpreter call binding, bytecode parameter metadata, VM call binding, bytecode disassembly and bundled bytecode encoding aligned. Function-loop generator calls also depend on arity metadata, so optional and variadic parameter changes must be tested against function loops too. Spread arguments should also be tested through both ordinary user-defined calls and builtin calls.

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
