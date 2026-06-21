# Versions

This document summarises notable Stult changes by released version.

For downloadable binaries and checksums, see the GitHub Releases page once releases are available.

Stult is currently pre-1.0. Until v1.0.0, language syntax, standard-library shape, bytecode details and command-line behaviour may still change between minor versions.

## Changelog

### Future release

#### Project

* Added the Apache License 2.0 as the project license and clarified that it applies to earlier repository versions unless otherwise stated.

#### Standard library

* Added `STD.TYPE.STRING.REPEAT` for repeating text a non-negative whole number of times.

### v0.8.0

#### Language

* Changed boolean literals from `\/` and `/\` to `+` and `-`.

#### Collections

* Added leading-dot map keys such as `.name : value` as shorthand for identifier-shaped string keys inside map literals.
* Added leading-dot field access for functions written inside maps, allowing `.name` reads and assignments to resolve against the nearest captured surrounding map.

#### Standard Library

* Added `STD.ASSERT.FALSE`.
* Made assertion messages optional for `STD.ASSERT.TRUE`, `STD.ASSERT.FALSE` and `STD.ASSERT.EQUAL`.
* Added directory helpers `STD.FILE.LIST`, `STD.FILE.IS_FILE`, `STD.FILE.IS_DIR` and `STD.FILE.MAKE_DIR`, with optional boolean arguments for absolute list paths and recursive directory creation.
* Changed `STD.FILE.READ` to accept optional `useBytes`, offset and length arguments for ranged text reads and byte-array reads.
* Changed `STD.FILE.WRITE` to accept byte-array content, accept an optional append-mode boolean argument and remove `STD.FILE.APPEND`.
* Moved file-path helpers from `STD.PATH` to `STD.FILE.PATH`.
* Split `STD.IO` into `STD.IO.INPUT` and `STD.IO.OUTPUT`, moving input helpers under `INPUT` and output helpers under `OUTPUT`.
* `STD.IO.OUTPUT.PRINT` was removed; use `WRITE_LINE` instead. Standard-error output now has both `WRITE_ERROR` and `WRITE_ERROR_LINE`.

### v0.7.8

* Centralised collection access and mutation behind value-owned method layers for arrays, strings and maps. This is an internal implementation cleanup with no intended user-visible language change.
* Array behaviour is now owned by `Array` methods such as `Len`, `Get`, `Set`, `Append`, `Clear` and `ForEach`, keeping ordinary and overflow storage details representation-local.
* String behaviour is now owned by `String` methods such as `Len`, `Get`, `Set`, `Append`, `Clear` and `ForEach`, keeping direct rune-slice access representation-local.
* Map behaviour is now owned by `Map` methods such as `Len`, `Get`, `Set`, `Define`, `ForEach`, `Keys` and `Clear`, while preserving the distinction between map frozen state and per-entry binding immutability.
* Cleaned up number internals by moving useful conversion and cloning operations onto `Number` methods and removing the unused `numberToArrayIndex` helper.

### v0.7.7

* Reworked the internal array representation into ordinary slice-backed storage plus overflow chunks for extremely large dense arrays.
* Array length and array iteration now use Stult-number indices rather than being inherently tied to the host slice index size.
* Clarified that arrays remain finite, dense and resource-bounded, while strings remain host-representation-bounded Unicode code-point sequences.
* Renamed the internal collection frozen-state fields from `IsImmutable` to `IsFrozen` for arrays, maps and strings, while keeping binding and map-entry rebinding metadata as `IsImmutable`.

### v0.7.6

* Added `STD["TYPE"]["COLLECTION"]["GET"]` for safe map, array and string access with an optional default value.
* `GET` returns the default, or void when no default is supplied, for missing map keys and out-of-bounds array/string indices while still raising runtime errors for non-collection receivers and invalid key/index kinds.

### v0.7.5

* Made array and map value formatting cycle-safe, displaying `<cyclical array>` or `<cyclical map>` for recursive collection references instead of recursing indefinitely.
* Added `STD["TYPE"]["COLLECTION"]["CLONE"]` for deep, cycle-safe, alias-preserving collection cloning.
* `CLONE` returns mutable cloned arrays, maps and strings, preserves map-entry mutability, copies numbers defensively and reuses function and builtin function values.

### v0.7.4

* Fixed bytecode compilation so early return outside a function is rejected instead of being emitted as a top-level return.
* Fixed bytecode outer-assignment fallback so `@name : value` updates only an existing outer global binding and never creates a missing global.
* Fixed bytecode map literal construction so duplicate keys raise a runtime error instead of silently overwriting an earlier entry.
* Fixed bytecode closure captures for block and loop locals so captured cells survive later local-slot resets and reuse.

### v0.7.3

* Added try-catch statements using `'{ ... },{ ... }` for recovering from runtime errors, with an optional catch parameter for the error message.
* Added percentage-suffixed number literals, so `50%` evaluates to `0.5`.

### v0.7.2

* Added an internal range-loop optimisation for direct single-range loop sources with zero, one or two loop parameters, avoiding unnecessary array materialisation while preserving observable collection semantics.
* Fixed range evaluation so exact integer range bounds and steps are not narrowed to host-sized integers.

### v0.7.1

* Added function loops, allowing user-defined functions to act as generator-style loop sources that stop by returning `_`.

### v0.7.0

* Added dot-access syntax for string-key map access, so `value.key` is equivalent to `value["key"]`.
* Added optional user-function parameters using `?`, with omitted optional parameters receiving `_`.
* Added parenthesized conditional expressions using `(CONDITION)?(WHEN_TRUE, WHEN_FALSE)`.
* Added match expressions using `(SUBJECT)?{ ... }` with scalar literal arms and `_` fallback.
* Added `STD["MATH"]["RAND"]` with random number, integer, choice and shuffle helpers.
* Added examples and documentation for the new syntax and random helpers.

### v0.6.1

* Added `STD["MATH"]["REM"]` for truncating remainder arithmetic.
* Clarified the difference between mathematical modulo and truncating remainder in the standard-library documentation.

### v0.6.0

* Reworked number handling so whole-number values are theoretically unbounded, subject to available memory and processing time.
* Added bounded decimal arithmetic, with digits after the decimal point rounded to a maximum number of decimal places.
* Separated ordinary number display from the maximum decimal-place limit.
* Added `STD["TYPE"]["NUMBER"]["DEFAULT_DECIMAL_PLACES"]` and `STD["TYPE"]["NUMBER"]["MAX_DECIMAL_PLACES"]`.
* Added `STD["TYPE"]["NUMBER"]["FORMAT"]` for fixed decimal formatting.
* Added `STD["TYPE"]["NUMBER"]["FORMAT_SCIENTIFIC"]` for scientific notation.
* Removed safe-integer and precision constants from `STD["TYPE"]["NUMBER"]` that have become misleading.
* Updated exact math helpers such as rounding and truncation to use the new decimal representation directly.

### v0.5.1

* Added an internal small-integer representation for number values, promoting to high-precision floating-point numbers when needed.
* Updated architecture documentation for the internal number representation.
* Expanded standard-library documentation for constants.

### v0.5.0

Initial public release.

This release introduces Stult as a small scripting language with a bytecode-first runtime, an interpreter fallback, manifest-based projects, a map-shaped standard library and support for bundled standalone executables.

#### Language

* Added `.stult` source files.
* Added one high-precision numeric value type.
* Added strings, arrays, maps, booleans, void, functions and builtin functions.
* Added immutable and mutable bindings based on identifier spelling.
* Added assignment with `:`.
* Added equality with `=` and inequality with binary `!`.
* Added compound assignment with `:+`, `:-`, `:*` and `:/`.
* Added logical operators `&`, `|` and prefix `!`.
* Added boolean literals `\/` and `/\`.
* Added void `_`.
* Added line comments with `#`.
* Added bounded comments with `## ... ##`.
* Added conditionals.
* Added dynamic loops over booleans and collections.
* Added loop after-blocks with `},{`.
* Added break with `^`.
* Added early return with `^(value)`.
* Added functions with a required final return expression.
* Added variadic function parameters.
* Added array ranges with inclusive, exclusive and stepped forms.
* Added indexing for arrays, maps and strings.
* Added indexed assignment for arrays, maps and strings.
* Added outer-scope reads and writes with `@name`.

#### Collections

* Added mutable arrays.
* Added mutable maps with per-entry binding mutability.
* Added mutable strings with character replacement and append-at-end indexed assignment.
* Added frozen collection support for arrays, maps and strings.
* Added deep freezing for nested arrays, maps and strings.
* Added collection identity semantics for arrays and maps.
* Added string equality by contents.

#### Standard library

* Added the immutable global `STD` binding.
* Added top-level standard-library maps:

```stult
STD["ASSERT"]
STD["IO"]
STD["SYSTEM"]
STD["FILE"]
STD["PATH"]
STD["TIME"]
STD["MATH"]
STD["TYPE"]
STD["DATA"]
```

* Added assertion helpers.
* Added IO helpers for writing output, writing errors, reading lines and prompting.
* Added system helpers for program arguments, current working directory, environment variables and process exit.
* Added file helpers for reading, writing, appending, checking, deleting, renaming, copying and sizing files.
* Added path helpers for joining, cleaning and inspecting filesystem paths.
* Added time helpers for timestamps, sleeping and calendar values.
* Added maths helpers for constants, numeric utilities and trigonometry.
* Added type-checking and type-conversion helpers.
* Added string helpers for conversion, trimming, case conversion, searching, replacement, splitting and joining.
* Added array append support.
* Added map key and value helpers.
* Added collection helpers for size, emptiness, membership, clearing, freezing and frozen-state checks.
* Added CSV, JSON and STULTON data helpers.

#### STULTON

* Added STULTON as Stult's native data notation.
* Added STULTON support for void, booleans, numbers, strings, arrays and maps.
* Added STULTON parsing and serialisation through the standard library.
* Added STULTON manifest support.
* Added support for Stult-style comments in STULTON text.

#### Manifests

* Added manifest-based project execution.
* Added `manifest.stulton`.
* Added `manifest.json`.
* Added ordered manifest run lists.
* Added shared runtime state across files in a manifest.
* Added upward manifest discovery for project execution.

#### Command line

* Added `stult run`.
* Added bytecode execution as the default runtime.
* Added `stult run --bytecode`.
* Added `stult run --interpreter`.
* Added `stult dump` for bytecode disassembly.
* Added `stult build` for standalone executable creation.
* Added `stult build --bytecode`.
* Added `stult build --interpreter`.
* Added source-string execution with `-e` / `--eval`.

#### Runtime implementation

* Added lexer, parser and AST pipeline.
* Added bytecode compiler.
* Added bytecode virtual machine.
* Added tree-walk interpreter.
* Added shared runtime context for standard-library builtins.
* Added source locations to parser and runtime errors.
* Added bytecode disassembly output for debugging.

#### Parser behaviour

* Added support for grouped conditional expressions that begin with touching `((`.
* Added support for grouped loop expressions such as `(((A & B) | (C & D)))`.
* Added support for loop bodies whose first statement is a conditional.
* Added support for function return expressions that begin with grouped expressions using touching `((`.
* Added support for array literals that begin with grouped expressions.

#### Bundling

* Added bytecode bundles.
* Added source/interpreter bundles.
* Added standalone executable startup from embedded bundles.
* Added support for bundling single source files and manifest-based projects.
* Added safeguards to avoid overwriting the currently running executable during bundling.

#### Examples

* Added basic language examples.
* Added standard-library overview example.
* Added manifest examples.
* Added CSV-to-JSON converter example.
* Added animated sine wave project.
* Added autonomous snake project.
* Added public example-test programs under `examples/tests/`.

#### Testing

* Added example-based runtime parity tests.
* Added interpreter and bytecode comparison for files under `examples/tests/`.
* Added parser regression examples for grouped conditionals, grouped loops, grouped return expressions and grouped array expressions.

#### Documentation

* Added `README.md`.
* Added standard-library reference documentation.
* Added manifest documentation.
* Added bundling documentation.
* Added architecture documentation.
* Added examples documentation.
* Added example-test documentation.
* Added contributing notes.

#### Release infrastructure

* Added GitHub Actions release workflow.
* Added automated test execution during release.
* Added distribution binary build helper.
* Added checksum generation for release artefacts.

#### Notes

This is the first public release of Stult.

The version number is intentionally below v1.0.0. The language is usable, but the public surface is still expected to evolve before a stable v1.0.0 release.
