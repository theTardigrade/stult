# Examples (Tests)

This document lists small public Stult programs used to exercise specific language or parser behaviours.

These examples live in [`../examples/tests/`](../examples/tests/) and are run by the Go test suite.

They are still valid Stult examples, but they are written primarily as regression fixtures rather than tutorials.

For documentation describing the ordinary examples, see [examples.md](examples.md).

## Contents

* [Collection behaviour](#collection-behaviour)
* [Loops](#loops)
* [Grouped expressions](#grouped-expressions)
* [Error handling](#error-handling)
* [Literal parsing](#literal-parsing)
* [Maps](#maps)
* [Functions](#functions)
* [Numbers](#numbers)
* [Mathematics](#mathematics)
* [Standard library](#standard-library)

## Collection behaviour

- [Frozen collections](../examples/tests/frozen_collections.stult)  
  Checks that frozen arrays, maps and strings cannot be internally modified and that `IS_FROZEN` reports their frozen state.

- [Deep frozen aliases](../examples/tests/deep_frozen_aliases.stult)  
  Checks that `FREEZE` deeply freezes nested collections and that aliases to nested collections observe the frozen state.

- [Cloning collections](../examples/tests/collection_clone.stult)  
  Checks that `CLONE` deeply clones collection graphs, preserves aliases and cycles, returns mutable collections, preserves map-entry mutability and reuses function closure values.

- [Collection loop parameters](../examples/tests/collection_loop_parameters.stult)  
  Checks that collection loops provide value, key, collection and position parameters correctly.

## Loops

- [Looping over functions](../examples/tests/function_loop.stult)  
  Checks that user-defined functions can be used as loop sources, including indexed generators, zero-argument generators, optional index parameters, ignored loop values and ordinary break behaviour.

- [Range loop optimisation](../examples/tests/range_loop_optimization.stult)  
  Checks that direct single-range loop sources can be streamed without changing observable behaviour, including very large integer bounds, descending ranges, stepped ranges and the collection-parameter fallback.

## Grouped expressions

- [Grouped expressions in array](../examples/tests/grouped_expressions_in_array.stult)  
  Checks that an array literal may begin with a grouped expression and that brace literals beginning with `{ (` are not automatically treated as function literals.

- [Grouped conditional and grouped loop logic](../examples/tests/grouped_conditional_and_grouped_loop_logic.stult)  
  Checks that grouped boolean expressions at the start of conditionals and loops are parsed correctly, including a grouped conditional that begins with touching `((` and a grouped loop whose body begins with a conditional.

- [Grouped return expression](../examples/tests/grouped_return_expression.stult)  
  Checks that a function return expression may begin with a grouped expression using touching `((` without being mistaken for a loop.

- [Conditional expression](../examples/tests/conditional_expression.stult)  
  Checks that `(CONDITION)?(WHEN_TRUE, WHEN_FALSE)` evaluates exactly one selected branch and returns its value.

- [Match expression](../examples/tests/match_expression.stult)  
  Checks that `(SUBJECT)?{ ... }` matches explicit scalar-literal arms before falling back to `_`.

## Error handling

- [Try-catch](../examples/tests/try_catch.stult)  
  Checks that try-catch catches runtime errors, supports an optional error-message parameter, skips the catch block when the try block succeeds, lets catch-block errors escape, catches the nearest handler and preserves break and early-return control flow.

## Literal parsing

- [Array starts with function literal](../examples/tests/array_starting_with_function_literal.stult)  
  Checks that arrays may contain function literals as their first elements, including multiline and single-line arrays.

- [Nested literals in array](../examples/tests/nested_literals_in_array.stult)  
  Checks that arrays may contain empty maps, empty arrays, non-empty maps, nested arrays and function literals as elements.

## Maps

- [Dot access](../examples/tests/dot_access_for_maps.stult)  
  Checks that `object.key` behaves like `object["key"]` for map string keys, including chained access, assignment, compound assignment and standard-library paths.

- [Leading dot access](../examples/tests/leading_dot_access.stult)  
  Checks that `.field` map keys create string-key entries, that functions written inside maps can use `.field` to read and update the nearest surrounding map, and that captured functions keep their original map receiver when called through another binding.

## Functions

- [Optional parameters](../examples/tests/optional_parameters.stult)  
  Checks that optional user-function parameters receive void when omitted and that variadic parameters still receive an empty array when no remaining arguments are supplied.

## Numbers

- [Huge numbers](../examples/tests/huge_numbers.stult)  
  Checks that very large whole-number values can be compared and changed without losing their whole-number precision, while decimal values remain bounded.

- [Percentages](../examples/tests/percentage_numbers.stult)  
  Checks that percentage-suffixed number literals divide their literal value by one hundred and work in ordinary arithmetic.

- [Number formatting](../examples/tests/number_formatting.stult)  
  Checks fixed decimal formatting with requested decimal places and scientific-notation formatting with requested significant digits.

## Mathematics

- [Modulo versus remainder](../examples/tests/modulo_and_remainder.stult)  
  Checks the difference between mathematical modulo and truncating remainder, including negative operands.

- [Random-number generation](../examples/tests/math_rand.stult)  
  Checks random number, integer, choice and shuffle helpers under `STD["MATH"]["RAND"]`.

## Standard library

- [Assertions](../examples/tests/assertions.stult)  
  Checks that `STD["ASSERT"]["TRUE"]`, `STD["ASSERT"]["FALSE"]` and `STD["ASSERT"]["EQUAL"]` succeed for valid assertions, accept optional messages and return void where appropriate.

- [File path namespace](../examples/tests/file_path_namespace.stult)  
  Checks that file-path helpers live under `STD.FILE.PATH` and that the old top-level `STD.PATH` map is no longer present.
