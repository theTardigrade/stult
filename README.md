# Stult

Stult is a small programming language and runtime written in Go.

It is designed as a terse but readable scripting language with:

- uppercase immutable bindings and lowercase mutable bindings,
- explicit outer-scope writes using `@`,
- one high-precision number type with an unbounded whole-number component and bounded decimal places,
- dense arrays of unbounded length that can grow dynamically,
- maps, strings, functions, conditionals, loops and ranges,
- try-catch blocks, conditional expressions and match expressions,
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
  - [Booleans](#booleans)
  - [Strings](#strings)
  - [Numbers](#numbers)
    - [Percentage literals](#percentage-literals)
  - [Bindings](#bindings)
    - [Outer bindings](#outer-bindings)
    - [Boolean bindings](#boolean-bindings)
  - [Operators](#operators)
  - [Compound assignment](#compound-assignment)
  - [Collections](#collections)
    - [Dot access for maps](#dot-access-for-maps)
    - [Cloning collections](#cloning-collections)
    - [Freezing collections](#freezing-collections)
    - [Ranges](#ranges)
  - [Functions](#functions)
    - [Early return](#early-return)
    - [Variadic function parameters](#variadic-function-parameters)
    - [Optional parameters](#optional-parameters)
    - [Immediately invoked function expressions](#immediately-invoked-function-expressions)
  - [Conditionals](#conditionals)
    - [Creating a local scope](#creating-a-local-scope)
    - [Conditional expressions](#conditional-expressions)
    - [Match expressions](#match-expressions)
  - [Error handling](#error-handling)
    - [Try-catch statements](#try-catch-statements)
  - [Loops](#loops)
    - [Infinite loops](#infinite-loops)
    - [Breaking out of a loop](#breaking-out-of-a-loop)
    - [Iterating over collections](#iterating-over-collections)
    - [Function loops](#function-loops)
  - [Commas and newlines](#commas-and-newlines)
- [Standard library](#standard-library)
- [STULTON](#stulton)
- [Repository layout](#repository-layout)
- [Architecture](#architecture)
- [License](#license)

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
./stult run examples/projects/animated_sine_wave
```

Evaluate a source string directly:

```bash
./stult run -e 'STD.IO.OUTPUT.WRITE_LINE("hello")'
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
stult build [--bytecode|--interpreter] [project-directory-or-file.stult] -o <output-executable>
```

`stult` with no subcommand prints usage.

`stult run` with no target searches upward from the current directory for a manifest.

With a `.stult` file target, `stult run` runs that file.

With a directory target, `stult run` looks for a manifest in that directory.

With a manifest target, `stult run` runs the files listed by that manifest.

Program arguments after the target are available to Stult code through `STD.SYSTEM.ARGS`.

For example:

```bash
stult run examples/csv_to_json_converter.stult input.csv output.json
```

makes this available to Stult code:

```stult
STD.SYSTEM.ARGS # {"input.csv", "output.json"}
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
stult run -e 'X : 10,STD.IO.OUTPUT.WRITE_LINE(X * 20)'
```

Or explicitly through the interpreter:

```bash
stult run --interpreter -e 'STD.IO.OUTPUT.WRITE_LINE("hello")'
```

On Windows PowerShell, quotes inside the evaluated source may need to be escaped:

```powershell
.\stult.exe run -e 'X : 10,STD.IO.OUTPUT.WRITE_LINE(X * 20)'
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
stult dump -e 'STD.IO.OUTPUT.WRITE_LINE("hello")'
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
stult run examples/projects/bool
```

Run a manifest file directly:

```bash
stult run examples/projects/bool/manifest.stulton
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
stult build examples/projects/bool -o bool-app
```

This is the same as:

```bash
stult build --bytecode examples/projects/bool -o bool-app
```

Bytecode bundles do not need the original `.stult` source at runtime.

If you explicitly want a source/interpreter bundle, use `--interpreter`:

```bash
stult build --interpreter examples/projects/bool -o bool-app
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

### Booleans

Booleans use symbolic literals:

```stult
+  # true
-  # false
```

When written alone, `+` and `-` are the boolean literals for true and false. When followed by an expression, they are numeric sign operators, so `+10` is positive ten and `-10` is negative ten.

### Strings

A string is an ordered sequence of characters. It is generally used to store text.

Strings use double quotes (*not* single quotes):

```stult
"hello"
```

There is no separate type for a single character in Stult, so a single character would simply be stored in its own short string:

```stult
"h"
```

### Numbers

Stult has one numeric type.

There are no separate integer and floating-point types.

```stult
1
3.14
-20
```

Stult stores numbers internally as a whole-number component plus a decimal component.

Stult numbers can contain extremely large whole-number values, while any digits after the decimal point are bounded.

More precisely, whole-number values are theoretically unbounded, subject to available memory and processing time, but digits after the decimal point are rounded to a maximum number of decimal places (currently 256).

Although Stult keeps more decimal places internally, numbers are ordinarily displayed with fewer decimal places (currently 32).

The number-formatting helpers in `STD.TYPE.NUMBER` can request more decimal places when needed.

#### Percentage literals

A number literal may end with `%`.

The suffix is part of the literal and divides that literal by one hundred, so `50%` is `0.5` and `99.9%` is `0.999`. It follows that `128 * 50%` is `64`.

The `%` must touch the number. `50%` is a percentage literal, but `50 %` is not.

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

If you prefer word-based boolean names to symbolic boolean literals, you can create immutable bindings for them and use those bindings like so:

```stult
TRUE : +
FALSE : -

WRITE_LINE : STD.IO.OUTPUT.WRITE_LINE

SHOULD_RUN : TRUE

# the code below contains a conditional statement,
# so the WRITE_LINE statement only runs if SHOULD_RUN is true
(SHOULD_RUN) {
	WRITE_LINE("running")
}
```

This approach is especially useful in manifest-based projects, where shared bindings can be placed in an earlier file and reused by later files.

The standard library also provides equivalent boolean bindings, which can be used like this:

```stult
TRUE : STD.TYPE.BOOL.TRUE
FALSE : STD.TYPE.BOOL.FALSE
```

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

c = d   # c equals d is false
c ! d   # c does not equal d is true

c < d   # c is less than d is true
c <= d  # c is less than or equal to d is true
c > d   # c is greater than d is false
c >= d  # c is greater than or equal to d is false
```

Logical operators:

```stult
e : +  # true
f : -  # false

e & f   # e and f is false
e | f   # e or f is true
!e      # not e is false
!f      # not f is true
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

Arrays are ordered lists of values. You can read or replace an item by index, and assigning to the next index appends a new item.

Strings are ordered sequences of characters. Strings are generally used to store text. They can be indexed and updated in much the same way as arrays.

Maps store values under string keys. Map entries can be mutable or immutable, depending on the same capitalization rules that apply to ordinary bindings.

Arrays and maps can grow dynamically as your program runs, subject to available memory and processing time. Strings can also grow when you append characters, but extremely large strings are still limited by the host system. Most programs will never come close to those limits.

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

#### Dot access for maps

Map entries with identifier-shaped string keys can also be accessed with dot access.

Dot access is syntactic sugar for bracket indexing with a string key.

```stult
record : {
	"title": "Example"
	"number": 90
}

record.title
record.number
```

The last two lines are equivalent to:

```stult
record["title"]
record["number"]
```

The dot, the value on the left and the identifier on the right must touch.

Keys that are not valid identifiers must still use bracket indexing:

```stult
record["content-type"]
record["first name"]
record["123"]
```

##### Leading dot for keys in a map

Inside a map literal, a leading dot can be used as shorthand for an identifier-shaped string key.

```stult
person : {
	.NAME : "Andrew"
	.age : 51
}
```

This is equivalent to:

```stult
person : {
	"NAME" : "Andrew"
	"age" : 51
}
```

The identifier spelling is preserved exactly. This means that `.NAME` creates the key `"NAME"` and `.age` creates the key `"age"`.

The usual map-entry mutability rules still apply, so `.NAME` creates an immutable map entry and `.age` creates a mutable map entry.

##### Leading dot for accessing fields within a map's function

Inside a function written inside a map, a leading dot can be used to access fields from that map.

```stult
person : {
	.NAME : "Erica"
	.age : 36

	.handle_birthday : { ()
		BIRTHDAY_GREETING : "Happy birthday, " + .NAME + "."

		.age :+ 1

		(BIRTHDAY_GREETING)
	}
}

STD.IO.OUTPUT.WRITE_LINE(person.handle_birthday())
```

Here, `.NAME` reads the `NAME` field from the surrounding `person` map, and `.age :+ 1` updates the `age` field from that same map.

A leading dot used in this way only looks within the nearest surrounding map. If there is no surrounding map, or if that map does not contain the requested field, the program raises an error.

#### Cloning collections

Collection values can be deeply cloned with `STD.TYPE.COLLECTION.CLONE`.

`CLONE` returns a new mutable collection graph. Nested arrays, maps and strings are cloned recursively, internal aliases and cycles are preserved, and numbers are copied defensively. Functions and builtin functions are reused.

```stult
WRITE_LINE : STD.IO.OUTPUT.WRITE_LINE

original : {
	"nested": {"value": 1}
}

copy : STD.TYPE.COLLECTION.CLONE(original)
copy.nested.value : 2

WRITE_LINE(original.nested.value) # 1
WRITE_LINE(copy.nested.value)     # 2
```

#### Freezing collections

Collection values can also be frozen with `STD.TYPE.COLLECTION.FREEZE`.

Freezing is deep, so nested arrays, maps and strings are frozen too.

```stult
WRITE_LINE : STD.IO.OUTPUT.WRITE_LINE

FREEZE : STD.TYPE.COLLECTION.FREEZE
IS_FROZEN : STD.TYPE.COLLECTION.IS_FROZEN

CONFIG : FREEZE({
	"name": "demo"
	"values": {1, 2, 3}
})

WRITE_LINE(IS_FROZEN(CONFIG))
WRITE_LINE(IS_FROZEN(CONFIG["values"]))
```

A frozen collection cannot be internally modified, even when it is held by a mutable binding.

In practical terms, this means:

- frozen arrays cannot have elements replaced or appended,
- frozen maps cannot have entries added or changed *and*
- frozen strings cannot have characters replaced or appended.

#### Ranges

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
WRITE_LINE : STD.IO.OUTPUT.WRITE_LINE

SUBTRACT : { (A, B)
	(A - B)
}

WRITE_LINE("hello")
WRITE_LINE(SUBTRACT(10, 2))
```

This means that neither `WRITE_LINE ("hello")` nor `WRITE_LINE( SUBTRACT (10, 2))` is a valid function call.

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
	STD.IO.OUTPUT.WRITE_LINE(label, ": ", values)

	(_)
}
```

#### Optional parameters

A user-defined function can mark an ordinary parameter as optional by writing `?` after the parameter name.

```stult
GREET : { (text, suffix?)
	(suffix = _) {
		^("Hello, " + text + "!")
	}

	("Hello, " + text + suffix)
}

GREET("world")       # "Hello, world!"
GREET("world", ".")  # "Hello, world."
```

An omitted optional parameter receives `_`.

Required parameters must come before optional parameters.

```stult
{ (left, right?) (_) }  # valid
{ (left?, right) (_) }  # invalid
```

A variadic parameter, if present, still comes last.

```stult
COLLECT : { (first, second?, ...rest)
	({
		first
		second
		rest
	})
}

COLLECT(1)          # {1, _, {}}
COLLECT(1, 2)       # {1, 2, {}}
COLLECT(1, 2, 3, 4) # {1, 2, {3, 4}}
```

Optional parameters and variadic parameters are different. An omitted optional parameter receives `_`, while a variadic parameter receives an empty array when no remaining arguments are supplied.

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

### Conditionals

Conditionals use a parenthesised condition followed by a brace-enclosed block:

```stult
score : 95

(score >= 90) {
	STD.IO.OUTPUT.WRITE_LINE("excellent")
}
```

An alternative block, which runs when the condition is false, follows `}|{`:

```stult
WRITE_LINE : STD.IO.OUTPUT.WRITE_LINE

score : 80

(score >= 90) {
	WRITE_LINE("excellent")
}|{
	WRITE_LINE("keep going")
}
```

Multiple branches can be chained:

```stult
WRITE_LINE : STD.IO.OUTPUT.WRITE_LINE

score : 10

(score >= 90) {
	WRITE_LINE("excellent")
}|(score >= 70) {
	WRITE_LINE("good")
}|(score >= 50) {
	WRITE_LINE("keep going")
}|(score >= 20) {
	WRITE_LINE("bad")
}|{
	WRITE_LINE("terrible")
}
```

#### Creating a local scope

A conditional with a true condition can also be used as an idiomatic way to create a temporary local scope:

```stult
(+) {
	message : "inside local scope"
	STD.IO.OUTPUT.WRITE_LINE(message)
}
```

Bindings created inside that block do not leak into the surrounding scope.

#### Conditional expressions

A conditional expression chooses between two branch expressions and returns the selected branch value.

```stult
marker : (CONDITION):("*"|" ")
```

The condition must be parenthesised.

```stult
count : 1

label : (count = 1):("item"|"items")
```

The idiomatic form keeps the `:` touching the parentheses on both sides, but horizontal whitespace is accepted around it. (In other words, the `:` must stay on the same line as both the closing parenthesis of the condition and the opening parenthesis of the branch list.)

Only the selected branch is evaluated.

```stult
denominator : 0

safe : (denominator = 0):(0|10 / denominator)
```

In this example, the division branch is not evaluated because the condition is true.

Conditional expressions are useful when a value depends on a condition and both outcomes are simple expressions. Use a conditional statement when either branch needs multiple statements.

#### Match expressions

A match expression chooses a value by comparing one expression with a list of cases.

```stult
TEXT : "yes"

NUMBER : (TEXT):{
	"yes": 50
	"no": 0
	"maybe": 2.5
	_: -1
}
```

Here, `NUMBER` becomes `50` because `TEXT` is `"yes"`.

The subject expression must be parenthesised.

The idiomatic form keeps the `:` touching the parentheses on both sides, but horizontal whitespace is accepted around it. (In other words, the `:` must stay on the same line as both the closing parenthesis of the subject and the opening parenthesis of the arm list.)

Match expressions evaluate the subject once, then check explicit arms before using the `_` default arm.

```stult
TEXT : "yes"

RESULT : (TEXT):{
	_: "unknown"
	"yes": "confirmed"
}
```

`RESULT` is `"confirmed"`, because explicit arms are checked before the default arm, even when `_` appears first.

The current version of match expressions supports only simple literal patterns.

Supported match patterns are:

```text
string literal
number literal
boolean literal
_ default
```

`_` is the fallback branch. It is used only when no explicit arm matches.

Only the selected result expression is evaluated.

```stult
denominator : 0

RESULT : ("safe"):{
	"safe": "ok"
	"divide": 10 / denominator
	_: "fallback"
}
```

In this example, the division arm is not evaluated.

### Error handling

#### Try-catch statements

A try-catch statement lets a program recover from runtime errors. The try block is introduced with an `'`, as shown below:

```stult
'{
	STD.ASSERT.EQUAL(1, 2, "these values should match")
}|{
	STD.IO.OUTPUT.WRITE_LINE("Recovered from the error")
}
```

The catch block may also receive the error message:

```stult
'{
	items : {:}
	items.missing
}|{ (error_message)
	STD.IO.OUTPUT.WRITE_LINE("Error: ", error_message)
}
```

The catch parameter is optional. You may use `_` when you want to show that the error message is intentionally ignored:

```stult
'{
	1()
}|{ (_)
	STD.IO.OUTPUT.WRITE_LINE("Something went wrong")
}
```

Try-catch statements catch runtime errors only. Syntax errors, parsing errors and bytecode compile errors happen before the program runs, so they cannot be caught by a try-catch block.

Break and early return are control flow, not runtime errors. A `^` inside a try block still breaks the nearest loop, and `^(value)` still returns from the nearest function.

### Loops

Loops use double parentheses:

```stult
count : 3

((count > 0)) {
	STD.IO.OUTPUT.WRITE_LINE(count)
	@count :- 1
}
```

Loops may have an after-loop block (which runs once, when the condition no longer holds true):

```stult
WRITE_LINE : STD.IO.OUTPUT.WRITE_LINE

count : 3

((count > 0)) {
	WRITE_LINE(count)
	@count :- 1
}|{
	WRITE_LINE("done")
}
```

#### Infinite loops

An infinite loop uses the true literal:

```stult
((+)) {
	STD.IO.OUTPUT.WRITE_LINE("forever")
}
```

#### Breaking out of a loop

A bare `^` breaks the nearest loop:

```stult
count : 0

((+)) {
	@count :+ 1

	(count = 3) {
		^
	}
}
```

#### Iterating over collections

The same loop syntax can iterate over collections:

```stult
values : {5, 30, 45}

((values)) { (value)
	STD.IO.OUTPUT.WRITE_LINE(value)
}
```

Collection loops can receive up to four parameters:

```stult
WRITE_LINE : STD.IO.OUTPUT.WRITE_LINE

items : {"hat", "coat", "jacket"}

((items)) { (value, key, collection, position)
	WRITE_LINE(position, ": ", key, " -> ", value)
}
```

For arrays and strings, `key` is the numeric index.

For maps, `key` is the string key.

For every type of collection, `position` is the zero-based iteration position.

##### Iterating efficiently over a range

When a loop goes straight over a range like `{1..1000}`, Stult can count through the range directly instead of building the whole array first. This keeps large range loops memory-friendly, as long as the loop body only asks for the value and position, not the collection itself.

```stult
(({1..1000000})) { (value)
	STD.IO.OUTPUT.WRITE_LINE(value)
}
```

#### Function loops

A loop can also use a function as its source.

The function is called repeatedly. Each returned value becomes the loop value for that iteration.

When the function returns `_`, the loop stops.

```stult
COUNTDOWN : { (index)
	(index >= 10) {
		^(_)
	}

	(10 - index)
}

((COUNTDOWN)) { (value)
	STD.IO.OUTPUT.WRITE_LINE(value)
}
```

If the function can accept one argument, the loop passes it the zero-based index.

In the example above, `COUNTDOWN` is called as `COUNTDOWN(0)`, `COUNTDOWN(1)`, `COUNTDOWN(2)` and so on, until it returns `_`.

A function-loop body can also receive the zero-based iteration position as a second argument:

```stult
((COUNTDOWN)) { (value, position)
	STD.IO.OUTPUT.WRITE_LINE(position, ": ", value)
}
```

If the function can accept no arguments, it is called without arguments instead.

```stult
count : 0

NEXT : { ()
	(@count >= 3) {
		^(_)
	}

	@count :+ 1
	(@count)
}

((NEXT)) { (value)
	STD.IO.OUTPUT.WRITE_LINE(value)
}
```

A function with an optional first parameter is treated as able to accept one argument, so it always receives the index.

```stult
NEXT : { (index?)
	(index >= 3) {
		^(_)
	}

	(index)
}
```

Function-loops support user-defined functions. Builtin functions are not function-loop sources. This may change in future versions of Stult, but it holds true for now.

### Commas and newlines

Stult uses both newlines and commas as separators.

Most examples use newlines:

```stult
PRINT : STD.IO.OUTPUT.WRITE_LINE
NAME : "Stult"
COUNT : 3
PRINT(NAME)
```

The same statements can be written with commas:

```stult
PRINT : STD.IO.OUTPUT.WRITE_LINE, NAME : "Stult", COUNT : 3, PRINT(NAME)
```

Commas can also separate function arguments, function parameters, loop parameters, array elements and map entries:

```stult
VALUES : {1, 2, 3}

ADD : { (left, right)
	(left + right)
}

STD.IO.OUTPUT.WRITE_LINE("sum: ", ADD(2, 3))

CONFIG : {"name": "demo", "enabled": +}
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
	"enabled": +
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

Control-flow alternatives use tightly written pipe-based separators. Else branches, catch blocks and after-loop blocks use `}|{`, while else-if branches use `}|(`.

## Standard library

The standard library is available as the immutable binding `STD`.

It is a map containing other maps that, in turn, contain functions.

```stult
STD["ASSERT"]
STD["IO"]
STD["SYSTEM"]
STD["FILE"]
STD["TIME"]
STD["MATH"]
STD["TYPE"]
STD["DATA"]
```

Here is some example code using functions from the standard library:

```stult
WRITE_LINE : STD["IO"]["OUTPUT"]["WRITE_LINE"]
ASSERT : STD["ASSERT"]
COLLECTION_SIZE : STD["TYPE"]["COLLECTION"]["SIZE"]
MATH : STD["MATH"]

ITEMS : {"a", "b", "c"}

ASSERT["EQUAL"](COLLECTION_SIZE(ITEMS), 3, "items should contain three values")
WRITE_LINE("square: ", MATH["SQUARE"](9))
```

Since the standard library is exposed as nested maps, dot-access syntax is a shorter way to write the same string-key lookups:

```stult
STD.IO.OUTPUT.WRITE_LINE
STD["IO"]["OUTPUT"]["WRITE_LINE"]
```

Both forms refer to the same value.

Program arguments are available through `STD["SYSTEM"]["ARGS"]`:

```stult
ARGS : STD.SYSTEM.ARGS
WRITE_LINE : STD.IO.OUTPUT.WRITE_LINE

WRITE_LINE(ARGS)
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
	"is_active": +
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

While JSON cannot contain JavaScript-style comments, STULTON files can contain comments in the same style as Stult source files.

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

## License

Stult is licensed under the Apache License 2.0. See [LICENSE.txt](LICENSE.txt) for the full license text.

Unless otherwise stated, all versions of Stult in this repository, including versions released before the addition of this license file, are licensed under the Apache License 2.0.

You may use, copy, modify and distribute Stult, including for commercial purposes, subject to the terms of the Apache License 2.0.

The name “Stult” refers to the official language and project maintained in this repository. Modified versions and forks should not present themselves as the official Stult project unless accepted by a project maintainer.
