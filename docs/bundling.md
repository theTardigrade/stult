# Bundled executables

Stult can bundle a manifest-based project into a standalone executable.

A bundled executable contains the Stult runtime and the project source files, so
a small Stult tool can be copied and run without shipping the source tree
separately.

This is one of Stult's most practical features.

## Contents

- [What bundling does](#what-bundling-does)
- [Basic example](#basic-example)
- [Building from a project directory](#building-from-a-project-directory)
- [Output paths](#output-paths)
- [Building without Go installed](#building-without-go-installed)
- [What gets embedded](#what-gets-embedded)
- [Subdirectories](#subdirectories)
- [Nested manifests](#nested-manifests)
- [Manifest order still matters](#manifest-order-still-matters)
- [Recommended bundled project structure](#recommended-bundled-project-structure)
- [Running after copying](#running-after-copying)
- [Bundling is not compilation](#bundling-is-not-compilation)
- [Development workflow](#development-workflow)
- [Troubleshooting](#troubleshooting)
  - [The project must have a manifest](#the-project-must-have-a-manifest)
  - [Do not use both manifest formats at once](#do-not-use-both-manifest-formats-at-once)
  - [Non-Stult data files are not embedded](#non-stult-data-files-are-not-embedded)
  - [The output executable must not overwrite the running executable](#the-output-executable-must-not-overwrite-the-running-executable)

## What bundling does

Bundling takes a Stult project directory and creates a new executable.

The generated executable contains:

- the Stult runtime,
- the root manifest file *and*
- the project's `.stult` source files.

When the generated executable starts, it detects the embedded bundle, loads the
bundled manifest and runs the files listed by that manifest.

The source files still run through the Stult interpreter. Bundling changes how
the project is packaged and loaded. It does not compile Stult source code into
native machine code.

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

Build it with:

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

## Building from a project directory

The build command has this form:

```bash
stult build <project-directory> -o <output-executable>
```

For example:

```bash
stult build examples/bool -o bool-app
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

## Output paths

Use `-o` or `--output` to choose the generated executable path:

```bash
stult build examples/bool -o bool-app
```

```bash
stult build examples/bool --output bool-app
```

If no output path is given, Stult uses the project directory name:

```bash
stult build examples/bool
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

Go is only needed if you are building or running Stult from its Go source code,
for example with:

```bash
go run ./src build my_tool -o my-tool
```

or:

```bash
go build -o stult ./src
```

## What gets embedded

The bundle currently includes:

- the root `manifest.stulton` or `manifest.json` file *and*
- `.stult` source files inside the project directory.

Other files are not embedded by default.

For example, this project embeds the manifest and both Stult files:

```text
my_tool/
  manifest.stulton
  bindings.stult
  main.stult
```

This project embeds the manifest and the `.stult` files, but not `data.csv`:

```text
my_other_tool/
  manifest.stulton
  bindings.stult
  main.stult
  data.csv
```

If a bundled program needs data, either embed the data in Stult source code or
make sure the program reads it from an external path at runtime.

## Subdirectories

`.stult` files in subdirectories can be included in the bundle.

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

Nested manifest files are not included as project manifests. Keep one manifest
at the project root and use it to list the files that should run.

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

Bundled executables run the bundled manifest in the same way as an ordinary
manifest project.

Files run deterministically in the order specified in the manifest file.

This means a project can use separate files for setup, configuration, helpers
and main program logic:

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

Then build it with:

```bash
stult build my_tool -o my-tool
```

## Running after copying

A bundled executable can be copied to another compatible computer and run there.

For example, a Linux amd64 executable should be run on a compatible Linux amd64
system. A Windows executable should be built for Windows and will normally use
the `.exe` extension.

The user running the bundled executable does not need:

- the original project directory,
- the original `.stult` source files as separate files *or*
- Go installed.

They only need the generated executable and a compatible operating system and
CPU architecture.

Of course, this makes it really easy to share runnable code with friends and colleagues.

## Bundling is not compilation

Remember that a bundled executable is not a native-code compilation of the Stult program.

The generated executable contains the Stult runtime and the Stult source files.
At startup, the runtime loads the embedded project and interprets it.

This keeps bundling simple and flexible:

- the project can be developed as readable source files,
- the final tool can be distributed as one executable *and*
- the runtime behaviour is the same as running the manifest project directly.

To the end user, the bundled executable behaves like a normal command-line program: it can be run directly, without separately invoking `stult` or shipping the source tree.

## Development workflow

A typical workflow is:

```bash
stult my_tool
```

Then, when the project works:

```bash
stult build my_tool -o my-tool
```

Then run the bundled executable:

```bash
./my-tool
```

If you are working from the Stult source repository and have not built a local
`stult` binary yet, you can use:

```bash
go run ./src build my_tool -o my-tool
```

But once you have a Stult binary, the normal command is:

```bash
stult build my_tool -o my-tool
```

## Troubleshooting

### The project must have a manifest

`stult build` expects the project directory to contain:

```text
manifest.stulton
```

or:

```text
manifest.json
```

If neither file exists, the build fails.

### Do not use both manifest formats at once

Use either `manifest.stulton` or `manifest.json`.

Do not include both in the same project root.

### Non-Stult data files are not embedded

Only `.stult` source files and the root manifest are embedded by default.

If a program reads a file such as `data.csv`, that file must still exist at
runtime unless its contents are embedded as a string in Stult source code.

### The output executable must not overwrite the running executable

For obvious reasons, Stult refuses to overwrite the executable that is currently running.

So choose a different output path:

```bash
stult build my_tool -o my-bundled-tool
```