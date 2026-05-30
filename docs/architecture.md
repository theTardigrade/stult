# Architecture

This document aims to give a high-level overview of how Stult is implemented.

Stult is an interpreted scripting language written in Go.

The implementation is split into stages: lexing, parsing, evaluation,
standard-library setup, manifest loading and bundled-executable support.

## Contents

- [Overview](#overview)
- [Source layout](#source-layout)
- [Lexing](#lexing)
- [Parsing](#parsing)
- [Abstract syntax tree](#abstract-syntax-tree)
- [Evaluation](#evaluation)
- [Environments and scope](#environments-and-scope)
- [Values](#values)
- [Standard library](#standard-library)
- [Manifests](#manifests)
- [Bundled executables](#bundled-executables)
- [Command-line entrypoint](#command-line-entrypoint)
- [Errors](#errors)
- [Tests](#tests)

## Overview

A Stult program is processed in this order:

1. Source text is tokenised by the lexer.
2. Tokens are parsed into an abstract syntax tree.
3. The interpreter evaluates the tree in an environment.
4. Standard-library values are exposed through the `STD` map.
5. Manifest files can run multiple source files in a deterministic order.
6. Project source files can be bundled into a standalone executable.

For a single `.stult` file, this path is direct: read the file, lex it, parse it
and evaluate the resulting program.

For a manifest-based project, each listed source file is loaded and evaluated in
order using the same interpreter state. This lets earlier files define bindings
that later files can use.

For a bundled executable, the embedded source files are loaded from the bundle
instead of the filesystem, then run through the same parsing and evaluation path.

## Source layout

The main implementation lives in `src/`.

```text
lexer*.go       tokenisation
parser*.go      parsing source code into AST nodes
ast.go          syntax-tree node definitions
interpreter.go  program evaluation
value.go        runtime value types
env.go          lexical environments and bindings
std*.go         standard-library maps and functions
manifest.go     manifest loading
bundle*.go      bundled executable loading and building
main.go         command-line entrypoint
```

## Lexing

The lexer converts source-code text into tokens.

It handles identifiers, numbers, strings, comments, operators, delimiters,
boolean literals and source locations used in error messages.

The lexer is responsible for recognising Stult's symbolic syntax, including
boolean literals such as `\/` and `/\`, assignment operators such as `:` and
compound assignment operators such as `:+`.

Comments are removed from the token stream, while line and column information is
preserved so later stages can report useful errors.

## Parsing

The parser consumes tokens and builds an abstract syntax tree.

It recognises assignments, compound assignments, expressions, function literals,
calls, conditionals, loops, array literals, map literals and indexing operations.

The parser does not execute code. Its job is both to check that the source code has a valid
shape and to build the tree that the interpreter will later walk.

Stult's parser also has to distinguish between visually similar constructs. For
example:

```stult
(value) {
	...
}
```

is a conditional block, while:

```stult
((value)) {
	...
}
```

is a loop.

Function calls are expressions, and calls require the callee to touch the opening
parenthesis:

```stult
PRINT("hello")
```

Function literals use a block with a parameter list:

```stult
ADD : { (A, B)
	(A + B)
}
```

## Abstract syntax tree

The abstract syntax tree, or AST, is the structured representation of a parsed
Stult program.

Each AST node represents a language construct such as an assignment, expression,
conditional, loop, function literal, array literal or map literal.

The interpreter evaluates this tree directly. Stult is therefore currently a
tree-walk interpreter rather than a compiler to bytecode or native code, even though
the bundler-build tool can produce a native executable with source-code embedded.

## Evaluation

The interpreter walks the syntax tree and evaluates nodes in an environment.

Bindings are stored in environments. Uppercase names are immutable, while names
containing lowercase letters are mutable.

Plain reads search outward through enclosing environments. Plain assignment
writes to the current environment.

The `@` prefix explicitly reads from or writes to an outer environment. The
prefix must be used when writing to an outer scope, but it is optional when
reading from an outer scope unless you need to skip a binding of the same name
found in the current scope.

For example:

```stult
count : 0

((\/)) {
	@count :+ 1

	(count = 3) {
		^
	}
}
```

Here `@count :+ 1` writes to the `count` binding in the outer scope.

## Environments and scope

Environments form a chain.

A new environment is created for constructs such as functions, conditionals
and loops. That is to say, they have their own scope.

## Values

Runtime values include numbers, strings, booleans, arrays, maps, functions,
native builtin functions and the void value.

Stult uses one high-precision numeric type rather than separate integer and
floating-point types.

Arrays, maps and strings are collection values, which means that they can be iterated over in loops.

Functions are also values, which means they can be stored in bindings, maps and arrays.

The void value is represented by `_` and is used when a value is intentionally
discarded or when a function returns has no meaningful result.

## Standard library

The standard library is exposed through the `STD` map.

It is organised into nested maps under keys such as `"IO"`, `"FILE"`, `"TIME"`,
`"MATH"`, `"TYPE"` and `"DATA"`.

For example:

```stult
PRINT : STD["IO"]["PRINT"]
SIN : STD["MATH"]["TRIG"]["SIN"]
CSV_PARSE : STD["DATA"]["CSV"]["PARSE"]
```

This map-shaped structure keeps the standard library explicit and inspectable
from Stult code.

For the public standard-library reference, see [`std.md`](std.md).

## Manifests

A manifest-based project can list multiple Stult source files.

Files run deterministically in the order specified in the manifest file. This
allows one file to define bindings that later files can use.

A small project might use one file for shared bindings, one for configuration,
one for helper functions and one for the main program.

For example:

```text
examples/animated_sine_wave/
  manifest.stulton
  bindings.stult
  config.stult
  wave.stult
  screen.stult
  main.stult
```

The manifest gives the project an explicit execution order.

For more information about manifest files, please see [manifests.md](manifests.md).

## Bundled executables

Stult can bundle a manifest-based project into a standalone executable.

The bundled executable contains the Stult runtime and the project source files,
so the resulting tool can be copied and run without shipping the source tree
separately.

This is one of Stult's main practical features. A project can be developed as
multiple readable `.stult` files, and then distributed as a single executable.

The bundled executable still runs the Stult source through the interpreter. The
bundle changes how the source files are stored and loaded, not the language
semantics.

## Command-line entrypoint

The command-line entrypoint handles direct source evaluation, single-file
execution, directory execution, manifest execution and project bundling.

The same runtime path is used for source files from disk and source files loaded
from a bundled executable.

In broad terms, the command-line layer decides where the source comes from. Once
the source has been loaded, the lexer, parser and interpreter handle it in the
same way.

## Errors

Lexer, parser and runtime errors include source-location information where
possible.

The goal is for errors to point back to the relevant file, line and column so a
Stult program can be debugged from the source text rather than from internal Go
details.

Parser errors are reported before evaluation begins. Runtime errors are reported
while the interpreter is walking the AST.

## Tests

The Go test suite checks example-code behaviour.

The example tests help to ensure that these public programs continue working as
the language changes.

This is especially useful for Stult because the examples double as public
documentation. If an example breaks, either the implementation has changed
improperly or the example needs to be updated to match the language.