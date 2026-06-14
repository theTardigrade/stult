# Standard library

The Stult standard library is available through the immutable binding `STD`.

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

The standard library is map-shaped. Standard library functions are ordinary callable values stored inside maps.

Here's an example:

```stult
PRINT : STD["IO"]["PRINT"]
SIZE : STD["TYPE"]["COLLECTION"]["SIZE"]

PRINT(
	"size: "
	SIZE({"a", "b", "c"})
)
```

All standard library names are uppercase because standard library maps and functions are intended to be immutable.

Some standard-library functions accept variadic arguments. In signatures, `...name` means that zero or more remaining arguments are collected into an array under that name.

## Contents

- [STD["ASSERT"]](#stdassert)
  - [`STD["ASSERT"]["TRUE"](condition, message)`](#stdasserttruecondition-message)
  - [`STD["ASSERT"]["EQUAL"](actual, expected, message)`](#stdassertequalactual-expected-message)
- [`STD["IO"]`](#stdio)
  - [`STD["IO"]["WRITE"](...values)`](#stdiowritevalues)
  - [`STD["IO"]["WRITE_LINE"](...values)`](#stdiowrite_linevalues)
  - [`STD["IO"]["PRINT"](...values)`](#stdioprintvalues)
  - [`STD["IO"]["WRITE_ERROR"](...values)`](#stdiowrite_errorvalues)
  - [`STD["IO"]["READ_LINE"]()`](#stdioread_line)
  - [`STD["IO"]["PROMPT"](...values)`](#stdiopromptvalues)
- [`STD["SYSTEM"]`](#stdsystem)
  - [`STD["SYSTEM"]["ARGS"]`](#stdsystemargs)
  - [`STD["SYSTEM"]["CWD"]()`](#stdsystemcwd)
  - [`STD["SYSTEM"]["ENV"](name)`](#stdsystemenvname)
  - [`STD["SYSTEM"]["EXIT"](code)`](#stdsystemexitcode)
- [`STD["FILE"]`](#stdfile)
  - [`STD["FILE"]["READ"](path)`](#stdfilereadpath)
  - [`STD["FILE"]["WRITE"](path, content)`](#stdfilewritepath-content)
  - [`STD["FILE"]["APPEND"](path, content)`](#stdfileappendpath-content)
  - [`STD["FILE"]["EXISTS"](path)`](#stdfileexistspath)
  - [`STD["FILE"]["DELETE"](path)`](#stdfiledeletepath)
  - [`STD["FILE"]["RENAME"](old_path, new_path)`](#stdfilerenameold_path-new_path)
  - [`STD["FILE"]["COPY"](source_path, destination_path)`](#stdfilecopysource_path-destination_path)
  - [`STD["FILE"]["SIZE"](path)`](#stdfilesizepath)
- [`STD["PATH"]`](#stdpath)
  - [`STD["PATH"]["JOIN"](...parts)`](#stdpathjoinparts)
  - [`STD["PATH"]["ABS"](path)`](#stdpathabspath)
  - [`STD["PATH"]["BASE"](path)`](#stdpathbasepath)
  - [`STD["PATH"]["CLEAN"](path)`](#stdpathcleanpath)
  - [`STD["PATH"]["DIR"](path)`](#stdpathdirpath)
  - [`STD["PATH"]["EXT"](path)`](#stdpathextpath)
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
    - [`STD["TYPE"]["NUMBER"]["PRECISION"]`](#stdtypenumberprecision)
    - [`STD["TYPE"]["NUMBER"]["FRACTION_DIGITS"]`](#stdtypenumberfraction_digits)
    - [`STD["TYPE"]["NUMBER"]["MAX_SAFE_INTEGER"]`](#stdtypenumbermax_safe_integer)
    - [`STD["TYPE"]["NUMBER"]["MIN_SAFE_INTEGER"]`](#stdtypenumbermin_safe_integer)
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
    - [`STD["TYPE"]["COLLECTION"]["CLEAR"](collection)`](#stdtypecollectionclearcollection)
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

### `STD["ASSERT"]["TRUE"](condition, message)`

Checks that `condition` is true.

`condition` must be a boolean.

`message` must be a string.

Returns `_` if `condition` is true.

Raises a runtime error if `condition` is false.

```stult
STD["ASSERT"]["TRUE"](1 + 1 = 2, "arithmetic should work")
```

### `STD["ASSERT"]["EQUAL"](actual, expected, message)`

Checks that `actual` and `expected` are equal.

`message` must be a string.

Comparison uses normal Stult equality semantics.

Returns `_` if `actual` and `expected` are equal.

Raises a runtime error if they are not equal.

```stult
STD["ASSERT"]["EQUAL"]("Stult", "Stult", "strings should match")
STD["ASSERT"]["EQUAL"](1 + 2, 3, "numbers should match")
```

## `STD["IO"]`

Console input and output.

### `STD["IO"]["WRITE"](...values)`

Writes values to standard output without adding a newline.

```stult
STD["IO"]["WRITE"]("Hello")
STD["IO"]["WRITE"](" ")
STD["IO"]["WRITE"]("world")
```

Returns `_`.

### `STD["IO"]["WRITE_LINE"](...values)`

Writes values to standard output and then writes a newline.

```stult
STD["IO"]["WRITE_LINE"]("Hello world")
STD["IO"]["WRITE_LINE"]("Count: ", 3)
```

Returns `_`.

### `STD["IO"]["PRINT"](...values)`

Alias for `STD["IO"]["WRITE_LINE"]`.

```stult
STD["IO"]["PRINT"]("Hello")
```

Returns `_`.

### `STD["IO"]["WRITE_ERROR"](...values)`

Writes values to standard error and then writes a newline.

```stult
STD["IO"]["WRITE_ERROR"]("Something went wrong")
```

Returns `_`.

### `STD["IO"]["READ_LINE"]()`

Reads one line from standard input.

```stult
line : STD["IO"]["READ_LINE"]()
```

Returns a string without the trailing newline.

Returns `_` if input ends before any text is read.

### `STD["IO"]["PROMPT"](...values)`

Writes prompt values to standard output, then reads a line from standard input.

```stult
name : STD["IO"]["PROMPT"]("Name: ")
STD["IO"]["PRINT"]("Hello ", name)
```

Returns a string without the trailing newline.

Returns `_` if input ends before any text is read.

## `STD["SYSTEM"]`

Runtime and process-context helpers.

### `STD["SYSTEM"]["ARGS"]`

Contains the arguments passed to the Stult program.

```stult
STD["IO"]["PRINT"](STD["SYSTEM"]["ARGS"])
```

The array does not include the Stult executable path.

The array also does not include the source file, manifest file or project directory used as the program target.

For example:

```bash
stult run examples/csv_to_json_converter.stult input.csv output.json
```

makes this available to Stult code:

```stult
STD["SYSTEM"]["ARGS"] # {"input.csv", "output.json"}
```

`ARGS` is an immutable array of strings.

### `STD["SYSTEM"]["CWD"]()`

Returns the current working directory.

```stult
cwd : STD["SYSTEM"]["CWD"]()

STD["IO"]["PRINT"](cwd)
```

Returns a string.

### `STD["SYSTEM"]["ENV"](name)`

Reads an environment variable.

```stult
home : STD["SYSTEM"]["ENV"]("HOME")

STD["IO"]["PRINT"](home)
```

The argument must be a string.

Returns a string when the environment variable is set.

Returns `_` when the environment variable is not set.

### `STD["SYSTEM"]["EXIT"](code)`

Exits the current process with the given status code.

```stult
STD["SYSTEM"]["EXIT"](0)
```

The exit code must be an integer number from `0` to `255`.

This function does not return, because it terminates the process.

## `STD["FILE"]`

File-system helpers.

### `STD["FILE"]["READ"](path)`

Reads a file as text.

```stult
text : STD["FILE"]["READ"]("notes.txt")
```

Returns a string.

### `STD["FILE"]["WRITE"](path, content)`

Writes content to a file, replacing existing contents.

```stult
STD["FILE"]["WRITE"]("notes.txt", "Hello")
```

The path must be a string.

The content may be a string or another Stult value. Non-string values are converted with their printed representation.

Returns `_`.

### `STD["FILE"]["APPEND"](path, content)`

Appends content to a file.

```stult
STD["FILE"]["APPEND"]("notes.txt", "\nAnother line")
```

Creates the file if it does not exist.

Returns `_`.

### `STD["FILE"]["EXISTS"](path)`

Checks whether a file-system path exists.

```stult
(STD["FILE"]["EXISTS"]("notes.txt")) {
	STD["IO"]["PRINT"]("notes.txt exists")
}
```

Returns a boolean.

### `STD["FILE"]["DELETE"](path)`

Deletes a file-system path.

```stult
STD["FILE"]["DELETE"]("notes.txt")
```

Returns `_`.

### `STD["FILE"]["RENAME"](old_path, new_path)`

Renames or moves a file-system path.

```stult
STD["FILE"]["RENAME"]("old.txt", "new.txt")
```

Returns `_`.

### `STD["FILE"]["COPY"](source_path, destination_path)`

Copies a file.

```stult
STD["FILE"]["COPY"]("source.txt", "copy.txt")
```

The destination is created or replaced.

Returns `_`.

### `STD["FILE"]["SIZE"](path)`

Returns the size of a file-system path in bytes.

```stult
size : STD["FILE"]["SIZE"]("notes.txt")
```

Returns a number.

## `STD["PATH"]`

Path manipulation helpers.

### `STD["PATH"]["JOIN"](...parts)`

Joins path parts into one path.

```stult
path : STD["PATH"]["JOIN"]("data", "input.csv")

STD["IO"]["PRINT"](path)
```

Requires at least one argument.

Each argument must be a string.

Returns a string.

### `STD["PATH"]["ABS"](path)`

Returns an absolute version of a path.

```stult
absolute_path : STD["PATH"]["ABS"]("input.csv")

STD["IO"]["PRINT"](absolute_path)
```

The argument must be a string.

Returns a string.

### `STD["PATH"]["BASE"](path)`

Returns the final element of a path.

```stult
name : STD["PATH"]["BASE"]("data/input.csv")

STD["IO"]["PRINT"](name)
```

The argument must be a string.

Returns a string.

### `STD["PATH"]["CLEAN"](path)`

Cleans a path by removing redundant path elements.

```stult
clean_path : STD["PATH"]["CLEAN"]("./data/../data/input.csv")

STD["IO"]["PRINT"](clean_path)
```

The argument must be a string.

Returns a string.

### `STD["PATH"]["DIR"](path)`

Returns the directory part of a path.

```stult
dir : STD["PATH"]["DIR"]("data/input.csv")

STD["IO"]["PRINT"](dir)
```

The argument must be a string.

Returns a string.

### `STD["PATH"]["EXT"](path)`

Returns the file extension of a path.

```stult
extension : STD["PATH"]["EXT"]("data/input.csv")

STD["IO"]["PRINT"](extension)
```

The argument must be a string.

Returns a string.

## `STD["TIME"]`

Timestamps, sleep and calendar snapshots.

### `STD["TIME"]["MILLI_TIMESTAMP"]()`

Returns the current Unix timestamp in milliseconds.

```stult
start : STD["TIME"]["MILLI_TIMESTAMP"]()
```

Returns a number.

### `STD["TIME"]["NANO_TIMESTAMP"]()`

Returns the current Unix timestamp in nanoseconds.

```stult
start : STD["TIME"]["NANO_TIMESTAMP"]()
```

Returns a number.

### `STD["TIME"]["MILLI_SLEEP"](milliseconds)`

Sleeps for the given number of milliseconds.

```stult
STD["TIME"]["MILLI_SLEEP"](500)
```

The argument must be a non-negative integer number.

Returns `_`.

### `STD["TIME"]["LOCAL_CALENDAR"]()`

Returns a map describing the current local time.

```stult
now : STD["TIME"]["LOCAL_CALENDAR"]()

STD["IO"]["PRINT"](now["YEAR"], "-", now["MONTH"], "-", now["DAY"])
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
utc : STD["TIME"]["UTC_CALENDAR"]()
```

The returned map has the same keys as `LOCAL_CALENDAR`.

## `STD["MATH"]`

Numeric helpers, constants and trigonometry.

### `STD["MATH"]["PI"]`

The mathematical constant pi.

```stult
STD["MATH"]["PI"]
```

### `STD["MATH"]["TAU"]`

The mathematical constant tau.

```stult
STD["MATH"]["TAU"]
```

`TAU` is `PI * 2`.

### `STD["MATH"]["E"]`

The mathematical constant e.

```stult
STD["MATH"]["E"]
```

### `STD["MATH"]["SQUARE"](number)`

Returns `number * number`.

```stult
STD["MATH"]["SQUARE"](9)
```

### `STD["MATH"]["CUBE"](number)`

Returns `number * number * number`.

```stult
STD["MATH"]["CUBE"](3)
```

### `STD["MATH"]["ABS"](number)`

Returns the absolute value.

```stult
STD["MATH"]["ABS"](-12.5)
```

### `STD["MATH"]["SIGN"](number)`

Returns:

```text
-1 for negative numbers
0 for zero
1 for positive numbers
```

```stult
STD["MATH"]["SIGN"](-12.5)
```

### `STD["MATH"]["MIN"](...numbers)`

Returns the smallest number.

```stult
STD["MATH"]["MIN"](8, -3, 10, 2)
```

Requires at least one argument.

### `STD["MATH"]["MAX"](...numbers)`

Returns the largest number.

```stult
STD["MATH"]["MAX"](8, -3, 10, 2)
```

Requires at least one argument.

### `STD["MATH"]["CLAMP"](value, minimum, maximum)`

Restricts a number to a range.

```stult
STD["MATH"]["CLAMP"](15, 0, 10)
```

Returns `minimum` if `value` is below `minimum`.

Returns `maximum` if `value` is above `maximum`.

Returns `value` otherwise.

### `STD["MATH"]["LERP"](start, end, amount)`

Linearly interpolates between `start` and `end`.

```stult
STD["MATH"]["LERP"](0, 10, 0.5)
```

### `STD["MATH"]["FLOOR"](number)`

Rounds down to an integer.

```stult
STD["MATH"]["FLOOR"](3.9)
```

### `STD["MATH"]["CEIL"](number)`

Rounds up to an integer.

```stult
STD["MATH"]["CEIL"](3.1)
```

### `STD["MATH"]["ROUND"](number)`

Rounds to the nearest integer.

```stult
STD["MATH"]["ROUND"](3.5)
```

Positive halves round upward.

Negative halves round downward.

### `STD["MATH"]["TRUNC"](number)`

Removes the fractional part.

```stult
STD["MATH"]["TRUNC"](-3.9)
```

### `STD["MATH"]["SQRT"](number)`

Returns the square root.

```stult
STD["MATH"]["SQRT"](2)
```

The argument must be non-negative.

### `STD["MATH"]["POWER"](base, exponent)`

Raises `base` to `exponent`.

```stult
STD["MATH"]["POWER"](2, 10)
STD["MATH"]["POWER"](2, 0.5)
```

Zero cannot be raised to a negative exponent.

Negative bases with non-integer exponents are not allowed.

### `STD["MATH"]["MOD"](left, right)`

Returns the modulo remainder.

```stult
STD["MATH"]["MOD"](10, 3)
```

The divisor cannot be zero.

## `STD["MATH"]["TRIG"]`

Trigonometric functions use radians.

### `STD["MATH"]["TRIG"]["SIN"](radians)`

Returns the sine of an angle.

```stult
STD["MATH"]["TRIG"]["SIN"](STD["MATH"]["PI"] / 2)
```

### `STD["MATH"]["TRIG"]["COS"](radians)`

Returns the cosine of an angle.

```stult
STD["MATH"]["TRIG"]["COS"](0)
```

### `STD["MATH"]["TRIG"]["TAN"](radians)`

Returns the tangent of an angle.

```stult
STD["MATH"]["TRIG"]["TAN"](STD["MATH"]["PI"] / 4)
```

Tangent is not defined where cosine is zero.

### `STD["MATH"]["TRIG"]["RADIANS"](degrees)`

Converts degrees to radians.

```stult
STD["MATH"]["TRIG"]["RADIANS"](180)
```

### `STD["MATH"]["TRIG"]["DEGREES"](radians)`

Converts radians to degrees.

```stult
STD["MATH"]["TRIG"]["DEGREES"](STD["MATH"]["PI"])
```

## `STD["TYPE"]`

Type predicates, conversions and collection helpers.

## `STD["TYPE"]` predicates

Each predicate accepts one value and returns a boolean.

```stult
STD["TYPE"]["IS_VOID"](_)
STD["TYPE"]["IS_BOOL"](\/)
STD["TYPE"]["IS_NUMBER"](123)
STD["TYPE"]["IS_STRING"]("hello")
STD["TYPE"]["IS_ARRAY"]({})
STD["TYPE"]["IS_MAP"]({:})
STD["TYPE"]["IS_FUNCTION"]({ () (_) })
STD["TYPE"]["IS_BUILTIN_FUNCTION"](STD["IO"]["PRINT"])
STD["TYPE"]["IS_COLLECTION"]({"a", "b"})
```

`STD["TYPE"]["IS_COLLECTION"]` returns true for arrays, maps and strings.

## `STD["TYPE"]["BOOL"]`

Boolean constants and conversion helpers.

### `STD["TYPE"]["BOOL"]["TRUE"]`

The standard-library boolean true constant.

```stult
STD["TYPE"]["BOOL"]["TRUE"]
```

This value is equivalent to the boolean literal `\/`.

### `STD["TYPE"]["BOOL"]["FALSE"]`

The standard-library boolean false constant.

```stult
STD["TYPE"]["BOOL"]["FALSE"]
```

This value is equivalent to the boolean literal `/\`.

### `STD["TYPE"]["BOOL"]["NEW"](value)`

Converts a value to a boolean when possible.

```stult
STD["TYPE"]["BOOL"]["NEW"]("true")
STD["TYPE"]["BOOL"]["NEW"]("false")
STD["TYPE"]["BOOL"]["NEW"](1)
STD["TYPE"]["BOOL"]["NEW"](0)
```

Conversion rules:

```text
bool      returns itself
number    false if zero, true otherwise
string    "T", "TRUE" and "1" become true
string    "F", "FALSE" and "0" become false
other     returns _
```

String conversion ignores surrounding whitespace and case.

## `STD["TYPE"]["NUMBER"]`

Number constants and conversion helpers.

### `STD["TYPE"]["NUMBER"]["PRECISION"]`

The default high-precision number precision, measured in bits.

```stult
STD["TYPE"]["NUMBER"]["PRECISION"]
```

### `STD["TYPE"]["NUMBER"]["FRACTION_DIGITS"]`

The default number of fractional digits used by ordinary number formatting.

```stult
STD["TYPE"]["NUMBER"]["FRACTION_DIGITS"]
```

### `STD["TYPE"]["NUMBER"]["MAX_SAFE_INTEGER"]`

The largest positive integer for which all integer values from `0` through that value are exactly representable at Stult's default high-precision number precision.

```stult
STD["TYPE"]["NUMBER"]["MAX_SAFE_INTEGER"]
```

This is a safe-integer bound, not the maximum numeric magnitude Stult can represent.

### `STD["TYPE"]["NUMBER"]["MIN_SAFE_INTEGER"]`

The negative counterpart of `STD["TYPE"]["NUMBER"]["MAX_SAFE_INTEGER"]`.

```stult
STD["TYPE"]["NUMBER"]["MIN_SAFE_INTEGER"]
```

This is a safe-integer bound, not the minimum numeric magnitude Stult can represent.

### `STD["TYPE"]["NUMBER"]["NEW"](value)`

Converts a value to a number when possible.

```stult
STD["TYPE"]["NUMBER"]["NEW"]("123.45")
STD["TYPE"]["NUMBER"]["NEW"]("\/")  # not useful; booleans are values, so use \/ directly
STD["TYPE"]["NUMBER"]["NEW"](\/)
```

Conversion rules:

```text
number    returns a cloned number
bool      true becomes 1, false becomes 0
string    parsed as a number after trimming whitespace
other     returns _
```

Invalid number strings return `_`.

## `STD["TYPE"]["STRING"]`

String conversion and string utility functions.

### `STD["TYPE"]["STRING"]["NEW"](value)`

Converts a value to a string.

```stult
STD["TYPE"]["STRING"]["NEW"](123)
STD["TYPE"]["STRING"]["NEW"](\/)
STD["TYPE"]["STRING"]["NEW"]({"a", "b"})
```

### `STD["TYPE"]["STRING"]["CHARACTERS"](text)`

Returns an array of one-character strings.

```stult
STD["TYPE"]["STRING"]["CHARACTERS"]("cat")
```

Result:

```stult
{"c", "a", "t"}
```

### `STD["TYPE"]["STRING"]["TRIM"](text)`

Removes leading and trailing whitespace.

```stult
STD["TYPE"]["STRING"]["TRIM"]("  hello  ")
```

### `STD["TYPE"]["STRING"]["TRIM_START"](text)`

Removes leading whitespace.

```stult
STD["TYPE"]["STRING"]["TRIM_START"]("  hello")
```

### `STD["TYPE"]["STRING"]["TRIM_END"](text)`

Removes trailing whitespace.

```stult
STD["TYPE"]["STRING"]["TRIM_END"]("hello  ")
```

### `STD["TYPE"]["STRING"]["TO_LOWER"](text)`

Converts text to lowercase.

```stult
STD["TYPE"]["STRING"]["TO_LOWER"]("Hello")
```

### `STD["TYPE"]["STRING"]["TO_UPPER"](text)`

Converts text to uppercase.

```stult
STD["TYPE"]["STRING"]["TO_UPPER"]("Hello")
```

### `STD["TYPE"]["STRING"]["IS_FOUND_IN"](search, text)`

Checks whether `search` appears inside `text`.

```stult
STD["TYPE"]["STRING"]["IS_FOUND_IN"]("ell", "hello")
```

### `STD["TYPE"]["STRING"]["IS_FOUND_AT_START"](search, text)`

Checks whether `text` starts with `search`.

```stult
STD["TYPE"]["STRING"]["IS_FOUND_AT_START"]("he", "hello")
```

### `STD["TYPE"]["STRING"]["IS_FOUND_AT_END"](search, text)`

Checks whether `text` ends with `search`.

```stult
STD["TYPE"]["STRING"]["IS_FOUND_AT_END"]("lo", "hello")
```

### `STD["TYPE"]["STRING"]["REPLACE"](text, old, new)`

Replaces every occurrence of `old` with `new`.

```stult
STD["TYPE"]["STRING"]["REPLACE"]("one two one", "one", "three")
```

### `STD["TYPE"]["STRING"]["SPLIT"](text, separator)`

Splits text into an array of strings.

```stult
STD["TYPE"]["STRING"]["SPLIT"]("a,b,c", ",")
```

### `STD["TYPE"]["STRING"]["JOIN"](array, separator)`

Joins array elements into a string.

```stult
STD["TYPE"]["STRING"]["JOIN"]({"a", "b", "c"}, ",")
```

String elements are used directly.

Non-string elements are converted with their printed representation.

## `STD["TYPE"]["ARRAY"]`

Array-specific helpers.

### `STD["TYPE"]["ARRAY"]["APPEND"](array, ...values)`

Appends one or more values to an array.

```stult
items : {}

STD["TYPE"]["ARRAY"]["APPEND"](items, "a", "b", "c")
```

Returns `_`.

The first argument must be an array.

Raises a runtime error if the array is frozen.

## `STD["TYPE"]["MAP"]`

Map-specific helpers.

### `STD["TYPE"]["MAP"]["KEYS"](map)`

Returns an array of map keys sorted lexicographically.

```stult
keys : STD["TYPE"]["MAP"]["KEYS"]({"b": 2, "a": 1})
```

Returns `_` when the value is not a map.

### `STD["TYPE"]["MAP"]["VALUES"](map)`

Returns an array of map values sorted by key order.

```stult
values : STD["TYPE"]["MAP"]["VALUES"]({"b": 2, "a": 1})
```

Returns `_` when the value is not a map.

## `STD["TYPE"]["COLLECTION"]`

Helpers shared by arrays, maps and strings.

### `STD["TYPE"]["COLLECTION"]["SIZE"](collection)`

Returns the size of a collection.

```stult
STD["TYPE"]["COLLECTION"]["SIZE"]({"a", "b", "c"})
STD["TYPE"]["COLLECTION"]["SIZE"]({"name": "example"})
STD["TYPE"]["COLLECTION"]["SIZE"]("hello")
```

For arrays, size is the number of elements.

For maps, size is the number of entries.

For strings, size is the number of runes.

Returns `_` for non-collections.

### `STD["TYPE"]["COLLECTION"]["IS_EMPTY"](collection)`

Checks whether a collection is empty.

```stult
STD["TYPE"]["COLLECTION"]["IS_EMPTY"]({})
STD["TYPE"]["COLLECTION"]["IS_EMPTY"]("")
```

Returns `_` for non-collections.

### `STD["TYPE"]["COLLECTION"]["HAS"](collection, key)`

Checks whether a collection contains a key or index.

```stult
STD["TYPE"]["COLLECTION"]["HAS"]({"name": "example"}, "name")
STD["TYPE"]["COLLECTION"]["HAS"]({"a", "b"}, 1)
STD["TYPE"]["COLLECTION"]["HAS"]("cat", 0)
```

For maps, the key must be a string.

For arrays and strings, the key must be a valid numeric index.

Returns `_` for non-collections.

### `STD["TYPE"]["COLLECTION"]["CLEAR"](collection)`

Removes all contents from a non-frozen collection.

```stult
items : {"a", "b"}
STD["TYPE"]["COLLECTION"]["CLEAR"](items)
```

For arrays, this removes all elements.

For maps, this removes all entries.

For strings, this removes all characters.

Returns `_`.

Raises a runtime error if the collection is frozen.

### `STD["TYPE"]["COLLECTION"]["FREEZE"](collection)`

Deeply freezes a collection.

```stult
CONFIG : STD["TYPE"]["COLLECTION"]["FREEZE"]({
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
STD["TYPE"]["COLLECTION"]["FREEZE"](items)

other_items : STD["TYPE"]["COLLECTION"]["FREEZE"]({4, 5, 6})
```

Returns `_` for non-collections.

### `STD["TYPE"]["COLLECTION"]["IS_FROZEN"](value)`

Checks whether a value is a frozen collection.

```stult
items : STD["TYPE"]["COLLECTION"]["FREEZE"]({1, 2, 3})

STD["TYPE"]["COLLECTION"]["IS_FROZEN"](items)
STD["TYPE"]["COLLECTION"]["IS_FROZEN"](items[0])
```

Returns `\/` for frozen arrays, maps and strings.

Returns `/\` for non-frozen arrays, maps and strings.

Returns `/\` for non-collections.

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

text : STD["DATA"]["CSV"]["NEW"](rows)
```

Fields are converted to strings when possible.

Returns a string.

### `STD["DATA"]["CSV"]["PARSE"](text)`

Parses CSV text into an array of row arrays.

```stult
rows : STD["DATA"]["CSV"]["PARSE"]("name,score\na,10\nb,20\n")
```

Returns an array of arrays of strings.

### `STD["DATA"]["CSV"]["IS_VALID"](text)`

Checks whether text is valid CSV.

```stult
STD["DATA"]["CSV"]["IS_VALID"]("name,score\na,10\n")
```

Returns a boolean.

Rows may have different field counts.

## `STD["DATA"]["JSON"]`

JSON encoding, parsing and validation helpers.

### `STD["DATA"]["JSON"]["NEW"](value)`

Encodes a Stult value as JSON text.

```stult
text : STD["DATA"]["JSON"]["NEW"]({
	"name": "example"
	"active": \/
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
value : STD["DATA"]["JSON"]["PARSE"]("{\"name\":\"example\",\"active\":true}")
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
STD["DATA"]["JSON"]["IS_VALID"]("{\"ok\":true}")
```

Returns a boolean.

## `STD["DATA"]["STULTON"]`

STULTON encoding, parsing and validation helpers.

### `STD["DATA"]["STULTON"]["NEW"](value)`

Encodes a Stult value as STULTON text.

```stult
text : STD["DATA"]["STULTON"]["NEW"]({
	"NAME": "example"
	"active": \/
	"items": {
		"one"
		"two"
	}
})
```

Conversion rules:

```text
_         _
bool      \/ or /\
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
value : STD["DATA"]["STULTON"]["PARSE"]("{\"active\": \\/}")
```

STULTON parsing only allows data expressions.

It allows:

```text
_
booleans
numbers
negative numbers
strings
arrays
maps
```

It does not allow executable syntax such as assignments, identifiers, function calls, function literals, index expressions, binary operators or ranges.

Exponential number notation is not allowed in STULTON.

### `STD["DATA"]["STULTON"]["IS_VALID"](text)`

Checks whether text is valid STULTON data.

```stult
STD["DATA"]["STULTON"]["IS_VALID"]("{\"active\": \\/}")
```

Returns a boolean.