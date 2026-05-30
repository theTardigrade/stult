# Examples

These examples show how to use Stult for small scripts, data processing, loops, functions, manifests and bundled executables.

## Data formats

- [STULTON encoding and parsing](../examples/stulton.stult)
  Encodes a Stult map as STULTON text, validates it, parses it back into a Stult value and encodes it again.

## Collections

- [Word count](../examples/word_count.stult)  
  Counts words in a string using string iteration, mutable counters and outer-scope assignment.

- [Average map values](../examples/average_map_values.stult)  
  Calculates the average of decimal values stored in a map and prints the result as a percentage.

## Control flow

- [Countdown loop](../examples/countdown_loop.stult)  
  Counts down with a while-style loop and runs an after-loop block when complete.

- [Break and early return](../examples/break_and_early_return.stult)  
  Demonstrates early return from functions and breaking out of loops.

## Projects

- [Boolean manifest project](../examples/bool/)  
  Shows how multiple files run in the order specified in a manifest file.