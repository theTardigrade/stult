# Examples (Tests)

This document lists small public Stult programs used to exercise specific language or parser behaviours.

These examples live in [`../examples/tests/`](../examples/tests/) and are run by the Go test suite.

They are still valid Stult examples, but they are written primarily as regression fixtures rather than tutorials.

For documentation describing the ordinary examples, see [examples.md](examples.md).

## Collection behaviour

- [Frozen collections](../examples/tests/frozen_collections.stult)  
  Checks that frozen arrays, maps and strings cannot be internally modified and that `IS_FROZEN` reports their frozen state.

- [Deep frozen aliases](../examples/tests/deep_frozen_aliases.stult)  
  Checks that `FREEZE` deeply freezes nested collections and that aliases to nested collections observe the frozen state.

- [Collection loop parameters](../examples/tests/collection_loop_parameters.stult)  
  Checks that collection loops provide value, key, collection and position parameters correctly.

## Grouped expressions

- [Grouped expressions in array](../examples/tests/grouped_expressions_in_array.stult)  
  Checks that an array literal may begin with a grouped expression and that brace literals beginning with `{ (` are not automatically treated as function literals.

- [Grouped conditional and grouped loop logic](../examples/tests/grouped_conditional_and_grouped_loop_logic.stult)  
  Checks that grouped boolean expressions at the start of conditionals and loops are parsed correctly, including a grouped conditional that begins with touching `((` and a grouped loop whose body begins with a conditional.

- [Grouped return expression](../examples/tests/grouped_return_expression.stult)  
  Checks that a function return expression may begin with a grouped expression using touching `((` without being mistaken for a loop.

## Literal parsing

- [Array starts with function literal](../examples/tests/array_starts_with_function_literal.stult)  
  Checks that arrays may contain function literals as their first elements, including multiline and single-line arrays.

- [Nested literals in array](../examples/tests/nested_literals_in_array.stult)  
  Checks that arrays may contain empty maps, empty arrays, non-empty maps, nested arrays and function literals as elements.

## Standard library

- [Assertions](../examples/tests/assertions.stult)  
  Checks that `STD["ASSERT"]["TRUE"]` and `STD["ASSERT"]["EQUAL"]` succeed for valid assertions and return void where appropriate.