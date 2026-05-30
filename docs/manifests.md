# Manifests

A Stult manifest describes a multi-file Stult project.

Manifests let you split a program across several `.stult` files, then run those
files in a deterministic order using one shared interpreter state.

This is useful for larger projects, reusable helper files, configuration files
and bundled executables.

## Contents

- [Manifest filenames](#manifest-filenames)
- [Basic STULTON manifest](#basic-stulton-manifest)
- [Basic JSON manifest](#basic-json-manifest)
- [Run order](#run-order)
- [Shared interpreter state](#shared-interpreter-state)
- [Paths](#paths)
- [`RUN` and `run`](#run-and-run)
- [Single-file manifests](#single-file-manifests)
- [Running a manifest project](#running-a-manifest-project)
- [Building a bundled executable](#building-a-bundled-executable)
- [Manifests are not imports](#manifests-are-not-imports)
- [Good manifest habits](#good-manifest-habits)

## Manifest filenames

A project may use either of these manifest files:

```text
manifest.stulton
manifest.json
```

`manifest.stulton` is the default and is the most Stult-native option.

`manifest.json` is also supported for tools or workflows that prefer to use JSON.

## Basic STULTON manifest

A STULTON manifest uses Stult's native data notation:

```stulton
{
	"RUN": {
		"bindings.stult"
		"helpers.stult"
		"main.stult"
	}
}
```

This manifest runs the three files:

```text
bindings.stult
helpers.stult
main.stult
```

in that order.

## Basic JSON manifest

A JSON manifest uses a lowercase `run` field:

```json
{
	"run": [
		"bindings.stult",
		"helpers.stult",
		"main.stult"
	]
}
```

This has the same effect as the STULTON manifest above.

## Run order

Files in a manifest-based project run deterministically, in the order specified
in the manifest file.

All listed files are evaluated by the same interpreter, so bindings created by
earlier files can be used by later files. That is to say, they share global scope.

For example, a project might contain:

```text
example_project/
  manifest.stulton
  bindings.stult
  helpers.stult
  main.stult
```

With this manifest:

```stulton
{
	"RUN": {
		"bindings.stult"
		"helpers.stult"
		"main.stult"
	}
}
```

`bindings.stult` might define shared standard-library aliases:

```stult
PRINT : STD["IO"]["PRINT"]
SIZE : STD["TYPE"]["COLLECTION"]["SIZE"]
```

`helpers.stult` might define functions that use those aliases:

```stult
PRINT_SIZE : { (value)
	PRINT("Size: ", SIZE(value))

	(_)
}
```

`main.stult` can then use both:

```stult
ITEMS : { "a", "b", "c" }

PRINT("******")
PRINT_SIZE(ITEMS)
PRINT("******")
```

## Shared interpreter state

A manifest does not create a separate interpreter for each file.

The files run in one shared interpreter state, which means:

- earlier files can define bindings for later files,
- helper functions can be split into their own files,
- configuration values can be kept separate from main program logic *and*
- file order matters.

This makes manifests a simple project system rather than an import system.

## Paths

Manifest run paths are normally written relative to the directory containing the
manifest file.

For example:

```text
tools/report/
  manifest.stulton
  src/bindings.stult
  src/main.stult
```

can use:

```stulton
{
	"RUN": {
		"src/bindings.stult"
		"src/main.stult"
	}
}
```

Use forward slashes in manifest paths for portability:

```text
src/main.stult
```

rather than:

```text
src\main.stult
```

## `RUN` and `run`

In `manifest.stulton`, either `RUN` or `run` may be used:

```stulton
{
	"RUN": {
		"main.stult"
	}
}
```

or:

```stulton
{
	"run": {
		"main.stult"
	}
}
```

`RUN` is more idiomatic in STULTON because uppercase map keys fit Stult's style
for immutable, project-level values.

Do not include both `RUN` and `run` in the same STULTON manifest.

In `manifest.json`, use lowercase `run`:

```json
{
	"run": [
		"main.stult"
	]
}
```

## Single-file manifests

The run field may contain a single source file.

STULTON:

```stulton
{
	"RUN": "main.stult"
}
```

JSON:

```json
{
	"run": "main.stult"
}
```

This is useful when you want a project directory and bundled executable support,
even though the project currently has only one source file.

## Running a manifest project

Run a project directory:

```bash
stult examples/bool
```

Run a manifest file directly:

```bash
stult examples/bool/manifest.stulton
```

Run from inside a project directory:

```bash
stult
```

With no arguments, `stult` searches upward from the current directory for a
manifest.

## Building a bundled executable

Manifest projects can be bundled into standalone executables:

```bash
stult build examples/bool -o bool-app
```

Then run the generated executable:

```bash
./bool-app
```

The bundled executable contains the Stult runtime and the project source files.

This means the project can be distributed as a single executable without shipping
the source tree separately.

## Manifests are not imports

A manifest is an execution list, not an import graph.

Files are not imported on demand. They run in the order listed by the manifest.

That means this order works:

```stulton
{
	"RUN": {
		"bindings.stult"
		"main.stult"
	}
}
```

if `main.stult` uses bindings from `bindings.stult`.

This order may fail:

```stulton
{
	"RUN": {
		"main.stult"
		"bindings.stult"
	}
}
```

because `main.stult` runs before `bindings.stult`.

## Good manifest habits

Use clear file names, such as:

```text
bindings.stult
config.stult
helpers.stult
main.stult
```

List files from most general to most specific:

```text
shared bindings
configuration
helper functions
main program
```

Keep the final file focused on running the program.

Prefer relative paths so the project can be moved, copied or bundled easily.

Use `manifest.stulton` unless you have a good reason to prefer JSON.