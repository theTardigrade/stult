# Stult

Stult is a small interpreted programming language and interpreter written in Go.

It is designed as a terse but readable scripting language with:

- uppercase immutable bindings and lowercase mutable bindings,
- explicit outer-scope writes using `@`,
- one high-precision 2048-bit number type,
- arrays, maps, strings, functions, conditionals, loops and ranges,
- concise literals for booleans, arrays, maps and void,
- manifest-based projects, direct source-string evaluation and bundled executables *and*
- a map-shaped standard library available through `STD`.

Idiomatic Stult code should be light on syntax without being deliberately cryptic.

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
  - [Evaluating source strings](#evaluating-source-strings)
- [Manifests](#manifests)
- [Bundled executables](#bundled-executables)
- [Language overview](#language-overview)
  - [Comments](#comments)
  - [Values](#values)
  - [Numbers](#numbers)
  - [Bindings](#bindings)
    - [Outer bindings](#outer-bindings)
  - [Operators](#operators)
  - [Compound assignment](#compound-assignment)
  - [Collections](#collections)
  - [Conditionals](#conditionals)
  - [Creating a local scope](#creating-a-local-scope)
  - [Loops](#loops)
    - [Infinite loops](#infinite-loops)
    - [Break and early return](#break-and-early-return)
  - [Functions](#functions)
- [Standard library](#standard-library)
- [STULTON](#stulton)
- [Repository layout](#repository-layout)

## Status

Stult is still evolving.

That is to say, the programming language, interpreter and standard library are all **works in progress**.

The language has not yet reached version 1.0.0, so its syntax, standard-library names and runtime behavior may change before the first stable release.

Even so, Stult can certainly be used, in its current state, to solve genuine problems and perform real-world tasks. I encourage you to do this.

## Why use Stult?

Stult is designed for small scripts that can grow into distributable command-line tools.

You can run a single file, execute a manifest-based project or bundle a project into
one executable that contains the Stult runtime and the project source files.

Stult may be useful when you want:

- a tiny interpreted language distributed as a single Go binary,
- concise syntax for local scripts and data-heavy automation,
- manifest-based multi-file projects,
- standalone bundled executables that include their source files *and*
- a language implementation small enough to read and quickly understand.

### No knowledge of Go required

Stult is implemented in Go, but Stult programs are written in Stult.

You do not need to know Go to write, run or bundle Stult code.

## Quick start

Run a single source file:

```bash
go run ./src examples/area_of_circle.stult
```

Run a manifest-based project:

```bash
go run ./src examples/bool
```

Evaluate a source string directly:

```bash
go run ./src -e 'STD["IO"]["PRINT"]("hello")'
```

Run from the current directory by discovering a manifest upward from `.`:

```bash
go run ./src
```

Build a local Stult binary:

```bash
go build -o stult ./src
```

Then run:

```bash
./stult examples/area_of_circle.stult
```

Or on Windows:

```bash
stult.exe examples/area_of_circle.stult
```

## Examples

See [docs/examples.md](docs/examples.md) for a guided list of example Stult programs that demonstrate the use of collections, control flow, data formats, manifests and the standard library.

## Development commands

The development helper lives at `util/build.go`.

Build all release binaries into `dist/`:

```bash
go run ./util/build.go dist
```

Build a local Stult binary:

```bash
go run ./util/build.go build
```

Run Stult through Go:

```bash
go run ./util/build.go run examples/area_of_circle.stult
```

Run tests:

```bash
go test ./...
```

Format source:

```bash
go fmt ./...
```

Clean generated build outputs:

```bash
go run ./util/build.go clean
```

## Running programs

The `stult` command accepts:

```text
stult
stult <file.stult>
stult <directory>
stult <manifest.stulton>
stult <manifest.json>
stult -e <code-string>
stult --eval <code-string>
stult build [project-directory] -o <output-executable>
```

With no arguments, `stult` searches upward from the current directory for a manifest.

With a `.stult` file argument, it runs that file.

With a directory argument, it looks for a manifest in that directory.

With a manifest argument, it runs the files listed by that manifest.

With `-e` or `--eval`, it runs the provided source string directly.

### Evaluating source strings

Stult can run source code passed directly on the command line with `-e` or `--eval`.

This is useful for quick experiments, shell scripts and short one-off commands.

For example:

```bash
stult -e 'X : 10,STD["IO"]["PRINT"](X*20)'
```

On Windows PowerShell, quotes inside the evaluated source may need to be escaped:

```powershell
stult -e 'X : 10,STD[\"IO\"][\"PRINT\"](X*20)'
```

The evaluated source runs in a fresh interpreter with the standard library available as `STD`.

## Manifests

A project may define either one of the two following files:

```text
manifest.stulton
manifest.json
```

A STULTON manifest uses Stult-style data:

```stulton
{
    "RUN": {
        "dir/helpers.stult"
		"dir/main.stult"
	}
}
```

A JSON manifest uses lowercase JSON-style fields:

```json
{
    "run": [
        "dir/helpers.stult",
        "dir/main.stult"
    ]
}
```

Manifest run files are executed in order in the same interpreter, so earlier files can define bindings used by later files.

## Bundled executables

Stult can build a project into a standalone executable with the project files embedded.

```bash
go run ./src build examples/bool -o bool-app
```

Or, after building the local binary:

```bash
./stult build examples/bool -o bool-app
```

The project directory must contain either `manifest.stulton` or `manifest.json`.

When the generated executable starts, it checks for an embedded bundle and runs the bundled manifest automatically.

## Language overview

### Comments

Line comments use `#`.

```stult
# This is a line comment.
```

Bounded comments use `##`.

```stult
## This is a bounded comment follow by some code. ## x : 10
```

Bounded comments can span across multiple lines.

```stult
##
EXAMPLE
##
```

Three or more consecutive `#` symbols are invalid.

### Values

```stult
_          # void
\/         # true
/\         # false
123.45     # number
-678.90    # negative number
"hello"    # string

{}         # empty array
{:}        # empty map

{1, 2, 3}  # array
{"x": 1, "y": 2}   # map
```

Map keys preserve case. Uppercase map keys create immutable map entries.

### Numbers

Stult has one numeric type: **number**.

There is no separate integer type.

Whole numbers, decimal numbers and negative numbers are all represented as numbers.

```stult
1
-10
123.45
0.000001
```

Internally, Stult numbers use high-precision floating-point values with 2048 bits of precision.

This means that a Stult number can safely store very large and very small integer values without losing integer precision in the range that ordinary programming languages often struggle with.

For example, values far beyond JavaScript’s usual safe integer range can still be represented exactly as integers in Stult.

Decimals are also supported, but they are still floating-point values, not a separate type.

```stult
1 / 3
```

By default, numbers are printed with up to 20 digits after the decimal point.

This is only the display format. Internally, numbers are stored with much higher precision, since Stult currently uses 2048-bit floating-point numbers. That means a number may store more precision than is shown when printed.

```stult
# 0.33333333333333333333
STD["IO"]["PRINT"](1/3)
```

### Bindings

Assignments use `:`.

```stult
name : "example"  # mutable
NAME : "example"  # immutable
```

An identifier is immutable when it contains at least one uppercase letter and no lowercase letters.

```stult
NAME      # immutable
MAX_SIZE  # immutable
X1        # immutable

name      # mutable
Name      # mutable
_name     # mutable
_         # void/discard-style name
```

#### Outer bindings

`@name` reads or writes the nearest outer binding.

Reads can usually omit `@` because ordinary reads search outward anyway:

```stult
value : 10

(\/) {
	STD["IO"]["PRINT"](value)
}
```

Writes to an outer binding must use `@`:

```stult
count : 0

((count < 3)) {
	@count :+ 1
}
```

Without `@`, an assignment writes in the current scope.

### Operators

```stult
a + b   # addition
a - b   # subtraction
a * b   # multiplication
a / b   # division

a = b   # equal
a ! b   # not equal

a < b   # less than
a <= b  # less than or equal
a > b   # greater than
a >= b  # greater than or equal

a & b   # boolean and
a | b   # boolean or
!a      # boolean not
```

Stult uses `=` for equality, not `==`.

Stult uses `!` for inequality, not `!=`.

### Compound assignment

The following syntax is available for performing arithmetic updates:

```stult
count : 10

count :+ 1
count :- 1
count :* 2
count :/ 5
```

The above Stult code is roughly equivalent to the following C code:

```c
double count = 10;

count += 1;
count -= 1;
count *= 2;
count /= 5;
```

### Collections

Arrays, maps and strings can be indexed. Assignment is done via indexing:

```stult
items : {"a", "b"}
items[2] : "c"

record : {"name": "example"}
record["city"] : "London"

text : "cat"
text[0] : "b"
```

Assigning to an array or string at index equal to its current size appends.

```stult
items : {"a", "b"}
items[2] : "c"
```

Ranges are available in array literals.

```stult
{1..5}       # 1, 2, 3, 4, 5
{1...5}      # 1, 2, 3, 4
{10..16[2]}  # 10, 12, 14, 16
```

Ranges can descend too:

```stult
{5..1}
{10...1[3]}
```

### Conditionals

```stult
(age >= 100) {
	STD["IO"]["PRINT"]("centenarian")
},{
	STD["IO"]["PRINT"]("not a centenarian")
}
```

Else-if blocks use `},(`.

```stult
(x < 0) {
	STD["IO"]["PRINT"]("negative")
},(x = 0) {
	STD["IO"]["PRINT"]("zero")
},{
	STD["IO"]["PRINT"]("positive")
}
```

### Creating a local scope

A conditional with a true literal can be used as an idiomatic way to create a local scope:

```stult
x : 10

(\/) {
	x : 20

	STD["IO"]["PRINT"](x)
	STD["IO"]["PRINT"](@x)
}
```

### Loops

While-style loops use touching double parentheses:

```stult
i : 0

((i < 3)) {
	STD["IO"]["PRINT"](i)
	@i :+ 1
}
```

A loop may have an after-loop block that runs when the condition no longer holds true:

```stult
i : 0

((i < 3)) {
	STD["IO"]["PRINT"](i)
	@i :+ 1
},{
	STD["IO"]["PRINT"]("done")
}
```

Collection loops work with arrays, maps and strings:

```stult
items : {"a", "b", "c"}

((items)) { (value, index)
	STD["IO"]["PRINT"](index, ": ", value)
}
```

Collection loop parameters are:

```text
value
key/index
collection
position
```

You may provide from zero to four parameters:

```stult
((items)) {
	STD["IO"]["PRINT"]("one item")
}

((items)) { (value)
	STD["IO"]["PRINT"](value)
}

((items)) { (value, key, collection, position)
	STD["IO"]["PRINT"](position, ": ", key, " -> ", value)
}
```

#### Infinite loops

An infinite loop can be written by giving a true literal as the condition:

```stult
((\/)) {
	STD["IO"]["PRINT"]("this runs forever")
}
```

That is the idiomatic way to write an eternally iterating loop in Stult.

#### Break and early return

Bare `^` breaks out of the nearest loop:

```stult
i : 0

((\/)) {
	(i = 3) {
		^
	}

	@i :+ 1
}
```

`^(value)` returns early from the nearest function:

```stult
first_positive : { (values)
	((values)) { (value)
		(value > 0) {
			^(@value)
		}
	}

	(_)
}
```

### Functions

Function literals use braces.

Parameters come first. The final parenthesized expression is the return value.

```stult
add : { (a, b)
	(a + b)
}

total : add(2, 3)

STD["IO"]["PRINT"](total)
```

Function calls require touching parentheses:

```stult
add(1, 2)
```

Functions always return exactly one value, even if that value is merely `_`.

## Standard library

The standard library is available as the immutable binding `STD`.

It is a map containing functions and other maps.

```stult
STD["IO"]
STD["FILE"]
STD["TIME"]
STD["MATH"]
STD["TYPE"]
STD["DATA"]
```

Here is some example code using some functions from the standard library:

```stult
PRINT : STD["IO"]["PRINT"]
SIZE : STD["TYPE"]["COLLECTION"]["SIZE"]
MATH : STD["MATH"]

PRINT("size: ", SIZE({"a", "b", "c"}))
PRINT("square: ", MATH["SQUARE"](9))
```

For the full standard library reference, please see [docs/std.md](docs/std.md).

## STULTON

STULTON is Stult’s native data notation. It is used for manifests, config files and simply storing Stult values.

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

  interpreter*.go             AST evaluation
  environment.go              lexical scopes and bindings
  control.go                  internal break/return control flow

  value*.go                   runtime value types and formatting

  std*.go                     standard library maps and functions

  bundle*.go                  embedded bundle loading and building

  manifest.go                 manifest loading
  main.go                     CLI entrypoint

examples/                     example Stult programs

docs/                         reference documentation

util/
  build.go                    development/release helper
```