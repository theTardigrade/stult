# Examples

These examples show how to use Stult for small scripts, data processing, loops, functions, manifests and bundled executables.

## Collections

- [Shopping list array](../examples/shopping_list_array.stult)  
  Creates arrays, appends new values to a mutable array and prints the updated array size.

- [Iterate over an array](../examples/iterate_over_array.stult)  
  Iterates over an array of names and prints each value with its position in the list.

- [Word count](../examples/word_count.stult)  
  Counts words in a string using string iteration, mutable counters and outer-scope assignment.

- [Average map values](../examples/average_map_values.stult)  
  Calculates the average of decimal values stored in a map and prints the result as a percentage.

- [Calculate circle area from a map](../examples/calculate_circle_area_from_map.stult)  
  Stores a circle radius in a map, calculates the circle area with a function stored in that map and recalculates it after changing the radius.

## Control flow

- [Countdown loop](../examples/countdown_loop.stult)  
  Counts down with a while-style loop and runs an after-loop block when complete.

- [Compare numbers with conditionals](../examples/compare_numbers_with_conditionals.stult)  
  Compares two numbers and prints a different message depending on whether one number is lower, equal or higher than the other.

- [Break and early return](../examples/break_and_early_return.stult)  
  Demonstrates early return from functions and breaking out of loops.

## Data formats

- [STULTON encoding and parsing](../examples/stulton.stult)  
  Encodes a Stult map as STULTON text, validates it, parses it back into a Stult value and encodes it again.

## Projects

- [Boolean manifest project](../examples/bool/)  
  Shows how multiple files run in the order specified in a manifest file.

## Standard library

- [Standard library overview](../examples/standard_library_overview.stult)  
  Recursively prints the maps and functions available in the Stult standard library.