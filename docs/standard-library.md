# Standard library

The Stult standard library is available through the immutable binding `STD`.

```stult
STD["IO"]
STD["FILE"]
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

PRINT("size: ", SIZE({"a", "b", "c"}))
```

All standard library names are uppercase because standard library maps and functions are intended to be immutable.

## Contents

- [Top-level maps](#top-level-maps)
- [IO](#io)
  - [`IO["WRITE"](...values)`](#iowritevalues)
  - [`IO["WRITE_LINE"](...values)`](#iowrite_linevalues)
  - [`IO["PRINT"](...values)`](#ioprintvalues)
  - [`IO["WRITE_ERROR"](...values)`](#iowrite_errorvalues)
  - [`IO["READ_LINE"]()`](#ioread_line)
  - [`IO["PROMPT"](...values)`](#iopromptvalues)
- [FILE](#file)
  - [`FILE["READ"](path)`](#filereadpath)
  - [`FILE["WRITE"](path, content)`](#filewritepath-content)
  - [`FILE["APPEND"](path, content)`](#fileappendpath-content)
  - [`FILE["EXISTS"](path)`](#fileexistspath)
  - [`FILE["DELETE"](path)`](#filedeletepath)
  - [`FILE["RENAME"](old_path, new_path)`](#filerenameold_path-new_path)
  - [`FILE["COPY"](source_path, destination_path)`](#filecopysource_path-destination_path)
  - [`FILE["SIZE"](path)`](#filesizepath)
- [TIME](#time)
  - [`TIME["MILLI_TIMESTAMP"]()`](#timemilli_timestamp)
  - [`TIME["NANO_TIMESTAMP"]()`](#timenano_timestamp)
  - [`TIME["MILLI_SLEEP"](milliseconds)`](#timemilli_sleepmilliseconds)
  - [`TIME["LOCAL_CALENDAR"]()`](#timelocal_calendar)
  - [`TIME["UTC_CALENDAR"]()`](#timeutc_calendar)
- [MATH](#math)
  - [Constants](#constants)
  - [`MATH["SQUARE"](number)`](#mathsquarenumber)
  - [`MATH["CUBE"](number)`](#mathcubenumber)
  - [`MATH["ABS"](number)`](#mathabsnumber)
  - [`MATH["SIGN"](number)`](#mathsignnumber)
  - [`MATH["MIN"](...numbers)`](#mathminnumbers)
  - [`MATH["MAX"](...numbers)`](#mathmaxnumbers)
  - [`MATH["CLAMP"](value, minimum, maximum)`](#mathclampvalue-minimum-maximum)
  - [`MATH["LERP"](start, end, amount)`](#mathlerpstart-end-amount)
  - [`MATH["FLOOR"](number)`](#mathfloornumber)
  - [`MATH["CEIL"](number)`](#mathceilnumber)
  - [`MATH["ROUND"](number)`](#mathroundnumber)
  - [`MATH["TRUNC"](number)`](#mathtruncnumber)
  - [`MATH["SQRT"](number)`](#mathsqrtnumber)
  - [`MATH["POWER"](base, exponent)`](#mathpowerbase-exponent)
  - [`MATH["MOD"](left, right)`](#mathmodleft-right)
  - [TRIG](#trig)
    - [`TRIG["SIN"](radians)`](#trigsinradians)
    - [`TRIG["COS"](radians)`](#trigcosradians)
    - [`TRIG["TAN"](radians)`](#trigtanradians)
    - [`TRIG["RADIANS"](degrees)`](#trigradiansdegrees)
    - [`TRIG["DEGREES"](radians)`](#trigdegreesradians)
- [TYPE](#type)
  - [TYPE predicates](#type-predicates)
  - [BOOL](#bool)
    - [Constants](#constants-1)
    - [`BOOL["NEW"](value)`](#boolnewvalue)
  - [NUMBER](#number)
    - [Constants](#constants-2)
    - [`NUMBER["NEW"](value)`](#numbernewvalue)
  - [STRING](#string)
    - [`STRING["NEW"](value)`](#stringnewvalue)
    - [`STRING["CHARACTERS"](text)`](#stringcharacterstext)
    - [`STRING["TRIM"](text)`](#stringtrimtext)
    - [`STRING["TRIM_START"](text)`](#stringtrim_starttext)
    - [`STRING["TRIM_END"](text)`](#stringtrim_endtext)
    - [`STRING["TO_LOWER"](text)`](#stringto_lowertext)
    - [`STRING["TO_UPPER"](text)`](#stringto_uppertext)
    - [`STRING["IS_FOUND_IN"](search, text)`](#stringis_found_insearch-text)
    - [`STRING["IS_FOUND_AT_START"](search, text)`](#stringis_found_at_startsearch-text)
    - [`STRING["IS_FOUND_AT_END"](search, text)`](#stringis_found_at_endsearch-text)
    - [`STRING["REPLACE"](text, old, new)`](#stringreplacetext-old-new)
    - [`STRING["SPLIT"](text, separator)`](#stringsplittext-separator)
    - [`STRING["JOIN"](array, separator)`](#stringjoinarray-separator)
  - [ARRAY](#array)
    - [`ARRAY["APPEND"](array, ...values)`](#arrayappendarray-values)
  - [MAP](#map)
    - [`MAP["KEYS"](map)`](#mapkeysmap)
    - [`MAP["VALUES"](map)`](#mapvaluesmap)
  - [COLLECTION](#collection)
    - [`COLLECTION["SIZE"](collection)`](#collectionsizecollection)
    - [`COLLECTION["IS_EMPTY"](collection)`](#collectionis_emptycollection)
    - [`COLLECTION["HAS"](collection, key)`](#collectionhascollection-key)
    - [`COLLECTION["CLEAR"](collection)`](#collectionclearcollection)
- [DATA](#data)
  - [CSV](#csv)
    - [`CSV["NEW"](rows)`](#csvnewrows)
    - [`CSV["PARSE"](text)`](#csvparsetext)
    - [`CSV["IS_VALID"](text)`](#csvis_validtext)
  - [JSON](#json)
    - [`JSON["NEW"](value)`](#jsonnewvalue)
    - [`JSON["PARSE"](text)`](#jsonparsetext)
    - [`JSON["IS_VALID"](text)`](#jsonis_validtext)
  - [STULTON](#stulton)
    - [`STULTON["NEW"](value)`](#stultonnewvalue)
    - [`STULTON["PARSE"](text)`](#stultonparsetext)
    - [`STULTON["IS_VALID"](text)`](#stultonis_validtext)

## Top-level maps

### `STD["IO"]`

Console input and output.

### `STD["FILE"]`

File-system helpers.

### `STD["TIME"]`

Timestamps, sleep and calendar snapshots.

### `STD["MATH"]`

Numeric helpers, constants and trigonometry.

### `STD["TYPE"]`

Type predicates, conversions and collection helpers.

### `STD["DATA"]`

Data encoding and decoding helpers.

## IO

```stult
IO : STD["IO"]
```

### `IO["WRITE"](...values)`

Writes values to standard output without adding a newline.

```stult
IO["WRITE"]("Hello")
IO["WRITE"](" ")
IO["WRITE"]("world")
```

Returns `_`.

### `IO["WRITE_LINE"](...values)`

Writes values to standard output and then writes a newline.

```stult
IO["WRITE_LINE"]("Hello world")
IO["WRITE_LINE"]("Count: ", 3)
```

Returns `_`.

### `IO["PRINT"](...values)`

Alias for `IO["WRITE_LINE"]`.

```stult
PRINT : STD["IO"]["PRINT"]

PRINT("Hello")
```

Returns `_`.

### `IO["WRITE_ERROR"](...values)`

Writes values to standard error and then writes a newline.

```stult
STD["IO"]["WRITE_ERROR"]("Something went wrong")
```

Returns `_`.

### `IO["READ_LINE"]()`

Reads one line from standard input.

```stult
line : STD["IO"]["READ_LINE"]()
```

Returns a string without the trailing newline.

Returns `_` if input ends before any text is read.

### `IO["PROMPT"](...values)`

Writes prompt values to standard output, then reads a line from standard input.

```stult
name : STD["IO"]["PROMPT"]("Name: ")
STD["IO"]["PRINT"]("Hello ", name)
```

Returns a string without the trailing newline.

Returns `_` if input ends before any text is read.

## FILE

```stult
FILE : STD["FILE"]
```

File paths must be strings.

### `FILE["READ"](path)`

Reads a file as text.

```stult
text : FILE["READ"]("notes.txt")
```

Returns a string.

### `FILE["WRITE"](path, content)`

Writes content to a file, replacing existing contents.

```stult
FILE["WRITE"]("notes.txt", "Hello")
```

The path must be a string.

The content may be a string or another Stult value. Non-string values are converted with their printed representation.

Returns `_`.

### `FILE["APPEND"](path, content)`

Appends content to a file.

```stult
FILE["APPEND"]("notes.txt", "\nAnother line")
```

Creates the file if it does not exist.

Returns `_`.

### `FILE["EXISTS"](path)`

Checks whether a file-system path exists.

```stult
(FILE["EXISTS"]("notes.txt")) {
	STD["IO"]["PRINT"]("notes.txt exists")
}
```

Returns a boolean.

### `FILE["DELETE"](path)`

Deletes a file-system path.

```stult
FILE["DELETE"]("notes.txt")
```

Returns `_`.

### `FILE["RENAME"](old_path, new_path)`

Renames or moves a file-system path.

```stult
FILE["RENAME"]("old.txt", "new.txt")
```

Returns `_`.

### `FILE["COPY"](source_path, destination_path)`

Copies a file.

```stult
FILE["COPY"]("source.txt", "copy.txt")
```

The destination is created or replaced.

Returns `_`.

### `FILE["SIZE"](path)`

Returns the size of a file-system path in bytes.

```stult
size : FILE["SIZE"]("notes.txt")
```

Returns a number.

## TIME

```stult
TIME : STD["TIME"]
```

### `TIME["MILLI_TIMESTAMP"]()`

Returns the current Unix timestamp in milliseconds.

```stult
start : TIME["MILLI_TIMESTAMP"]()
```

Returns a number.

### `TIME["NANO_TIMESTAMP"]()`

Returns the current Unix timestamp in nanoseconds.

```stult
start : TIME["NANO_TIMESTAMP"]()
```

Returns a number.

### `TIME["MILLI_SLEEP"](milliseconds)`

Sleeps for the given number of milliseconds.

```stult
TIME["MILLI_SLEEP"](500)
```

The argument must be a non-negative integer number.

Returns `_`.

### `TIME["LOCAL_CALENDAR"]()`

Returns a map describing the current local time.

```stult
now : TIME["LOCAL_CALENDAR"]()

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

### `TIME["UTC_CALENDAR"]()`

Returns a map describing the current UTC time and date.

```stult
utc : TIME["UTC_CALENDAR"]()
```

The returned map has the same keys as `LOCAL_CALENDAR`.

## MATH

```stult
MATH : STD["MATH"]
```

### Constants

```stult
MATH["PI"]
MATH["TAU"]
MATH["E"]
```

`TAU` is `PI * 2`.

### `MATH["SQUARE"](number)`

Returns `number * number`.

```stult
MATH["SQUARE"](9)
```

### `MATH["CUBE"](number)`

Returns `number * number * number`.

```stult
MATH["CUBE"](3)
```

### `MATH["ABS"](number)`

Returns the absolute value.

```stult
MATH["ABS"](-12.5)
```

### `MATH["SIGN"](number)`

Returns:

```text
-1 for negative numbers
0 for zero
1 for positive numbers
```

```stult
MATH["SIGN"](-12.5)
```

### `MATH["MIN"](...numbers)`

Returns the smallest number.

```stult
MATH["MIN"](8, -3, 10, 2)
```

Requires at least one argument.

### `MATH["MAX"](...numbers)`

Returns the largest number.

```stult
MATH["MAX"](8, -3, 10, 2)
```

Requires at least one argument.

### `MATH["CLAMP"](value, minimum, maximum)`

Restricts a number to a range.

```stult
MATH["CLAMP"](15, 0, 10)
```

Returns `minimum` if `value` is below `minimum`.

Returns `maximum` if `value` is above `maximum`.

Returns `value` otherwise.

### `MATH["LERP"](start, end, amount)`

Linearly interpolates between `start` and `end`.

```stult
MATH["LERP"](0, 10, 0.5)
```

### `MATH["FLOOR"](number)`

Rounds down to an integer.

```stult
MATH["FLOOR"](3.9)
```

### `MATH["CEIL"](number)`

Rounds up to an integer.

```stult
MATH["CEIL"](3.1)
```

### `MATH["ROUND"](number)`

Rounds to the nearest integer.

```stult
MATH["ROUND"](3.5)
```

Positive halves round upward.

Negative halves round downward.

### `MATH["TRUNC"](number)`

Removes the fractional part.

```stult
MATH["TRUNC"](-3.9)
```

### `MATH["SQRT"](number)`

Returns the square root.

```stult
MATH["SQRT"](2)
```

The argument must be non-negative.

### `MATH["POWER"](base, exponent)`

Raises `base` to `exponent`.

```stult
MATH["POWER"](2, 10)
MATH["POWER"](2, 0.5)
```

Zero cannot be raised to a negative exponent.

Negative bases with non-integer exponents are not allowed.

### `MATH["MOD"](left, right)`

Returns the modulo remainder.

```stult
MATH["MOD"](10, 3)
```

The divisor cannot be zero.

## TRIG

```stult
TRIG : STD["MATH"]["TRIG"]
```

Trigonometric functions use radians.

### `TRIG["SIN"](radians)`

Returns the sine of an angle.

```stult
TRIG["SIN"](STD["MATH"]["PI"] / 2)
```

### `TRIG["COS"](radians)`

Returns the cosine of an angle.

```stult
TRIG["COS"](0)
```

### `TRIG["TAN"](radians)`

Returns the tangent of an angle.

```stult
TRIG["TAN"](STD["MATH"]["PI"] / 4)
```

Tangent is not defined where cosine is zero.

### `TRIG["RADIANS"](degrees)`

Converts degrees to radians.

```stult
TRIG["RADIANS"](180)
```

### `TRIG["DEGREES"](radians)`

Converts radians to degrees.

```stult
TRIG["DEGREES"](STD["MATH"]["PI"])
```

## TYPE

```stult
TYPE : STD["TYPE"]
```

`TYPE` contains type predicates, conversion helpers and collection helpers.

## TYPE predicates

Each predicate accepts one value and returns a boolean.

```stult
TYPE["IS_VOID"](_)
TYPE["IS_BOOL"](\/)
TYPE["IS_NUMBER"](123)
TYPE["IS_STRING"]("hello")
TYPE["IS_ARRAY"]({})
TYPE["IS_MAP"]({:})
TYPE["IS_FUNCTION"]({ () (_) })
TYPE["IS_BUILTIN_FUNCTION"](STD["IO"]["PRINT"])
TYPE["IS_COLLECTION"]({"a", "b"})
```

`TYPE["IS_COLLECTION"]` returns true for arrays, maps and strings.

## BOOL

```stult
BOOL : STD["TYPE"]["BOOL"]
```

### Constants

```stult
BOOL["TRUE"]
BOOL["FALSE"]
```

### `BOOL["NEW"](value)`

Converts a value to a boolean when possible.

```stult
BOOL["NEW"]("true")
BOOL["NEW"]("false")
BOOL["NEW"](1)
BOOL["NEW"](0)
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

## NUMBER

```stult
NUMBER : STD["TYPE"]["NUMBER"]
```

### Constants

```stult
NUMBER["PRECISION"]
NUMBER["FRACTION_DIGITS"]
NUMBER["MAX_SAFE_INTEGER"]
NUMBER["MIN_SAFE_INTEGER"]
```

### `NUMBER["NEW"](value)`

Converts a value to a number when possible.

```stult
NUMBER["NEW"]("123.45")
NUMBER["NEW"]("\/")  # not useful; booleans are values, so use \/ directly
NUMBER["NEW"](\/)
```

Conversion rules:

```text
number    returns a cloned number
bool      true becomes 1, false becomes 0
string    parsed as a number after trimming whitespace
other     returns _
```

Invalid number strings return `_`.

## STRING

```stult
STRING : STD["TYPE"]["STRING"]
```

### `STRING["NEW"](value)`

Converts a value to a string.

```stult
STRING["NEW"](123)
STRING["NEW"](\/)
STRING["NEW"]({"a", "b"})
```

### `STRING["CHARACTERS"](text)`

Returns an array of one-character strings.

```stult
STRING["CHARACTERS"]("cat")
```

Result:

```stult
{"c", "a", "t"}
```

### `STRING["TRIM"](text)`

Removes leading and trailing whitespace.

```stult
STRING["TRIM"]("  hello  ")
```

### `STRING["TRIM_START"](text)`

Removes leading whitespace.

```stult
STRING["TRIM_START"]("  hello")
```

### `STRING["TRIM_END"](text)`

Removes trailing whitespace.

```stult
STRING["TRIM_END"]("hello  ")
```

### `STRING["TO_LOWER"](text)`

Converts text to lowercase.

```stult
STRING["TO_LOWER"]("Hello")
```

### `STRING["TO_UPPER"](text)`

Converts text to uppercase.

```stult
STRING["TO_UPPER"]("Hello")
```

### `STRING["IS_FOUND_IN"](search, text)`

Checks whether `search` appears inside `text`.

```stult
STRING["IS_FOUND_IN"]("ell", "hello")
```

### `STRING["IS_FOUND_AT_START"](search, text)`

Checks whether `text` starts with `search`.

```stult
STRING["IS_FOUND_AT_START"]("he", "hello")
```

### `STRING["IS_FOUND_AT_END"](search, text)`

Checks whether `text` ends with `search`.

```stult
STRING["IS_FOUND_AT_END"]("lo", "hello")
```

### `STRING["REPLACE"](text, old, new)`

Replaces every occurrence of `old` with `new`.

```stult
STRING["REPLACE"]("one two one", "one", "three")
```

### `STRING["SPLIT"](text, separator)`

Splits text into an array of strings.

```stult
STRING["SPLIT"]("a,b,c", ",")
```

### `STRING["JOIN"](array, separator)`

Joins array elements into a string.

```stult
STRING["JOIN"]({"a", "b", "c"}, ",")
```

String elements are used directly.

Non-string elements are converted with their printed representation.

## ARRAY

```stult
ARRAY : STD["TYPE"]["ARRAY"]
```

### `ARRAY["APPEND"](array, ...values)`

Appends one or more values to an array.

```stult
items : {}

ARRAY["APPEND"](items, "a", "b", "c")
```

Returns `_`.

The first argument must be an array.

## MAP

```stult
MAP : STD["TYPE"]["MAP"]
```

### `MAP["KEYS"](map)`

Returns an array of map keys sorted lexicographically.

```stult
keys : MAP["KEYS"]({"b": 2, "a": 1})
```

Returns `_` when the value is not a map.

### `MAP["VALUES"](map)`

Returns an array of map values sorted by key order.

```stult
values : MAP["VALUES"]({"b": 2, "a": 1})
```

Returns `_` when the value is not a map.

## COLLECTION

```stult
COLLECTION : STD["TYPE"]["COLLECTION"]
```

Collections are arrays, maps and strings.

### `COLLECTION["SIZE"](collection)`

Returns the size of a collection.

```stult
COLLECTION["SIZE"]({"a", "b", "c"})
COLLECTION["SIZE"]({"name": "example"})
COLLECTION["SIZE"]("hello")
```

For arrays, size is the number of elements.

For maps, size is the number of entries.

For strings, size is the number of runes.

Returns `_` for non-collections.

### `COLLECTION["IS_EMPTY"](collection)`

Checks whether a collection is empty.

```stult
COLLECTION["IS_EMPTY"]({})
COLLECTION["IS_EMPTY"]("")
```

Returns `_` for non-collections.

### `COLLECTION["HAS"](collection, key)`

Checks whether a collection contains a key or index.

```stult
COLLECTION["HAS"]({"name": "example"}, "name")
COLLECTION["HAS"]({"a", "b"}, 1)
COLLECTION["HAS"]("cat", 0)
```

For maps, the key must be a string.

For arrays and strings, the key must be a valid numeric index.

Returns `_` for non-collections.

### `COLLECTION["CLEAR"](collection)`

Removes all contents from a mutable collection.

```stult
items : {"a", "b"}
COLLECTION["CLEAR"](items)
```

For arrays, this removes all elements.

For maps, this removes all entries.

For strings, this removes all characters.

Returns `_`.

## DATA

```stult
DATA : STD["DATA"]
```

`DATA` contains helpers for converting Stult values to and from data formats.

```stult
DATA["CSV"]
DATA["JSON"]
DATA["STULTON"]
```

## CSV

```stult
CSV : STD["DATA"]["CSV"]
```

### `CSV["NEW"](rows)`

Encodes an array of row arrays as CSV text.

```stult
rows : {
	{"name", "score"}
	{"a", 10}
	{"b", 20}
}

text : CSV["NEW"](rows)
```

Fields are converted to strings when possible.

Returns a string.

### `CSV["PARSE"](text)`

Parses CSV text into an array of row arrays.

```stult
rows : CSV["PARSE"]("name,score\na,10\nb,20\n")
```

Returns an array of arrays of strings.

### `CSV["IS_VALID"](text)`

Checks whether text is valid CSV.

```stult
CSV["IS_VALID"]("name,score\na,10\n")
```

Returns a boolean.

Rows may have different field counts.

## JSON

```stult
JSON : STD["DATA"]["JSON"]
```

### `JSON["NEW"](value)`

Encodes a Stult value as JSON text.

```stult
text : JSON["NEW"]({
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

### `JSON["PARSE"](text)`

Parses JSON text into a Stult value.

```stult
value : JSON["PARSE"]("{\"name\":\"example\",\"active\":true}")
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

### `JSON["IS_VALID"](text)`

Checks whether text is valid JSON.

```stult
JSON["IS_VALID"]("{\"ok\":true}")
```

Returns a boolean.

## STULTON

```stult
STULTON : STD["DATA"]["STULTON"]
```

### `STULTON["NEW"](value)`

Encodes a Stult value as STULTON text.

```stult
text : STULTON["NEW"]({
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

### `STULTON["PARSE"](text)`

Parses STULTON text into a Stult value.

```stult
value : STULTON["PARSE"]("{\"active\": \\/}")
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

### `STULTON["IS_VALID"](text)`

Checks whether text is valid STULTON data.

```stult
STULTON["IS_VALID"]("{\"active\": \\/}")
```

Returns a boolean.