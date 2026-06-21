# Standard library

The Stult standard library is available through the immutable binding `STD`.

`STD` is a nested map. Its top-level entries can be accessed with dot access:

```stult
STD.ASSERT
STD.IO
STD.SYSTEM
STD.FILE
STD.TIME
STD.MATH
STD.TYPE
STD.DATA
```

Because `STD` is a nested map, the same paths can also be written with bracket indexing:

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

Dot access is syntactic sugar for string-key map access. This document keeps reference headings in bracket form so the underlying string keys are explicit, while examples use dot access where possible.

Standard-library functions are ordinary callable values stored inside maps.

Here's an example:

```stult
WRITE_LINE : STD.IO.OUTPUT.WRITE_LINE
SIZE : STD.TYPE.COLLECTION.SIZE

WRITE_LINE(
	"size: "
	SIZE({"a", "b", "c"})
)
```

All standard-library entry names are uppercase because those entries are intended to be immutable.

Some standard-library functions accept variadic arguments. In signatures, `...name` means that zero or more remaining arguments are collected into an array under that name.

## Contents

- [STD["ASSERT"]](#stdassert)
  - [`STD["ASSERT"]["TRUE"](condition, message?)`](#stdasserttruecondition-message)
  - [`STD["ASSERT"]["FALSE"](condition, message?)`](#stdassertfalsecondition-message)
  - [`STD["ASSERT"]["EQUAL"](actual, expected, message?)`](#stdassertequalactual-expected-message)
- [`STD["IO"]`](#stdio)
  - [`STD["IO"]["INPUT"]`](#stdioinput)
    - [`STD["IO"]["INPUT"]["READ_LINE"]()`](#stdioinputread_line)
    - [`STD["IO"]["INPUT"]["PROMPT"](...values)`](#stdioinputpromptvalues)
  - [`STD["IO"]["OUTPUT"]`](#stdiooutput)
    - [`STD["IO"]["OUTPUT"]["WRITE"](...values)`](#stdiooutputwritevalues)
    - [`STD["IO"]["OUTPUT"]["WRITE_LINE"](...values)`](#stdiooutputwrite_linevalues)
    - [`STD["IO"]["OUTPUT"]["WRITE_ERROR"](...values)`](#stdiooutputwrite_errorvalues)
    - [`STD["IO"]["OUTPUT"]["WRITE_ERROR_LINE"](...values)`](#stdiooutputwrite_error_linevalues)
- [`STD["SYSTEM"]`](#stdsystem)
  - [`STD["SYSTEM"]["ARGS"]`](#stdsystemargs)
  - [`STD["SYSTEM"]["CWD"]()`](#stdsystemcwd)
  - [`STD["SYSTEM"]["ENV"](name)`](#stdsystemenvname)
  - [`STD["SYSTEM"]["EXIT"](code)`](#stdsystemexitcode)
- [`STD["FILE"]`](#stdfile)
  - [`STD["FILE"]["READ"](path)`](#stdfilereadpath)
  - [`STD["FILE"]["WRITE"](path, content, append?)`](#stdfilewritepath-content-append)
  - [`STD["FILE"]["EXISTS"](path)`](#stdfileexistspath)
  - [`STD["FILE"]["LIST"](path, absolute?)`](#stdfilelistpath-absolute)
  - [`STD["FILE"]["IS_FILE"](path)`](#stdfileis_filepath)
  - [`STD["FILE"]["IS_DIR"](path)`](#stdfileis_dirpath)
  - [`STD["FILE"]["MAKE_DIR"](path, recursive?)`](#stdfilemake_dirpath-recursive)
  - [`STD["FILE"]["DELETE"](path)`](#stdfiledeletepath)
  - [`STD["FILE"]["RENAME"](old_path, new_path)`](#stdfilerenameold_path-new_path)
  - [`STD["FILE"]["COPY"](source_path, destination_path)`](#stdfilecopysource_path-destination_path)
  - [`STD["FILE"]["SIZE"](path)`](#stdfilesizepath)
  - [`STD["FILE"]["PATH"]`](#stdfilepath)
    - [`STD["FILE"]["PATH"]["JOIN"](...parts)`](#stdfilepathjoinparts)
    - [`STD["FILE"]["PATH"]["ABS"](path)`](#stdfilepathabspath)
    - [`STD["FILE"]["PATH"]["BASE"](path)`](#stdfilepathbasepath)
    - [`STD["FILE"]["PATH"]["CLEAN"](path)`](#stdfilepathcleanpath)
    - [`STD["FILE"]["PATH"]["DIR"](path)`](#stdfilepathdirpath)
    - [`STD["FILE"]["PATH"]["EXT"](path)`](#stdfilepathextpath)
- [`STD["TIME"]`](#stdtime)
  - [`STD["TIME"]["MILLI_TIMESTAMP"]()`](#stdtimemilli_timestamp)
  - [`STD["TIME"]["NANO_TIMESTAMP"]()`](#stdtimenano_timestamp)
  - [`STD["TIME"]["MILLI_SLEEP"](milliseconds)`](#stdtimemilli_sleepmilliseconds)
  - [`STD["TIME"]["LOCAL_CALENDAR"]()`](#stdtimelocal_calendar)
  - [`STD["TIME"]["UTC_CALENDAR"]()`](#stdtimeutc_calendar)
- [`STD["MATH"]`](#stdmath)
  - [`STD["MATH"]["PI"]`](#stdmathpi)
  - [`STD["MATH"]["TAU"]`](#stdmathtau)
  - [`STD["MATH"]["E"]`](#stdmathe)
  - [`STD["MATH"]["SQUARE"](number)`](#stdmathsquarenumber)
  - [`STD["MATH"]["CUBE"](number)`](#stdmathcubenumber)
  - [`STD["MATH"]["ABS"](number)`](#stdmathabsnumber)
  - [`STD["MATH"]["SIGN"](number)`](#stdmathsignnumber)
  - [`STD["MATH"]["MIN"](...numbers)`](#stdmathminnumbers)
  - [`STD["MATH"]["MAX"](...numbers)`](#stdmathmaxnumbers)
  - [`STD["MATH"]["CLAMP"](value, minimum, maximum)`](#stdmathclampvalue-minimum-maximum)
  - [`STD["MATH"]["LERP"](start, end, amount)`](#stdmathlerpstart-end-amount)
  - [`STD["MATH"]["FLOOR"](number)`](#stdmathfloornumber)
  - [`STD["MATH"]["CEIL"](number)`](#stdmathceilnumber)
  - [`STD["MATH"]["ROUND"](number)`](#stdmathroundnumber)
  - [`STD["MATH"]["TRUNC"](number)`](#stdmathtruncnumber)
  - [`STD["MATH"]["SQRT"](number)`](#stdmathsqrtnumber)
  - [`STD["MATH"]["POWER"](base, exponent)`](#stdmathpowerbase-exponent)
  - [`STD["MATH"]["MOD"](left, right)`](#stdmathmodleft-right)
  - [`STD["MATH"]["REM"](left, right)`](#stdmathremleft-right)
  - [`STD["MATH"]["RAND"]`](#stdmathrand)
    - [`STD["MATH"]["RAND"]["NUMBER"](inclusive_lower_bound, exclusive_upper_bound)`](#stdmathrandnumberinclusive_lower_bound-exclusive_upper_bound)
    - [`STD["MATH"]["RAND"]["INTEGER"](minimum, maximum)`](#stdmathrandintegerminimum-maximum)
    - [`STD["MATH"]["RAND"]["CHOICE"](collection)`](#stdmathrandchoicecollection)
    - [`STD["MATH"]["RAND"]["SHUFFLE"](collection)`](#stdmathrandshufflecollection)
  - [`STD["MATH"]["TRIG"]`](#stdmathtrig)
    - [`STD["MATH"]["TRIG"]["SIN"](radians)`](#stdmathtrigsinradians)
    - [`STD["MATH"]["TRIG"]["COS"](radians)`](#stdmathtrigcosradians)
    - [`STD["MATH"]["TRIG"]["TAN"](radians)`](#stdmathtrigtanradians)
    - [`STD["MATH"]["TRIG"]["RADIANS"](degrees)`](#stdmathtrigradiansdegrees)
    - [`STD["MATH"]["TRIG"]["DEGREES"](radians)`](#stdmathtrigdegreesradians)
- [`STD["TYPE"]`](#stdtype)
  - [`STD["TYPE"]` predicates](#stdtype-predicates)
  - [`STD["TYPE"]["BOOL"]`](#stdtypebool)
    - [`STD["TYPE"]["BOOL"]["TRUE"]`](#stdtypebooltrue)
    - [`STD["TYPE"]["BOOL"]["FALSE"]`](#stdtypeboolfalse)
    - [`STD["TYPE"]["BOOL"]["NEW"](value)`](#stdtypeboolnewvalue)
  - [`STD["TYPE"]["NUMBER"]`](#stdtypenumber)
    - [`STD["TYPE"]["NUMBER"]["DEFAULT_DECIMAL_PLACES"]`](#stdtypenumberdefault_decimal_places)
    - [`STD["TYPE"]["NUMBER"]["MAX_DECIMAL_PLACES"]`](#stdtypenumbermax_decimal_places)
    - [`STD["TYPE"]["NUMBER"]["FORMAT"](number, decimal_places)`](#stdtypenumberformatnumber-decimal_places)
    - [`STD["TYPE"]["NUMBER"]["FORMAT_SCIENTIFIC"](number, significant_digits)`](#stdtypenumberformat_scientificnumber-significant_digits)
    - [`STD["TYPE"]["NUMBER"]["NEW"](value)`](#stdtypenumbernewvalue)
  - [`STD["TYPE"]["STRING"]`](#stdtypestring)
    - [`STD["TYPE"]["STRING"]["NEW"](value)`](#stdtypestringnewvalue)
    - [`STD["TYPE"]["STRING"]["CHARACTERS"](text)`](#stdtypestringcharacterstext)
    - [`STD["TYPE"]["STRING"]["TRIM"](text)`](#stdtypestringtrimtext)
    - [`STD["TYPE"]["STRING"]["TRIM_START"](text)`](#stdtypestringtrim_starttext)
    - [`STD["TYPE"]["STRING"]["TRIM_END"](text)`](#stdtypestringtrim_endtext)
    - [`STD["TYPE"]["STRING"]["TO_LOWER"](text)`](#stdtypestringto_lowertext)
    - [`STD["TYPE"]["STRING"]["TO_UPPER"](text)`](#stdtypestringto_uppertext)
    - [`STD["TYPE"]["STRING"]["IS_FOUND_IN"](search, text)`](#stdtypestringis_found_insearch-text)
    - [`STD["TYPE"]["STRING"]["IS_FOUND_AT_START"](search, text)`](#stdtypestringis_found_at_startsearch-text)
    - [`STD["TYPE"]["STRING"]["IS_FOUND_AT_END"](search, text)`](#stdtypestringis_found_at_endsearch-text)
    - [`STD["TYPE"]["STRING"]["REPLACE"](text, old, new)`](#stdtypestringreplacetext-old-new)
    - [`STD["TYPE"]["STRING"]["SPLIT"](text, separator)`](#stdtypestringsplittext-separator)
    - [`STD["TYPE"]["STRING"]["JOIN"](array, separator)`](#stdtypestringjoinarray-separator)
  - [`STD["TYPE"]["ARRAY"]`](#stdtypearray)
    - [`STD["TYPE"]["ARRAY"]["APPEND"](array, ...values)`](#stdtypearrayappendarray-values)
  - [`STD["TYPE"]["MAP"]`](#stdtypemap)
    - [`STD["TYPE"]["MAP"]["KEYS"](map)`](#stdtypemapkeysmap)
    - [`STD["TYPE"]["MAP"]["VALUES"](map)`](#stdtypemapvaluesmap)
  - [`STD["TYPE"]["COLLECTION"]`](#stdtypecollection)
    - [`STD["TYPE"]["COLLECTION"]["SIZE"](collection)`](#stdtypecollectionsizecollection)
    - [`STD["TYPE"]["COLLECTION"]["IS_EMPTY"](collection)`](#stdtypecollectionis_emptycollection)
    - [`STD["TYPE"]["COLLECTION"]["HAS"](collection, key)`](#stdtypecollectionhascollection-key)
    - [`STD["TYPE"]["COLLECTION"]["GET"](collection, key_or_index, default?)`](#stdtypecollectiongetcollection-key_or_index-default)
    - [`STD["TYPE"]["COLLECTION"]["CLEAR"](collection)`](#stdtypecollectionclearcollection)
    - [`STD["TYPE"]["COLLECTION"]["CLONE"](value)`](#stdtypecollectionclonevalue)
    - [`STD["TYPE"]["COLLECTION"]["FREEZE"](collection)`](#stdtypecollectionfreezecollection)
    - [`STD["TYPE"]["COLLECTION"]["IS_FROZEN"](value)`](#stdtypecollectionis_frozenvalue)
- [`STD["DATA"]`](#stddata)
  - [`STD["DATA"]["CSV"]`](#stddatacsv)
    - [`STD["DATA"]["CSV"]["NEW"](rows)`](#stddatacsvnewrows)
    - [`STD["DATA"]["CSV"]["PARSE"](text)`](#stddatacsvparsetext)
    - [`STD["DATA"]["CSV"]["IS_VALID"](text)`](#stddatacsvis_validtext)
  - [`STD["DATA"]["JSON"]`](#stddatajson)
    - [`STD["DATA"]["JSON"]["NEW"](value)`](#stddatajsonnewvalue)
    - [`STD["DATA"]["JSON"]["PARSE"](text)`](#stddatajsonparsetext)
    - [`STD["DATA"]["JSON"]["IS_VALID"](text)`](#stddatajsonis_validtext)
  - [`STD["DATA"]["STULTON"]`](#stddatastulton)
    - [`STD["DATA"]["STULTON"]["NEW"](value)`](#stddatastultonnewvalue)
    - [`STD["DATA"]["STULTON"]["PARSE"](text)`](#stddatastultonparsetext)
    - [`STD["DATA"]["STULTON"]["IS_VALID"](text)`](#stddatastultonis_validtext)

### `STD["ASSERT"]`

Runtime assertion helpers.

These helpers are useful for self-checking scripts, examples and test programs.

Failed assertions raise runtime errors.

### `STD["ASSERT"]["TRUE"](condition, message?)`

Checks that `condition` is true.

`condition` must be a boolean.

`message`, if supplied, must be a string.

Returns `_` if `condition` is true.

Raises a runtime error if `condition` is false.

```stult
STD.ASSERT.TRUE(1 + 1 = 2)
STD.ASSERT.TRUE(1 + 1 = 2, "arithmetic should work")
```

### `STD["ASSERT"]["FALSE"](condition, message?)`

Checks that `condition` is false.

`condition` must be a boolean.

`message`, if supplied, must be a string.

Returns `_` if `condition` is false.

Raises a runtime error if `condition` is true.

```stult
STD.ASSERT.FALSE(10 < 5)
STD.ASSERT.FALSE(10 < 5, "10 should not be less than 5")
```

### `STD["ASSERT"]["EQUAL"](actual, expected, message?)`

Checks that `actual` and `expected` are equal.

`message`, if supplied, must be a string.

Comparison uses normal Stult equality semantics.

Returns `_` if `actual` and `expected` are equal.

Raises a runtime error if they are not equal.

```stult
STD.ASSERT.EQUAL("Stult", "Stult")
STD.ASSERT.EQUAL(1 + 2, 3, "numbers should match")
```

## `STD["IO"]`

Console input and output.

`STD["IO"]` contains two submaps:

```stult
STD.IO.INPUT
STD.IO.OUTPUT
```

`STD.IO.INPUT` contains functions that read from standard input.

`STD.IO.OUTPUT` contains functions that write to standard output or standard error.

### `STD["IO"]["INPUT"]`

Console input helpers.

### `STD["IO"]["INPUT"]["READ_LINE"]()`

Reads one line from standard input.

```stult
line : STD.IO.INPUT.READ_LINE()
```

Returns a string without the trailing newline.

Returns `_` if input ends before any text is read.

### `STD["IO"]["INPUT"]["PROMPT"](...values)`

Writes prompt values to standard output, then reads a line from standard input.

```stult
name : STD.IO.INPUT.PROMPT("Name: ")
STD.IO.OUTPUT.WRITE_LINE("Hello ", name)
```

Returns a string without the trailing newline.

Returns `_` if input ends before any text is read.

### `STD["IO"]["OUTPUT"]`

Console output helpers.

### `STD["IO"]["OUTPUT"]["WRITE"](...values)`

Writes values to standard output without adding a newline.

```stult
STD.IO.OUTPUT.WRITE("Hello")
STD.IO.OUTPUT.WRITE(" ")
STD.IO.OUTPUT.WRITE("world")
```

Returns `_`.

### `STD["IO"]["OUTPUT"]["WRITE_LINE"](...values)`

Writes values to standard output and then writes a newline.

```stult
STD.IO.OUTPUT.WRITE_LINE("Hello world")
STD.IO.OUTPUT.WRITE_LINE("Count: ", 3)
```

Returns `_`.

### `STD["IO"]["OUTPUT"]["WRITE_ERROR"](...values)`

Writes values to standard error without adding a newline.

```stult
STD.IO.OUTPUT.WRITE_ERROR("Warning: ")
STD.IO.OUTPUT.WRITE_ERROR("something changed")
```

Returns `_`.

### `STD["IO"]["OUTPUT"]["WRITE_ERROR_LINE"](...values)`

Writes values to standard error and then writes a newline.

```stult
STD.IO.OUTPUT.WRITE_ERROR_LINE("Something went wrong")
```

Returns `_`.

## `STD["SYSTEM"]`

Runtime and process-context helpers.

### `STD["SYSTEM"]["ARGS"]`

Contains the arguments passed to the Stult program.

```stult
STD.IO.OUTPUT.WRITE_LINE(STD.SYSTEM.ARGS)
```

The array does not include the Stult executable path.

The array also does not include the source file, manifest file or project directory used as the program target.

For example:

```bash
stult run examples/csv_to_json_converter.stult input.csv output.json
```

makes this available to Stult code:

```stult
STD.SYSTEM.ARGS # {"input.csv", "output.json"}
```

`ARGS` is an immutable array of strings.

### `STD["SYSTEM"]["CWD"]()`

Returns the current working directory.

```stult
cwd : STD.SYSTEM.CWD()

STD.IO.OUTPUT.WRITE_LINE(cwd)
```

Returns a string.

### `STD["SYSTEM"]["ENV"](name)`

Reads an environment variable.

```stult
home : STD.SYSTEM.ENV("HOME")

STD.IO.OUTPUT.WRITE_LINE(home)
```

The argument must be a string.

Returns a string when the environment variable is set.

Returns `_` when the environment variable is not set.

### `STD["SYSTEM"]["EXIT"](code)`

Exits the current process with the given status code.

```stult
STD.SYSTEM.EXIT(0)
```

The exit code must be an integer number from `0` to `255`.

This function does not return, because it terminates the process.

## `STD["FILE"]`

File-system helpers.

### `STD["FILE"]["READ"](path)`

Reads a file as text.

```stult
text : STD.FILE.READ("notes.txt")
```

Returns a string.

### `STD["FILE"]["WRITE"](path, content, append?)`

Writes content to a file.

```stult
STD.FILE.WRITE("notes.txt", "Hello")
STD.FILE.WRITE("notes.txt", "\nAnother line", +)
```

The path must be a string.

The content may be a string or another Stult value. Non-string values are converted with their printed representation.

The optional `append` argument must be a boolean and defaults to `-`.

When `append` is `-`, replaces existing file contents.

When `append` is `+`, appends to the file and creates the file if it does not exist.

Returns `_`.

### `STD["FILE"]["EXISTS"](path)`

Checks whether a file-system path exists.

```stult
(STD.FILE.EXISTS("notes.txt")) {
	STD.IO.OUTPUT.WRITE_LINE("notes.txt exists")
}
```

Returns a boolean.

### `STD["FILE"]["LIST"](path, absolute?)`

Lists the direct entries inside a directory.

```stult
names : STD.FILE.LIST("docs")
absolute_paths : STD.FILE.LIST("docs", +)
```

The path must be a string.

The optional `absolute` argument must be a boolean and defaults to `-`.

When `absolute` is `-`, returns sorted entry names.

When `absolute` is `+`, returns sorted absolute paths for the same entries.

This function is non-recursive.

Returns an array of strings.

### `STD["FILE"]["IS_FILE"](path)`

Checks whether a path exists and is a regular file.

```stult
STD.FILE.IS_FILE("notes.txt")
```

The path must be a string.

Returns a boolean.

Returns `-` when the path does not exist or exists but is not a regular file.

### `STD["FILE"]["IS_DIR"](path)`

Checks whether a path exists and is a directory.

```stult
STD.FILE.IS_DIR("docs")
```

The path must be a string.

Returns a boolean.

Returns `-` when the path does not exist or exists but is not a directory.

### `STD["FILE"]["MAKE_DIR"](path, recursive?)`

Creates a directory.

```stult
STD.FILE.MAKE_DIR("out")
STD.FILE.MAKE_DIR("out/reports/2026", +)
```

The path must be a string.

The optional `recursive` argument must be a boolean and defaults to `-`.

When `recursive` is `-`, creates exactly one directory. The parent directory must already exist, and it is an error if the path already exists.

When `recursive` is `+`, creates the directory and any missing parent directories. It succeeds if the directory already exists, and errors if the path exists but is not a directory.

Returns `_`.

### `STD["FILE"]["DELETE"](path)`

Deletes a file-system path.

```stult
STD.FILE.DELETE("notes.txt")
```

Returns `_`.

### `STD["FILE"]["RENAME"](old_path, new_path)`

Renames or moves a file-system path.

```stult
STD.FILE.RENAME("old.txt", "new.txt")
```

Returns `_`.

### `STD["FILE"]["COPY"](source_path, destination_path)`

Copies a file.

```stult
STD.FILE.COPY("source.txt", "copy.txt")
```

The destination is created or replaced.

Returns `_`.

### `STD["FILE"]["SIZE"](path)`

Returns the size of a file-system path in bytes.

```stult
size : STD.FILE.SIZE("notes.txt")
```

Returns a number.

### `STD["FILE"]["PATH"]`

File-path manipulation helpers.

#### `STD["FILE"]["PATH"]["JOIN"](...parts)`

Joins path parts into one path.

```stult
path : STD.FILE.PATH.JOIN("data", "input.csv")

STD.IO.OUTPUT.WRITE_LINE(path)
```

Requires at least one argument.

Each argument must be a string.

Returns a string.

#### `STD["FILE"]["PATH"]["ABS"](path)`

Returns an absolute version of a path.

```stult
absolute_path : STD.FILE.PATH.ABS("input.csv")

STD.IO.OUTPUT.WRITE_LINE(absolute_path)
```

The argument must be a string.

Returns a string.

#### `STD["FILE"]["PATH"]["BASE"](path)`

Returns the final element of a path.

```stult
name : STD.FILE.PATH.BASE("data/input.csv")

STD.IO.OUTPUT.WRITE_LINE(name)
```

The argument must be a string.

Returns a string.

#### `STD["FILE"]["PATH"]["CLEAN"](path)`

Cleans a path by removing redundant path elements.

```stult
clean_path : STD.FILE.PATH.CLEAN("./data/../data/input.csv")

STD.IO.OUTPUT.WRITE_LINE(clean_path)
```

The argument must be a string.

Returns a string.

#### `STD["FILE"]["PATH"]["DIR"](path)`

Returns the directory part of a path.

```stult
dir : STD.FILE.PATH.DIR("data/input.csv")

STD.IO.OUTPUT.WRITE_LINE(dir)
```

The argument must be a string.

Returns a string.

#### `STD["FILE"]["PATH"]["EXT"](path)`

Returns the file extension of a path.

```stult
extension : STD.FILE.PATH.EXT("data/input.csv")

STD.IO.OUTPUT.WRITE_LINE(extension)
```

The argument must be a string.

Returns a string.

## `STD["TIME"]`

Timestamps, sleep and calendar snapshots.

### `STD["TIME"]["MILLI_TIMESTAMP"]()`

Returns the current Unix timestamp in milliseconds.

```stult
start : STD.TIME.MILLI_TIMESTAMP()
```

Returns a number.

### `STD["TIME"]["NANO_TIMESTAMP"]()`

Returns the current Unix timestamp in nanoseconds.

```stult
start : STD.TIME.NANO_TIMESTAMP()
```

Returns a number.

### `STD["TIME"]["MILLI_SLEEP"](milliseconds)`

Sleeps for the given number of milliseconds.

```stult
STD.TIME.MILLI_SLEEP(500)
```

The argument must be a non-negative integer number.

Returns `_`.

### `STD["TIME"]["LOCAL_CALENDAR"]()`

Returns a map describing the current local time.

```stult
now : STD.TIME.LOCAL_CALENDAR()

STD.IO.OUTPUT.WRITE_LINE(now.YEAR, "-", now.MONTH, "-", now.DAY)
```

The returned map contains:

```text
YEAR
MONTH
DAY
HOUR
MINUTE
SECOND
NANOSECOND
WEEKDAY
YEARDAY
ZONE
OFFSET
```

### `STD["TIME"]["UTC_CALENDAR"]()`

Returns a map describing the current UTC time and date.

```stult
utc : STD.TIME.UTC_CALENDAR()
```

The returned map has the same keys as `LOCAL_CALENDAR`.

## `STD["MATH"]`

Numeric helpers, constants and trigonometry.

### `STD["MATH"]["PI"]`

The mathematical constant pi.

```stult
STD.MATH.PI
```

### `STD["MATH"]["TAU"]`

The mathematical constant tau.

```stult
STD.MATH.TAU
```

`TAU` is `PI * 2`.

### `STD["MATH"]["E"]`

The mathematical constant e.

```stult
STD.MATH.E
```

### `STD["MATH"]["SQUARE"](number)`

Returns `number * number`.

```stult
STD.MATH.SQUARE(9)
```

### `STD["MATH"]["CUBE"](number)`

Returns `number * number * number`.

```stult
STD.MATH.CUBE(3)
```

### `STD["MATH"]["ABS"](number)`

Returns the absolute value.

```stult
STD.MATH.ABS(-12.5)
```

### `STD["MATH"]["SIGN"](number)`

Returns:

```text
-1 for negative numbers
0 for zero
1 for positive numbers
```

```stult
STD.MATH.SIGN(-12.5)
```

### `STD["MATH"]["MIN"](...numbers)`

Returns the smallest number.

```stult
STD.MATH.MIN(8, -3, 10, 2)
```

Requires at least one argument.

### `STD["MATH"]["MAX"](...numbers)`

Returns the largest number.

```stult
STD.MATH.MAX(8, -3, 10, 2)
```

Requires at least one argument.

### `STD["MATH"]["CLAMP"](value, minimum, maximum)`

Restricts a number to a range.

```stult
STD.MATH.CLAMP(15, 0, 10)
```

Returns `minimum` if `value` is below `minimum`.

Returns `maximum` if `value` is above `maximum`.

Returns `value` otherwise.

### `STD["MATH"]["LERP"](start, end, amount)`

Linearly interpolates between `start` and `end`.

```stult
STD.MATH.LERP(0, 10, 0.5)
```

### `STD["MATH"]["FLOOR"](number)`

Rounds down to an integer.

```stult
STD.MATH.FLOOR(3.9)
```

### `STD["MATH"]["CEIL"](number)`

Rounds up to an integer.

```stult
STD.MATH.CEIL(3.1)
```

### `STD["MATH"]["ROUND"](number)`

Rounds to the nearest integer.

```stult
STD.MATH.ROUND(3.5)
```

Positive halves round upward.

Negative halves round downward.

### `STD["MATH"]["TRUNC"](number)`

Removes the fractional part.

```stult
STD.MATH.TRUNC(-3.9)
```

### `STD["MATH"]["SQRT"](number)`

Returns the square root.

```stult
STD.MATH.SQRT(2)
```

The argument must be non-negative.

### `STD["MATH"]["POWER"](base, exponent)`

Raises `base` to `exponent`.

```stult
STD.MATH.POWER(2, 10)
STD.MATH.POWER(2, 0.5)
```

Zero cannot be raised to a negative exponent.

Negative bases with non-integer exponents are not allowed.

### `STD["MATH"]["MOD"](left, right)`

Returns the mathematical modulo.

```stult
STD.MATH.MOD(10, 3)
STD.MATH.MOD(-10, 3)
```

The divisor cannot be zero.

For a positive divisor, the result is greater than or equal to zero and less than the divisor.

### `STD["MATH"]["REM"](left, right)`

Returns the truncating remainder.

```stult
STD.MATH.REM(10, 3)
STD.MATH.REM(-10, 3)
```

The divisor cannot be zero.

The result follows the sign of the dividend, or is zero. This matches the remainder behaviour of the `%` operator in C-family programming languages.

## `STD["MATH"]["RAND"]`

Random helpers.

These functions use the host cryptographic random source.

### `STD["MATH"]["RAND"]["NUMBER"](inclusive_lower_bound, exclusive_upper_bound)`

Returns a random number.

```stult
STD.MATH.RAND.NUMBER(0, 1)
STD.MATH.RAND.NUMBER(-10, 10)
```

The result is greater than or equal to `inclusive_lower_bound` and less than `exclusive_upper_bound`.

```text
inclusive_lower_bound <= result < exclusive_upper_bound
```

Both arguments must be numbers.

The inclusive lower bound must be less than the exclusive upper bound.

The returned number may include a decimal part. Stult may generate decimal places up to the implementation's maximum decimal-place limit.

### `STD["MATH"]["RAND"]["INTEGER"](minimum, maximum)`

Returns a random integer.

```stult
STD.MATH.RAND.INTEGER(1, 6)
STD.MATH.RAND.INTEGER(-10, 10)
```

The result is greater than or equal to `minimum` and less than or equal to `maximum`.

```text
minimum <= result <= maximum
```

Both arguments must be integer numbers.

The minimum must be less than or equal to the maximum.

Large integer bounds are supported, subject to available memory and processing time.

### `STD["MATH"]["RAND"]["CHOICE"](collection)`

Returns one randomly chosen value from a collection.

```stult
STD.MATH.RAND.CHOICE({"red", "green", "blue"})
STD.MATH.RAND.CHOICE("abc")
STD.MATH.RAND.CHOICE({"dog": 50, "cat": 80})
```

For arrays, the result is one array element.

For strings, the result is a one-character string.

For maps, the result is one map value.

The collection must not be empty.

### `STD["MATH"]["RAND"]["SHUFFLE"](collection)`

Returns a shuffled copy of a collection.

```stult
STD.MATH.RAND.SHUFFLE({1, 2, 3})
STD.MATH.RAND.SHUFFLE("abc")
STD.MATH.RAND.SHUFFLE({"dog": 50, "cat": 80, "chicken": 90})
```

For arrays, the result is a new array containing the same values in random order.

For strings, the result is a new string containing the same characters in random order.

For maps, the result is a new map with the same keys and entry mutability, but with the existing values randomly reassigned among those keys.

```stult
scores : {
	"dog": 50
	"cat": 80
	"chicken": 90
}

STD.MATH.RAND.SHUFFLE(scores)
# might return:
# {
#     "dog": 90
#     "cat": 50
#     "chicken": 80
# }
```

The original collection is not modified.

## `STD["MATH"]["TRIG"]`

Trigonometric functions use radians.

### `STD["MATH"]["TRIG"]["SIN"](radians)`

Returns the sine of an angle.

```stult
STD.MATH.TRIG.SIN(STD.MATH.PI / 2)
```

### `STD["MATH"]["TRIG"]["COS"](radians)`

Returns the cosine of an angle.

```stult
STD.MATH.TRIG.COS(0)
```

### `STD["MATH"]["TRIG"]["TAN"](radians)`

Returns the tangent of an angle.

```stult
STD.MATH.TRIG.TAN(STD.MATH.PI / 4)
```

Tangent is not defined where cosine is zero.

### `STD["MATH"]["TRIG"]["RADIANS"](degrees)`

Converts degrees to radians.

```stult
STD.MATH.TRIG.RADIANS(180)
```

### `STD["MATH"]["TRIG"]["DEGREES"](radians)`

Converts radians to degrees.

```stult
STD.MATH.TRIG.DEGREES(STD.MATH.PI)
```

## `STD["TYPE"]`

Type predicates, conversions and collection helpers.

## `STD["TYPE"]` predicates

Each predicate accepts one value and returns a boolean.

```stult
STD.TYPE.IS_VOID(_)
STD.TYPE.IS_BOOL(+)
STD.TYPE.IS_NUMBER(123)
STD.TYPE.IS_STRING("hello")
STD.TYPE.IS_ARRAY({})
STD.TYPE.IS_MAP({:})
STD.TYPE.IS_FUNCTION({ () (_) })
STD.TYPE.IS_BUILTIN_FUNCTION(STD.IO.OUTPUT.WRITE_LINE)
STD.TYPE.IS_COLLECTION({"a", "b"})
```

`STD["TYPE"]["IS_COLLECTION"]` returns true for arrays, maps and strings.

## `STD["TYPE"]["BOOL"]`

Boolean constants and conversion helpers.

### `STD["TYPE"]["BOOL"]["TRUE"]`

The standard-library boolean true constant.

```stult
STD.TYPE.BOOL.TRUE
```

This value is equivalent to the boolean literal `+`.

### `STD["TYPE"]["BOOL"]["FALSE"]`

The standard-library boolean false constant.

```stult
STD.TYPE.BOOL.FALSE
```

This value is equivalent to the boolean literal `-`.

### `STD["TYPE"]["BOOL"]["NEW"](value)`

Converts a value to a boolean when possible.

```stult
STD.TYPE.BOOL.NEW("true")
STD.TYPE.BOOL.NEW("false")
STD.TYPE.BOOL.NEW(1)
STD.TYPE.BOOL.NEW(0)
```

Conversion rules:

```text
bool      returns itself
number    false if zero, true otherwise
string    "T", "TRUE" and "1" become true
string    "F", "FALSE" and "0" become false
string    "+" becomes true
string    "-" becomes false
other     returns _
```

String conversion ignores surrounding whitespace and case.

## `STD["TYPE"]["NUMBER"]`

Number constants, formatting helpers and conversion helpers.

### `STD["TYPE"]["NUMBER"]["DEFAULT_DECIMAL_PLACES"]`

The number of decimal places used by ordinary number display.

```stult
STD.TYPE.NUMBER.DEFAULT_DECIMAL_PLACES
```

This controls display only. It does not limit how many decimal places Stult can store internally.

### `STD["TYPE"]["NUMBER"]["MAX_DECIMAL_PLACES"]`

The maximum number of digits Stult stores after the decimal point.

```stult
STD.TYPE.NUMBER.MAX_DECIMAL_PLACES
```

This limit applies to the decimal part of a number, not to the whole-number part. Whole-number values are theoretically unbounded, subject to available memory and processing time.

### `STD["TYPE"]["NUMBER"]["FORMAT"](number, decimal_places)`

Formats `number` as a fixed decimal string.

```stult
NUMBER : STD.TYPE.NUMBER

NUMBER.FORMAT(1 / 3, 32)
# "0.33333333333333333333333333333333"

NUMBER.FORMAT(10, 256)
# "10"
```

`decimal_places` must be an integer from `0` to `STD["TYPE"]["NUMBER"]["MAX_DECIMAL_PLACES"]`.

The result uses up to the requested number of decimal places and trims trailing zeroes.

### `STD["TYPE"]["NUMBER"]["FORMAT_SCIENTIFIC"](number, significant_digits)`

Formats `number` as a scientific-notation string.

```stult
NUMBER : STD.TYPE.NUMBER

NUMBER.FORMAT_SCIENTIFIC(12345, 3)
# "1.23e+4"
```

`significant_digits` must be an integer from `1` to `STD["TYPE"]["NUMBER"]["MAX_DECIMAL_PLACES"]`.

The result is rounded to the requested number of significant digits.

### `STD["TYPE"]["NUMBER"]["NEW"](value)`

Converts a value to a number when possible.

```stult
STD.TYPE.NUMBER.NEW("123.45")
STD.TYPE.NUMBER.NEW("75%")
STD.TYPE.NUMBER.NEW("+")  # returns _; boolean text is not a number
STD.TYPE.NUMBER.NEW(+)    # returns 1
```

Conversion rules:

```text
number    returns a cloned number
bool      true becomes 1, false becomes 0
string    parsed as a number after trimming whitespace; percentage strings such as "75%" are accepted
other     returns _
```

Invalid number strings return `_`.

## `STD["TYPE"]["STRING"]`

String conversion and string utility functions.

### `STD["TYPE"]["STRING"]["NEW"](value)`

Converts a value to a string.

```stult
STD.TYPE.STRING.NEW(123)
STD.TYPE.STRING.NEW(+)
STD.TYPE.STRING.NEW({"a", "b"})
```

### `STD["TYPE"]["STRING"]["CHARACTERS"](text)`

Returns an array of one-character strings.

```stult
STD.TYPE.STRING.CHARACTERS("cat")
```

Result:

```stult
{"c", "a", "t"}
```

### `STD["TYPE"]["STRING"]["TRIM"](text)`

Removes leading and trailing whitespace.

```stult
STD.TYPE.STRING.TRIM("  hello  ")
```

### `STD["TYPE"]["STRING"]["TRIM_START"](text)`

Removes leading whitespace.

```stult
STD.TYPE.STRING.TRIM_START("  hello")
```

### `STD["TYPE"]["STRING"]["TRIM_END"](text)`

Removes trailing whitespace.

```stult
STD.TYPE.STRING.TRIM_END("hello  ")
```

### `STD["TYPE"]["STRING"]["TO_LOWER"](text)`

Converts text to lowercase.

```stult
STD.TYPE.STRING.TO_LOWER("Hello")
```

### `STD["TYPE"]["STRING"]["TO_UPPER"](text)`

Converts text to uppercase.

```stult
STD.TYPE.STRING.TO_UPPER("Hello")
```

### `STD["TYPE"]["STRING"]["IS_FOUND_IN"](search, text)`

Checks whether `search` appears inside `text`.

```stult
STD.TYPE.STRING.IS_FOUND_IN("ell", "hello")
```

### `STD["TYPE"]["STRING"]["IS_FOUND_AT_START"](search, text)`

Checks whether `text` starts with `search`.

```stult
STD.TYPE.STRING.IS_FOUND_AT_START("he", "hello")
```

### `STD["TYPE"]["STRING"]["IS_FOUND_AT_END"](search, text)`

Checks whether `text` ends with `search`.

```stult
STD.TYPE.STRING.IS_FOUND_AT_END("lo", "hello")
```

### `STD["TYPE"]["STRING"]["REPLACE"](text, old, new)`

Replaces every occurrence of `old` with `new`.

```stult
STD.TYPE.STRING.REPLACE("one two one", "one", "three")
```

### `STD["TYPE"]["STRING"]["SPLIT"](text, separator)`

Splits text into an array of strings.

```stult
STD.TYPE.STRING.SPLIT("a,b,c", ",")
```

### `STD["TYPE"]["STRING"]["JOIN"](array, separator)`

Joins array elements into a string.

```stult
STD.TYPE.STRING.JOIN({"a", "b", "c"}, ",")
```

String elements are used directly.

Non-string elements are converted with their printed representation.

## `STD["TYPE"]["ARRAY"]`

Array-specific helpers.

### `STD["TYPE"]["ARRAY"]["APPEND"](array, ...values)`

Appends one or more values to an array.

```stult
items : {}

STD.TYPE.ARRAY.APPEND(items, "a", "b", "c")
```

Returns `_`.

The first argument must be an array.

Raises a runtime error if the array is frozen.

## `STD["TYPE"]["MAP"]`

Map-specific helpers.

### `STD["TYPE"]["MAP"]["KEYS"](map)`

Returns an array of map keys sorted lexicographically.

```stult
keys : STD.TYPE.MAP.KEYS({"b": 2, "a": 1})
```

Returns `_` when the value is not a map.

### `STD["TYPE"]["MAP"]["VALUES"](map)`

Returns an array of map values sorted by key order.

```stult
values : STD.TYPE.MAP.VALUES({"b": 2, "a": 1})
```

Returns `_` when the value is not a map.

## `STD["TYPE"]["COLLECTION"]`

Helpers shared by arrays, maps and strings.

### `STD["TYPE"]["COLLECTION"]["SIZE"](collection)`

Returns the size of a collection.

```stult
STD.TYPE.COLLECTION.SIZE({"a", "b", "c"})
STD.TYPE.COLLECTION.SIZE({"name": "example"})
STD.TYPE.COLLECTION.SIZE("hello")
```

For arrays, size is the number of elements.

For maps, size is the number of entries.

For strings, size is the number of runes.

Returns `_` for non-collections.

### `STD["TYPE"]["COLLECTION"]["IS_EMPTY"](collection)`

Checks whether a collection is empty.

```stult
STD.TYPE.COLLECTION.IS_EMPTY({})
STD.TYPE.COLLECTION.IS_EMPTY("")
```

Returns `_` for non-collections.

### `STD["TYPE"]["COLLECTION"]["HAS"](collection, key)`

Checks whether a collection contains a key or index.

```stult
STD.TYPE.COLLECTION.HAS({"name": "example"}, "name")
STD.TYPE.COLLECTION.HAS({"a", "b"}, 1)
STD.TYPE.COLLECTION.HAS("cat", 0)
```

For maps, the key must be a string.

For arrays and strings, the key must be a valid numeric index.

Returns `_` for non-collections.

### `STD["TYPE"]["COLLECTION"]["GET"](collection, key_or_index, default?)`

Safely retrieves an item from a collection.

```stult
GET : STD.TYPE.COLLECTION.GET

GET({"name": "example"}, "name")
GET({"a", "b"}, 1)
GET("cat", 0)
```

For maps, `key_or_index` must be a string key. If the key exists, `GET` returns the corresponding map entry value. If the key is missing, `GET` returns the optional default value, or `_` when no default value was supplied.

```stult
config : {"timeout": 1000}

STD.TYPE.COLLECTION.GET(config, "timeout", 500) # 1000
STD.TYPE.COLLECTION.GET(config, "retries", 3)   # 3
STD.TYPE.COLLECTION.GET(config, "missing")      # _
```

For arrays, `key_or_index` must be an exact integer number. If the index is in bounds, `GET` returns the array element. If the index is negative or out of bounds for the dense array, `GET` returns the optional default value, or `_` when no default value was supplied. Array indices are checked against the array's length.

```stult
items : {"a", "b"}

STD.TYPE.COLLECTION.GET(items, 1, "missing")  # b
STD.TYPE.COLLECTION.GET(items, 30, "missing") # missing
STD.TYPE.COLLECTION.GET(items, -1)             # _
```

For strings, `key_or_index` must be an exact integer number. If the index is in bounds, `GET` returns a new one-rune string. If the index is negative or out of bounds, `GET` returns the optional default value, or `_` when no default value was supplied. String indices are checked against the string's code-point length.

```stult
STD.TYPE.COLLECTION.GET("cat", 1, "missing")  # a
STD.TYPE.COLLECTION.GET("cat", 10, "missing") # missing
```

`GET` raises a runtime error when the first argument is not a collection, when a map key is not a string, or when an array/string index is not an exact integer number.

The optional default is an ordinary argument, so it is evaluated before `GET` is called.

### `STD["TYPE"]["COLLECTION"]["CLEAR"](collection)`

Removes all contents from a non-frozen collection.

```stult
items : {"a", "b"}
STD.TYPE.COLLECTION.CLEAR(items)
```

For arrays, this removes all elements.

For maps, this removes all entries.

For strings, this removes all characters.

Returns `_`.

Raises a runtime error if the collection is frozen.


### `STD["TYPE"]["COLLECTION"]["CLONE"](value)`

Deeply clones a value.

```stult
original : {
	"nested": {"value": 1}
}

copy : STD.TYPE.COLLECTION.CLONE(original)
copy.nested.value : 2

STD.IO.OUTPUT.WRITE_LINE(original.nested.value) # 1
STD.IO.OUTPUT.WRITE_LINE(copy.nested.value)     # 2
```

For arrays, maps and strings, `CLONE` returns new mutable collection values. Nested arrays, maps and strings are cloned recursively.

Internal aliases are preserved inside the cloned graph.

```stult
shared : {"value": 1}
original : {
	"a": shared
	"b": shared
}

copy : STD.TYPE.COLLECTION.CLONE(original)
copy.a.value : 9

STD.IO.OUTPUT.WRITE_LINE(original.a.value) # 1
STD.IO.OUTPUT.WRITE_LINE(copy.b.value)     # 9
```

Cyclical collection graphs are preserved.

```stult
array : {}
array[0] : array

copy : STD.TYPE.COLLECTION.CLONE(array)
copy[1] : "copy only"

STD.IO.OUTPUT.WRITE_LINE(array) # {<cyclical array>}
STD.IO.OUTPUT.WRITE_LINE(copy)  # {<cyclical array>, "copy only"}
```

`CLONE` returns mutable collections even when the original collections are frozen. Map-entry mutability is preserved: entries whose keys are immutable-form in the source map remain immutable entries in the clone.

Numbers are copied defensively. Booleans, void, functions and builtin functions are reused. Reusing function values means cloned collections containing closures still refer to the same closure state.

### `STD["TYPE"]["COLLECTION"]["FREEZE"](collection)`

Deeply freezes a collection.

```stult
CONFIG : STD.TYPE.COLLECTION.FREEZE({
	"name": "demo"
	"values": {1, 2, 3}
})
```

`FREEZE` accepts arrays, maps and strings.

A frozen array cannot have elements replaced or appended.

A frozen map cannot have entries replaced or added.

A frozen string cannot have characters replaced or appended.

The freeze is deep. Nested arrays, maps and strings inside the collection are frozen too.

`FREEZE` modifies the collection in place and returns the same collection value. This means it can be used either as a statement or inside an assignment.

```stult
items : {1, 2, 3}
STD.TYPE.COLLECTION.FREEZE(items)

other_items : STD.TYPE.COLLECTION.FREEZE({4, 5, 6})
```

Returns `_` for non-collections.

### `STD["TYPE"]["COLLECTION"]["IS_FROZEN"](value)`

Checks whether a value is a frozen collection.

```stult
items : STD.TYPE.COLLECTION.FREEZE({1, 2, 3})

STD.TYPE.COLLECTION.IS_FROZEN(items)
STD.TYPE.COLLECTION.IS_FROZEN(items[0])
```

Returns `+` for frozen arrays, maps and strings.

Returns `-` for non-frozen arrays, maps and strings.

Returns `-` for non-collections.

## `STD["DATA"]`

Data encoding and decoding helpers.

## `STD["DATA"]["CSV"]`

CSV encoding, parsing and validation helpers.

### `STD["DATA"]["CSV"]["NEW"](rows)`

Encodes an array of row arrays as CSV text.

```stult
rows : {
	{"name", "score"}
	{"a", 10}
	{"b", 20}
}

text : STD.DATA.CSV.NEW(rows)
```

Fields are converted to strings when possible.

Returns a string.

### `STD["DATA"]["CSV"]["PARSE"](text)`

Parses CSV text into an array of row arrays.

```stult
rows : STD.DATA.CSV.PARSE("name,score\na,10\nb,20\n")
```

Returns an array of arrays of strings.

### `STD["DATA"]["CSV"]["IS_VALID"](text)`

Checks whether text is valid CSV.

```stult
STD.DATA.CSV.IS_VALID("name,score\na,10\n")
```

Returns a boolean.

Rows may have different field counts.

## `STD["DATA"]["JSON"]`

JSON encoding, parsing and validation helpers.

### `STD["DATA"]["JSON"]["NEW"](value)`

Encodes a Stult value as JSON text.

```stult
text : STD.DATA.JSON.NEW({
	"name": "example"
	"active": +
})
```

Conversion rules:

```text
_         null
bool      boolean
number    number
string    string
array     array
map       object
function  error
```

Returns a string.

### `STD["DATA"]["JSON"]["PARSE"](text)`

Parses JSON text into a Stult value.

```stult
value : STD.DATA.JSON.PARSE("{\"name\":\"example\",\"active\":true}")
```

Conversion rules:

```text
null      _
boolean   bool
number    number
string    string
array     array
object    map
```

The input must contain exactly one JSON value.

### `STD["DATA"]["JSON"]["IS_VALID"](text)`

Checks whether text is valid JSON.

```stult
STD.DATA.JSON.IS_VALID("{\"ok\":true}")
```

Returns a boolean.

## `STD["DATA"]["STULTON"]`

STULTON encoding, parsing and validation helpers.

### `STD["DATA"]["STULTON"]["NEW"](value)`

Encodes a Stult value as STULTON text.

```stult
text : STD.DATA.STULTON.NEW({
	"NAME": "example"
	"active": +
	"items": {
		"one"
		"two"
	}
})
```

Conversion rules:

```text
_         _
bool      + or -
number    number
string    quoted string
array     STULTON array
map       STULTON map
function  error
```

Map keys are written in sorted order.

Returns a string.

### `STD["DATA"]["STULTON"]["PARSE"](text)`

Parses STULTON text into a Stult value.

```stult
value : STD.DATA.STULTON.PARSE("{\"active\": +}")
```

STULTON parsing only allows data expressions.

It allows:

```text
_
booleans
numbers
percentage-suffixed numbers
negative numbers
strings
arrays
maps
```

It does not allow executable syntax such as assignments, identifiers, function calls, function literals, index expressions, binary operators or ranges.

Exponential number notation is not allowed in STULTON. Percentage-suffixed numbers are allowed when the underlying number does not use exponent notation.

### `STD["DATA"]["STULTON"]["IS_VALID"](text)`

Checks whether text is valid STULTON data.

```stult
STD.DATA.STULTON.IS_VALID("{\"active\": +}")
```

Returns a boolean.
