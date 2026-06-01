# Stult

Stult is a small programming language and runtime written in Go.

It is designed as a terse but readable scripting language with:

- uppercase immutable bindings and lowercase mutable bindings,
- explicit outer-scope writes using `@`,
- one high-precision 2048-bit number type,
- arrays, maps, strings, functions, conditionals, loops and ranges,
- concise literals for booleans, arrays, maps and void,
- manifest-based projects, direct source-string evaluation and bundled executables *and*
- a map-shaped standard library available through `STD`.

Idiomatic Stult code should be light on syntax, but never deliberately cryptic.

Source files use the `.stult` extension.

STULTON, Stult’s native data notation, uses the `.stulton` extension.

## Contents

- [Status](#status)
- [Why use Stult?](#why-use-stult)
  - [No knowledge of Go required](#no-knowledge-of-go-required)
- [Quick start](#quick-start)
- [Examples](#examples)
- [Development commands](#development-commands)
- [Running programs](#running-programs)
  - [Runtime modes](#runtime-modes)
  - [Evaluating source strings](#evaluating-source-strings)
  - [Dumping bytecode](#dumping-bytecode)
- [Manifests](#manifests)
- [Bundled executables](#bundled-executables)
- [Language overview](#language-overview)
  - [Comments](#comments)
  - [Values](#values)
  - [Numbers](#numbers)
  - [Bindings](#bindings)
    - [Outer bindings](#outer-bindings)
    - [Boolean bindings](#boolean-bindings)
  - [Operators](#operators)
  - [Compound assignment](#compound-assignment)
  - [Collections](#collections)
  - [Conditionals](#conditionals)
    - [Creating a local scope](#creating-a-local-scope)
  - [Loops](#loops)
    - [Iterating over collections](#iterating-over-collections)
    - [Infinite loops](#infinite-loops)
    - [Break](#break)
  - [Functions](#functions)
    - [Early return](#early-return)
    - [Variadic function parameters](#variadic-function-parameters)
    - [Immediately invoked function expressions](#immediately-invoked-function-expressions)
  - [Commas and newlines](#commas-and-newlines)
- [Standard library](#standard-library)
- [STULTON](#stulton)
- [Repository layout](#repository-layout)
- [Architecture](#architecture)

## Status

Stult is still evolving.

That is to say, the programming language, runtime and standard library are all **works in progress**.

The language has not yet reached version 1.0.0, so its syntax, standard-library names and runtime behavior may change before the first stable release.

Even so, Stult can be used, in its current state, to solve genuine problems and perform real-world tasks. I encourage you to do this.

## Why use Stult?

Stult is designed for small scripts that can grow into distributable command-line tools.

You can run a single file, execute a manifest-based project or bundle a project into one executable that contains the Stult runtime and the program.

Stult may be useful when you want:

- a tiny scripting language distributed as a single Go binary,
- concise syntax for writing local scripts that process data and automate tasks,
- manifest-based multi-file projects,
- standalone bundled executables *and*
- a language implementation small enough to read and quickly understand.

### No knowledge of Go required

Stult is implemented in Go, but Stult programs are written in Stult.

You do *not* need to know Go to write, run or bundle Stult code.

When you download a distributed Stult binary, you do not need Go installed on your computer at all.

## Quick start

If you have not already downloaded a Stult binary, you can build a local one:

```bash
go run ./util/build_helper.go local
```

This creates a `stult` executable in the repository root.

Run a single source file:

```bash
./stult run examples/calculate_circle_area_from_map.stult
```

On Windows, use `.\stult.exe` instead:

```powershell
.\stult.exe run examples\calculate_circle_area_from_map.stult
```

Run a manifest-based project:

```bash
./stult run examples/bool
```

Evaluate a source string directly:

```bash
./stult run -e 'STD["IO"]["PRINT"]("hello")'
```

Run from the current directory by discovering a manifest upward from `.`:

```bash
./stult run
```

Dump bytecode for a source file:

```bash
./stult dump --bytecode examples/calculate_circle_area_from_map.stult
```

## Examples

See [docs/examples.md](docs/examples.md) for a guided list of example Stult programs that demonstrate the use of collections, control flow, data formats, manifests and the standard library.

## Development commands

The development helper lives at `util/build_helper.go`.

Show helper usage:

```bash
go run ./util/build_helper.go
```

Build a local Stult executable:

```bash
go run ./util/build_helper.go local
```

Build all release executables into `dist/`:

```bash
go run ./util/build_helper.go dist
```

Clean generated build outputs:

```bash
go run ./util/build_helper.go clean
```

Run tests:

```bash
go test ./...
```

Format Go source:

```bash
gofmt -w src util
```

You can also build the local executable directly with Go:

```bash
go build -o stult ./src
```

## Running programs

The `stult` command uses explicit subcommands:

```text
stult run [--bytecode|--interpreter] [file.stult|directory|manifest] [args...]
stult run [--bytecode|--interpreter] -e|--eval <source-string> [args...]
stult dump [--bytecode] [file.stult|directory|manifest]
stult dump [--bytecode] -e|--eval <source-string>
stult build [--bytecode] [project-directory-or-file.stult] -o <output-executable>
```

`stult` with no subcommand prints usage.

`stult run` with no target searches upward from the current directory for a manifest.

With a `.stult` file target, `stult run` runs that file.

With a directory target, `stult run` looks for a manifest in that directory.

With a manifest target, `stult run` runs the files listed by that manifest.

Program arguments after the target are available to Stult code through `STD["SYSTEM"]["ARGS"]`.

For example:

```bash
stult run examples/csv_to_json_converter.stult input.csv output.json
```

makes this available to Stult code:

```stult
STD["SYSTEM"]["ARGS"] # {"input.csv", "output.json"}
```

### Runtime modes

Bytecode is the default runtime mode:

```bash
stult run examples/calculate_circle_area_from_map.stult
```

This is the same as:

```bash
stult run --bytecode examples/calculate_circle_area_from_map.stult
```

The original tree-walk interpreter remains available as an explicit mode:

```bash
stult run --interpreter examples/calculate_circle_area_from_map.stult
```

The interpreter is useful as a reference implementation, debugging fallback and test oracle.

The bytecode runtime is intended to have best performance.

### Evaluating source strings

Stult can run source code passed directly on the command line with `-e` or `--eval`.

This is useful for quick experiments, shell scripts and short one-off commands.

For example:

```bash
stult run -e 'X : 10,STD["IO"]["PRINT"](X * 20)'
```

Or explicitly through the interpreter:

```bash
stult run --interpreter -e 'STD["IO"]["PRINT"]("hello")'
```

On Windows PowerShell, quotes inside the evaluated source may need to be escaped:

```powershell
.\stult.exe run -e 'X : 10,STD[\"IO\"][\"PRINT\"](X * 20)'
```

The evaluated source runs with the standard library available as `STD`.

### Dumping bytecode

`stult dump` compiles source to bytecode and prints a human-readable disassembly.

```bash
stult dump examples/calculate_circle_area_from_map.stult
```

This is the same as:

```bash
stult dump --bytecode examples/calculate_circle_area_from_map.stult
```

You can also dump bytecode for an evaluated source string:

```bash
stult dump -e 'STD["IO"]["PRINT"]("hello")'
```

`dump` is bytecode-only. There is no interpreter dump mode.

## Manifests

A manifest-based project can list multiple Stult source files.

Files run deterministically in the order specified in the manifest file. This allows one file to define bindings that later files can use.

A project may use either one of the two following files:

```text
manifest.stulton
manifest.json
```

A STULTON manifest uses Stult-style syntax:

```stulton
{
	"RUN": {
		"bindings.stult"
		"helpers.stult"
		"main.stult"
	}
}
```

A JSON manifest uses lowercase JSON-style fields:

```json
{
	"run": [
		"bindings.stult",
		"helpers.stult",
		"main.stult"
	]
}
```

Run a project directory that contains a manifest:

```bash
stult run examples/bool
```

Run a manifest file directly:

```bash
stult run examples/bool/manifest.stulton
```

Run from inside a project directory:

```bash
stult run
```

For more information about manifest files, please see [docs/manifests.md](docs/manifests.md).

## Bundled executables

Stult can bundle a single source file or manifest-based project into a standalone executable.

By default, `stult build` creates a bytecode bundle.

A bytecode bundle embeds:

- the Stult runtime,
- a manifest,
- compiled bytecode *and*
- bytecode metadata needed to map manifest entries to bundled bytecode.

Build a bytecode bundle:

```bash
stult build examples/bool -o bool-app
```

This is the same as:

```bash
stult build --bytecode examples/bool -o bool-app
```

Bytecode bundles do not need the original `.stult` source at runtime.

If you explicitly want a source/interpreter bundle, use `--interpreter`:

```bash
stult build --interpreter examples/bool -o bool-app
```

A source/interpreter bundle embeds:

- the Stult runtime,
- a manifest *and*
- the `.stult` source files needed by that manifest.

In either case, run the generated executable directly:

```bash
./bool-app
```

## Language overview

### Comments

Line comments start with `#`:

```stult
# This is a line comment.
```

Bounded comments use `##` at both ends:

```stult
##
This is a bounded comment.
##
```

Bounded comments can span across multiple lines.

The use of three or more consecutive `#` characters is considered invalid.

### Values

Stult has these main value types:

```text
_
booleans
numbers
strings
arrays
maps
functions
builtin functions
```

The void value is written as `_`.

Booleans use symbolic literals:

```stult
\/  # true
/\  # false
```

Strings use double quotes (*not* single quotes):

```stult
"hello"
```

### Numbers

Stult has one high-precision numeric type.

There are no separate integer and floating-point types.

```stult
1
3.14
-20
```

Numbers are stored internally with high precision and printed with up to 20 fractional digits by default.

### Bindings

Assignment uses `:`.

```stult
NAME : "Stult"
count : 0
```

Identifier case controls mutability.

Names containing uppercase letters and no lowercase letters are immutable:

```stult
PI : 3.14159
```

Names containing lowercase letters are mutable:

```stult
count : 0
count : 1
```

Plain reads search outward through enclosing scopes.

Plain assignment writes to the current scope.

#### Outer bindings

Always use `@` when writing to a mutable binding in an outer scope:

```stult
count : 0

(count < 1) {
	@count :+ 1
}
```

Reads can usually omit `@` because ordinary reads search outward anyway.

Even so, `@name` reads the nearest outer binding, skipping the current scope.

This is useful when an inner scope has a binding with the same name as an outer scope.

#### Boolean bindings

If you prefer word-based boolean names to symbolic boolean literals, you can create immutable bindings for them:

```stult
TRUE : \/
FALSE : /\
```

Then use those bindings elsewhere:

```stult
TRUE : \/
FALSE : /\

PRINT : STD["IO"]["PRINT"]

SHOULD_RUN : TRUE

(SHOULD_RUN) {
	PRINT("running")
}
```

This approach is especially useful in manifest-based projects, where shared bindings can be placed in an earlier file and reused by later files.

### Operators

Arithmetic:

```stult
a : 10
b : 2

a + b   # 12
a - b   # 8
a * b   # 20
a / b   # 5
```

Comparison:

```stult
c : 99
d : 100

c = d   # c equals d: false
c ! d   # c does not equal d: true

c < d   # c is less than d: true
c <= d  # c is less than or equal to d: true
c > d   # c is greater than d: false
c >= d  # c is greater than or equal to d: false
```

Logical operators:

```stult
e : \/  # true
f : /\  # false

e & f   # e and f: false
e | f   # e or f: true
!e      # not e: false
!f      # not f: true
```

`=` means equality.

Binary `!` means inequality, but unary `!` used as a prefix means logical not.

### Compound assignment

Stult supports compound assignment:

```stult
count : 10

count :+ 1  # 11
count :- 1  # 10
count :* 2  # 20
count :/ 5  # 4
```

The above Stult code is roughly equivalent to the following C code:

```c
double count = 10;

count += 1;
count -= 1;
count *= 2;
count /= 5;
```

Compound assignment can also update mutable outer bindings:

```stult
total : 0
VALUE : 5

(total = 0) {
	@total :+ VALUE
}
```

### Collections

Arrays, maps and strings can be indexed. For this reason, we call them collections.

Arrays use `{}`:

```stult
values : {"red", "green", "blue"}
```

Maps use string keys:

```stult
person : {
	"NAME": "John"
	"role": "programmer"
}
```

An empty map is written as:

```stult
{:}
```

But an empty array is written as:

```stult
{}
```

Indexing uses square brackets:

```stult
values : {"red", "green", "blue"}
person : {"NAME": "John", "role": "programmer"}

values[0]
person["NAME"]
```

Arrays can include ranges:

```stult
numbers : {1..5}
```

Ranges may be inclusive or exclusive:

```stult
inclusive : {1..5}
exclusive : {1...5}
```

Ranges may also include a step:

```stult
evens : {2..10[2]}
```

### Conditionals

Conditionals use a parenthesised condition followed by a brace-enclosed block:

```stult
PRINT : STD["IO"]["PRINT"]

score : 95

(score >= 90) {
	PRINT("excellent")
}
```

An alternative block, which runs when the condition is false, follows `},{`:

```stult
PRINT : STD["IO"]["PRINT"]

score : 80

(score >= 90) {
	PRINT("excellent")
},{
	PRINT("keep going")
}
```

Multiple branches can be chained:

```stult
PRINT : STD["IO"]["PRINT"]

score : 10

(score >= 90) {
	PRINT("excellent")
},(score >= 70) {
	PRINT("good")
},(score >= 50) {
	PRINT("keep going")
},(score >= 20) {
	PRINT("bad")
},{
	PRINT("terrible")
}
```

#### Creating a local scope

A conditional with a true condition can also be used as an idiomatic way to create a temporary local scope:

```stult
PRINT : STD["IO"]["PRINT"]

(\/) {
	message : "inside local scope"
	PRINT(message)
}
```

Bindings created inside that block do not leak into the surrounding scope.

### Loops

Loops use double parentheses:

```stult
PRINT : STD["IO"]["PRINT"]

count : 3

((count > 0)) {
	PRINT(count)
	@count :- 1
}
```

Loops may have an after-loop block (which runs once, when the condition no longer holds true):

```stult
PRINT : STD["IO"]["PRINT"]

count : 3

((count > 0)) {
	PRINT(count)
	@count :- 1
},{
	PRINT("done")
}
```

#### Iterating over collections

The same loop syntax can iterate over collections:

```stult
PRINT : STD["IO"]["PRINT"]

values : {5, 30, 45}

((values)) { (value)
	PRINT(value)
}
```

Collection loops can receive up to four parameters:

```stult
PRINT : STD["IO"]["PRINT"]

items : {"hat", "coat", "jacket"}

((items)) { (value, key, collection, position)
	PRINT(position, ": ", key, " -> ", value)
}
```

For arrays and strings, `key` is the numeric index.

For maps, `key` is the string key.

For every type of collection, `position` is the zero-based iteration position.

#### Infinite loops

An infinite loop uses the true literal:

```stult
PRINT : STD["IO"]["PRINT"]

((\/)) {
	PRINT("forever")
}
```

#### Break

A bare `^` breaks the nearest loop:

```stult
count : 0

((\/)) {
	@count :+ 1

	(count = 3) {
		^
	}
}
```

### Functions

Functions are values.

A function literal is a block with a parameter list:

```stult
ADD : { (A, B)
	(A + B)
}
```

The final expression is the return value.

Functions return exactly one value.

Function calls require the callee to touch the opening parenthesis:

```stult
PRINT : STD["IO"]["PRINT"]

SUBTRACT : { (A, B)
	(A - B)
}

PRINT("hello")
PRINT(SUBTRACT(10, 2))
```

This means `PRINT ("hello")` or `PRINT( SUBTRACT (10, 2))` are not a valid function calls.

Functions can be stored in maps and arrays:

```stult
MULTIPLY : { (A, B)
	(A * B)
}

TOOLS : {
	"MULT": MULTIPLY
}

TOOLS["MULT"](2, 3)
```

#### Early return

Inside a function, `^(value)` returns early:

```stult
FIND_FIRST : { (items)
	((items)) { (item)
		(item = "target") {
			^(item)
		}
	}

	(_)
}
```

#### Variadic function parameters

A function can collect remaining arguments into an array using a variadic parameter:

```stult
SUM : { (...numbers)
	total : 0

	((numbers)) { (number)
		@total :+ number
	}

	(total)
}
```

The variadic parameter must be last.

```stult
DESCRIBE : { (label, ...values)
	STD["IO"]["PRINT"](label, ": ", values)

	(_)
}
```

#### Immediately invoked function expressions

Stult supports immediately invoked function expressions, or **IIFEs**, which are useful when a value needs a small temporary scope while it is being calculated.

```stult
STATUS : ({ ()
	done : 7
	total : 10

	(done = total) {
		^("complete")
	}

	("in progress")
})()
```

### Commas and newlines

Stult uses both newlines and commas as separators.

Most examples use newlines:

```stult
PRINT : STD["IO"]["PRINT"]
NAME : "Stult"
COUNT : 3
PRINT(NAME)
```

The same statements can be written with commas:

```stult
PRINT : STD["IO"]["PRINT"], NAME : "Stult", COUNT : 3, PRINT(NAME)
```

Commas can also separate function arguments, function parameters, loop parameters, array elements and map entries:

```stult
VALUES : {1, 2, 3}

ADD : { (left, right)
	(left + right)
}

STD["IO"]["PRINT"]("sum: ", ADD(2, 3))

CONFIG : {"name": "demo", "enabled": \/}
```

Newlines may be used in the same places, which is usually clearer for longer code:

```stult
VALUES : {
	1
	2
	3
}

CONFIG : {
	"name": "demo"
	"enabled": \/
}
```

Trailing commas are allowed in list-like syntax:

```stult
VALUES : {
	1,
	2,
	3,
}
```

Some control-flow syntax also uses commas as part of a tightly written separator. Else branches and after-loop blocks use `},{`, while else-if branches use `},(`.

## Standard library

The standard library is available as the immutable binding `STD`.

It is a map containing other maps that, in turn, contain functions.

```stult
STD["IO"]
STD["SYSTEM"]
STD["FILE"]
STD["PATH"]
STD["TIME"]
STD["MATH"]
STD["TYPE"]
STD["DATA"]
```

Here is some example code using functions from the standard library:

```stult
PRINT : STD["IO"]["PRINT"]
SIZE : STD["TYPE"]["COLLECTION"]["SIZE"]
MATH : STD["MATH"]

PRINT("size: ", SIZE({"a", "b", "c"}))
PRINT("square: ", MATH["SQUARE"](9))
```

Program arguments are available through `STD["SYSTEM"]["ARGS"]`:

```stult
ARGS : STD["SYSTEM"]["ARGS"]
PRINT : STD["IO"]["PRINT"]

PRINT(ARGS)
```

For the full standard-library reference, please see [docs/standard_library.md](docs/standard_library.md).

You can also run [examples/standard_library_overview.stult](examples/standard_library_overview.stult) to print a dynamically produced list of everything that the standard library contains:

```bash
stult run examples/standard_library_overview.stult
```

## STULTON

STULTON is Stult’s native data notation. It is used for manifests, config files and storing Stult values.

```stulton
{
	"NAME": "example"
	"is_active": \/
	"empty_array": {}
	"empty_map": {:}
	"items": {
		"one"
		"two"
		"three"
	}
}
```

Use JSON for external systems.

Use STULTON for native data in Stult or for data shared between Stult programs.

## Repository layout

```text
src/
  lexer.go                    source text to tokens
  token.go                    token definitions

  parser*.go                  tokens to AST
  ast.go                      AST node definitions

  bytecode*.go                bytecode compiler, VM, disassembler and bundle support

  interpreter*.go             tree-walk interpreter
  environment.go              lexical scopes and bindings
  control.go                  internal break/return control flow

  value*.go                   runtime value types and formatting

  std*.go                     standard-library maps and functions

  bundle*.go                  embedded bundle loading and building

  manifest.go                 manifest loading

  main.go                     CLI entrypoint
  main_flags.go               CLI flag parsing and usage text

examples/                     example Stult programs

docs/                         reference documentation

util/
  build_helper.go             development/release build script
```

## Architecture

For a technical overview of the implementation, including the compiler pipeline, bytecode virtual machine, interpreter, manifests, bundling and test strategy, see [docs/architecture.md](docs/architecture.md).

<!--
For the formal language definition, see [docs/specification.md](docs/specification.md).
-->