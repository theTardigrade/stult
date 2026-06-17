# Bundled executables

Stult can bundle a single source file or manifest-based project into a standalone executable.

A bundled executable contains the Stult runtime and the program bundle, so a small Stult tool can be copied and run without shipping the source tree separately.

By default, `stult build` creates a bytecode bundle.

Source/interpreter bundles are available with the following command: `stult build --interpreter`.

## Contents

- [What bundling does](#what-bundling-does)
- [Bundle modes](#bundle-modes)
  - [Bytecode bundles](#bytecode-bundles)
  - [Source/interpreter bundles](#sourceinterpreter-bundles)
- [Basic example](#basic-example)
- [Building from a project directory](#building-from-a-project-directory)
- [Building from a single source file](#building-from-a-single-source-file)
- [Output paths](#output-paths)
- [Building without Go installed](#building-without-go-installed)
- [What gets embedded](#what-gets-embedded)
- [Subdirectories](#subdirectories)
- [Nested manifests](#nested-manifests)
- [Manifest order still matters](#manifest-order-still-matters)
- [Manifest paths](#manifest-paths)
- [Recommended bundled project structure](#recommended-bundled-project-structure)
- [Running after copying](#running-after-copying)
- [Bundling is not native compilation](#bundling-is-not-native-compilation)
- [Development workflow](#development-workflow)
- [Troubleshooting](#troubleshooting)
  - [The project must have a manifest](#the-project-must-have-a-manifest)
  - [Do not use both manifest formats at once](#do-not-use-both-manifest-formats-at-once)
  - [Non-Stult data files are not embedded](#non-stult-data-files-are-not-embedded)
  - [Bytecode bundles do not embed source files](#bytecode-bundles-do-not-embed-source-files)
  - [Absolute manifest paths](#absolute-manifest-paths)
  - [The output executable must not overwrite the running executable](#the-output-executable-must-not-overwrite-the-running-executable)

## What bundling does

Bundling takes a Stult source file or project directory and creates a new executable.

When the generated executable starts, it detects its embedded bundle and runs the bundled program automatically.

The generated executable behaves like a normal command-line program:

```bash
./my-tool
```

On Windows:

```powershell
.\my-tool.exe
```

The user does not need to invoke `stult run` when running the generated executable.

The bundled executable still uses the Stult runtime. The difference is that the Stult program is packaged inside the executable.

## Bundle modes

Stult supports two bundle modes:

```bash
stult build --bytecode my_tool -o my-tool
stult build --interpreter my_tool -o my-tool
```

`--bytecode` is the default.

`--interpreter` is available when you explicitly want a source bundle that runs through the tree-walk interpreter.

### Bytecode bundles

A bytecode bundle embeds:

- the Stult runtime,
- a manifest,
- compiled bytecode *and*
- bytecode metadata needed to map manifest entries to bundled bytecode.

Build a bytecode bundle:

```bash
stult build my_tool -o my-tool
```

This is the same as:

```bash
stult build --bytecode my_tool -o my-tool
```

A bytecode bundle does not need the original `.stult` source files at runtime.

This is the recommended bundle mode for ordinary distribution.

### Source/interpreter bundles

A source/interpreter bundle embeds:

- the Stult runtime,
- a manifest
- the `.stult` source files needed by that manifest.

Build a source/interpreter bundle:

```bash
stult build --interpreter my_tool -o my-tool
```

When the generated executable starts, it loads the embedded source files and runs them through the tree-walk interpreter.

This mode is useful when you explicitly want the generated executable to include source code.

## Basic example

Given this project:

```text
example_tool/
  manifest.stulton
  main.stult
```

with this manifest:

```stulton
{
	"RUN": {
		"main.stult"
	}
}
```

and this source file:

```stult
PRINT : STD["IO"]["PRINT"]

PRINT("Hello from a bundled Stult executable.")
```

Build the default bytecode bundle:

```bash
stult build ./example_tool -o example-tool
```

Run the generated executable:

```bash
./example-tool
```

On Windows:

```powershell
.\example-tool.exe
```

Build a source/interpreter bundle instead:

```bash
stult build --interpreter ./example_tool -o example-tool
```

## Building from a project directory

The build command has this form:

```bash
stult build [--bytecode|--interpreter] <project-directory> -o <output-executable>
```

For example:

```bash
stult build examples/projects/bool -o bool-app
```

The project directory must contain one root manifest file:

```text
manifest.stulton
```

or:

```text
manifest.json
```

Do not put both manifest files in the same project root.

## Building from a single source file

You can also build a standalone executable from a single `.stult` source file:

```bash
stult build examples/calculate_circle_area_from_map.stult -o circle-area-app
```

Stult creates an internal manifest for the single source file, then bundles that program.

The default is still a bytecode bundle.

Use `--interpreter` if you want a source/interpreter bundle:

```bash
stult build --interpreter examples/calculate_circle_area_from_map.stult -o circle-area-app
```

## Output paths

Use `-o` or `--output` to choose the generated executable path:

```bash
stult build examples/projects/bool -o bool-app
```

```bash
stult build examples/projects/bool --output bool-app
```

If no output path is given, Stult uses the project directory name or source file name:

```bash
stult build examples/projects/bool
```

```bash
stult build examples/calculate_circle_area_from_map.stult
```

On Windows, `.exe` is added automatically when needed.

If the project directory is the current directory, the default output name is:

```text
stult-app
```

or, on Windows:

```text
stult-app.exe
```

## Building without Go installed

You need a `stult` executable to build a bundled executable.

You do not need Go installed if you are using a distributed Stult binary:

```bash
stult build my_tool -o my-tool
```

Go is only needed if you are building Stult itself from source.

For example, from the Stult repository:

```bash
go run ./util/build_helper.go local
```

or:

```bash
go build -o stult ./src
```

Once you have a Stult binary, the normal bundling command is:

```bash
stult build my_tool -o my-tool
```

## What gets embedded

What gets embedded depends on the bundle mode.

A bytecode bundle embeds:

```text
manifest.stulton or manifest.json
.stult-bytecode-bundle
.stult-bytecode/run-map.json
.stult-bytecode/... compiled bytecode files
```

A source/interpreter bundle embeds:

```text
manifest.stulton or manifest.json
.stult source files
```

Other files are not embedded by default.

For example, this project embeds the manifest and compiled bytecode in the default build:

```text
my_tool/
  manifest.stulton
  bindings.stult
  main.stult
```

This project does not embed `data.csv`:

```text
my_other_tool/
  manifest.stulton
  bindings.stult
  main.stult
  data.csv
```

If a bundled program needs data, either embed the data in Stult source code or make sure the program reads it from an external path at runtime.

## Subdirectories

`.stult` files in subdirectories can be included in a bundle.

For example:

```text
my_tool/
  manifest.stulton
  src/bindings.stult
  src/helpers.stult
  src/main.stult
```

can use this manifest:

```stulton
{
	"RUN": {
		"src/bindings.stult"
		"src/helpers.stult"
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

## Nested manifests

Only the root manifest belongs to the bundled project.

Nested manifest files are not included as project manifests. Keep one manifest at the project root and use it to list the files that should run.

For example:

```text
my_tool/
  manifest.stulton
  main.stult
```

is the intended shape.

Avoid layouts like:

```text
my_tool/
  manifest.stulton
  nested/manifest.stulton
```

## Manifest order still matters

Bundled executables run the bundled manifest in the same order as an ordinary manifest project.

Files run deterministically in the order specified in the manifest file.

This means a project can use separate files for setup, configuration, helpers and main program logic:

```text
my_tool/
  manifest.stulton
  bindings.stult
  config.stult
  helpers.stult
  main.stult
```

```stulton
{
	"RUN": {
		"bindings.stult"
		"config.stult"
		"helpers.stult"
		"main.stult"
	}
}
```

Earlier files can define bindings used by later files.

The default bytecode runtime preserves this shared runtime state across manifest files.

## Manifest paths

Manifest run paths are usually relative to the directory containing the manifest.

For example:

```stulton
{
	"RUN": {
		"src/bindings.stult"
		"src/main.stult"
	}
}
```

Relative paths are the most portable choice and are recommended for bundled projects.

Bytecode bundles can also handle absolute manifest run paths by compiling those files at build time and storing a mapping from the manifest entry to the embedded bytecode.

Source/interpreter bundles should avoid absolute manifest run paths. If a source/interpreter bundle contains an absolute manifest entry, the generated executable may try to read that absolute path from the runtime filesystem.

## Recommended bundled project structure

A good structure for a bundled command-line tool is:

```text
my_tool/
  manifest.stulton
  bindings.stult
  config.stult
  helpers.stult
  main.stult
```

With such a structure, the Stult files might contain the following:

```text
bindings.stult  shared standard-library aliases
config.stult    constants and settings
helpers.stult   functions
main.stult      program entrypoint
```

Run it during development with:

```bash
stult run my_tool
```

Then build it with:

```bash
stult build my_tool -o my-tool
```

## Running after copying

A bundled executable can be copied to another compatible computer and run there.

For example, a Linux amd64 executable should be run on a compatible Linux amd64 system. A Windows executable should be built for Windows and will normally use the `.exe` extension.

The user running the bundled executable does not need:

- the original project directory,
- Go installed *or*
- a separate `stult` command.

Source/interpreter bundles generally include the `.stult` source files inside the generated executable, so the user does not need a separate copy of the source tree.

Bytecode bundles go further: they do not rely on the original `.stult` source files at runtime because they run the embedded compiled bytecode instead.

The user only needs the generated executable to run the program, as well as a compatible operating system and CPU architecture.

## Bundling is not native compilation

A bytecode bundle ordinarily compiles Stult source to Stult bytecode, not native machine code.

The generated executable contains the Stult runtime and bytecode for the Stult program. At startup, the runtime loads the embedded bytecode and runs it in the bytecode VM.

A source/interpreter bundle contains source instead of bytecode. At startup, the runtime loads the embedded source and runs it through the tree-walk interpreter.

To the end user, both modes behave like a normal command-line program: the bundled executable can be run directly, without separately invoking `stult` or shipping the source tree.

## Development workflow

A typical workflow is:

```bash
stult run my_tool
```

Then, when the project works:

```bash
stult build my_tool -o my-tool
```

Then run the bundled executable:

```bash
./my-tool
```

If you explicitly want to use the interpreter runtime:

```bash
stult run --interpreter my_tool
```

If you explicitly want to build a source/interpreter bundle:

```bash
stult build --interpreter my_tool -o my-tool
```

If you are working from the Stult source repository and have not built a local `stult` binary yet, you can use:

```bash
go run ./util/build_helper.go local
```

Then use the generated binary:

```bash
./stult build my_tool -o my-tool
```

## Troubleshooting

### The project must have a manifest

`stult build` expects a project directory to contain:

```text
manifest.stulton
```

or:

```text
manifest.json
```

If neither file exists, the build fails.

Single `.stult` files can be built directly and do not need a manifest beside them.

### Do not use both manifest formats at once

Use either `manifest.stulton` or `manifest.json`.

Do not include both in the same project root.

### Non-Stult data files are not embedded

Only Stult program files and bundle metadata are embedded by default.

If a program reads a file such as `data.csv`, that file must still exist at runtime unless its contents are embedded as a string in Stult source code.

### Bytecode bundles do not embed source files

The default bundle does not embed the original `.stult` source files, only the bytecode needed to run the program.

This is intentional.

Use `--interpreter` if you explicitly want a source bundle:

```bash
stult build --interpreter my_tool -o my-tool
```

### Avoid absolute manifest paths

Prefer relative manifest paths for bundled projects.

Bytecode bundles can compile absolute manifest run files at build time, but relative paths are easier to move, copy and understand.

Source/interpreter bundles should avoid absolute manifest run paths. Absolute run paths are read from the target machine’s filesystem at runtime rather than from the embedded bundle.

### The output executable must not overwrite the running executable

Stult refuses to overwrite the executable that is currently running the build.

For example, if you are running the build with `stult`, do not attempt to write the output back to `stult`:

```bash
stult build my_tool -o stult
```

Choose a different output path:

```bash
stult build my_tool -o my-bundled-tool
```
