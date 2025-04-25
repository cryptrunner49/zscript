# ZScript Basic Usage Guide

Welcome to the ZScript Basic Usage Guide! ZScript is a versatile scripting language that supports modern programming constructs, a variety of built-in features, and Unicode/emoji identifiers. This guide walks through the language features with practical examples and explanations.

---

## Table of Contents

1. [Variables](#1-variables)
2. [Control Flow](#2-control-flow)
3. [Loops](#3-loops)
4. [Closures](#4-closures)
5. [Fibonacci Recursive](#5-fibonacci-recursive)
6. [Fibonacci Iterative](#6-fibonacci-iterative)
7. [Structs](#7-structs)
8. [Arrays](#8-arrays)
9. [Maps](#9-maps)
10. [File Operations](#10-file-operations)
11. [Modules](#11-modules)
12. [Import](#12-import)
13. [Additional Features](#13-additional-features)
    - [13.1. Input Handling](#131-input-handling)
    - [13.2. String Formatting](#132-string-formatting)
    - [13.3. Shadowing](#133-shadowing)
14. [Operators](#14-operators)
    - [14.1. Arithmetic Operators](#141-arithmetic-operators)
    - [14.2. Assignment Operators](#142-assignment-operators)
    - [14.3. Comparison Operators](#143-comparison-operators)
    - [14.4. Logical Operators](#144-logical-operators)
    - [14.5. Unary Operators](#145-unary-operators)
    - [14.6. Force Operator](#146-force-operator)
    - [14.7. Operator Precedence](#147-operator-precedence)
15. [Unicode Support](#15-unicode-support)
16. [Native Functions](#16-native-functions)

---

## 1. Variables

Variables can be declared using the `var` keyword or by assigning a value directly. A variable can hold a number, string, boolean, or even null. Variables are available within the block where they’re created, and they also exist in any inner blocks (child scopes), unless a variable with the same name is declared there. In that case, the inner variable takes priority. ZScript figures out the variable type automatically based on the value you assign.

```z
// Number
var num = 42
println("Number:", num)

// String
var message = "Hello, World!"
println("Message:", message)

// Boolean
isTrue = true
println("Boolean:", isTrue)

// Null
nothing = null
println("Nothing:", nothing)
```

---

## 2. Control Flow

The `if`, `else`, and `else if` (written as `|`) keywords allow for conditional logic. Conditions must be inside parentheses, and blocks are defined using colons `:`. There are no curly braces in ZScript. You can write single-line or multi-line blocks depending on your needs. Nested conditions are supported, and multi-branch checks can be done using the `|` syntax.

```z
// Single-line if
var n = 2
if (n < 1):
    println(n + 1)

// Single-line if-else
if (n < 2):
    println(n + 2)
else:
    println(n * 2)

// Multi-line if with nested if
var color = "blue"
if (color == "blue"):
    println("Color blue detected")
    if (n > 0):
        println("Positive number:", n)

// Multi-line if-else-if chain
if (color == "red"):
    println("Color red detected")
| (color == "blue"):
    println("Color blue detected")
| (color == "yellow"):
    println("Color yellow detected")
else:
    println("Unknown color:", color)
```

---

## 3. Loops

The `while` keyword creates condition-based loops, using colons `:` for blocks. The `for` keyword iterates with an initializer, condition, and increment expression, followed by a colon. The `iter` keyword iterates over arrays, using `var` to declare the loop variable and `in` to specify the array.

```z
// While Loop Iteration
println("--- While Loop ---")
var index = 0
while (index < len(arr)):
    println("Element at index", index, ":", arr[index])
    index = index + 1

// For Loop Iteration
println("--- For Loop ---")
var arr = [10, 20, 30]
for (var i = 0; i < len(arr); i++):
    println("Element at index", i, ":", arr[i])

// Iterator Loop
println("--- Iterator Loop ---")
iter (var item in [10, 20, 30]):
    println("Item:", item)
```

---

## 4. Closures

The `func` keyword defines functions, which can capture outer variables, forming closures. Closures retain access to their lexical scope, enabling stateful behavior across calls.

```z
func makeCounter():
    var value = 0
    func increment():
        value = value + 1
        return value
    return increment

var counter = makeCounter()
println("First call:", counter())  // Outputs: 1
println("Second call:", counter()) // Outputs: 2
```

---

## 5. Fibonacci Recursive

The `func` keyword with `if` and `return`, supports recursive functions. A function calls itself with modified arguments until a base case terminates recursion.

```z
func fib(n):
    if (n < 2):
        return n
    return fib(n - 2) + fib(n - 1)

var start = clock()
println("Fibonacci(16):", fib(16))
printf("Time taken: %v seconds\n", clock() - start)
```

---

## 6. Fibonacci Iterative

The `for` loop, `var` declarations, and assignment (`=`) enable iterative algorithms. Variables are updated in a loop, providing efficient computation without recursion.

```z
func fib(n):
    if (n < 2):
        return n
    var a = 0
    var b = 1
    for (var i = 2; i <= n; i = i + 1):
        var temp = a + b
        a = b
        b = temp
    return b

var start = clock()
println("Fibonacci(16):", fib(16))
printf("Time taken: %v seconds\n", clock() - start)
```

---

## 7. Structs

The `struct` keyword defines custom types with fields initialized using `=`. Instances are created with curly braces `{}`, and fields are accessed with dot notation (`.`). The force operator `!{}` allows initializing structs with fields not defined in the struct, overriding defaults.

```z
struct Point:
    x = 10
    y = 22

var p = Point{}  // Creates a Point instance with defaults x=10, y=22
println(p)

var p2 = Point{x = 1, y = 2}  // Creates a Point instance with x=1, y=2
println(p2)

struct Vec3
var v3 = Vec3!{x = 1, y = 2, z = 3}  // Forces creation with fields x, y, z
println(v3)
```

---

## 8. Arrays

Arrays are defined with square brackets `[]`, and elements are accessed via indices (e.g., `arr[i]`). Built-in functions like `push` and `pop` dynamically manage array contents.

```z
var arr = [1, 2, 3, 4, 5]
println("Original array:", array_to_string(arr))
push(arr, 6)
println("After push(6):", array_to_string(arr))
println("Popped:", pop(arr))

for (var i = 0; i < len(arr); i = i + 1):
    println("Element", i, ":", arr[i])
```

---

## 9. Maps

Maps are defined with curly braces `{}`, using key-value pairs (e.g., `"key": value`). Keys are accessed with square brackets `[]`, and built-in functions manage entries.

```z
var map = { "name": "Alice", "age": 30 }
println("Name:", map["name"])
println("Age:", map["age"])
map["age"] = 31
println("Updated age:", map["age"])

var m = {"a": 1, "b": 2}
map_remove(m, "a")
println("After removing 'a':", m)
println("Contains key 'b':", map_contains_key(m, "b"))
```

---

## 10. File Operations

Built-in functions `write_file` and `read_file` perform file operations, accepting filename and content arguments. They enable data persistence and retrieval.

```z
var filename = "test.txt"
var content = "This is a file handling example.\nDemonstrating how to read and write files."
write_file(filename, content)
println("Wrote to file:", filename)

var readContent = read_file(filename)
println("Read from file:", readContent)
```

---

## 11. Modules

The `mod` keyword defines modules, organizing code into namespaces. Nested modules and variables are supported, accessed with dot notation (e.g., `Module.Submodule`).

```z
mod Geometry:
    mod Shapes:
        func area(r):
            return 3.14 * r * r
    var PI = 3.14
```

---

## 12. Import

The `import` keyword loads external modules by filename (e.g., `"file.z"`). The `mod as` syntax aliases modules, simplifying access to their contents.

```z
import "geometry.z"

println("Circle Area:", Geometry.Shapes.area(5))
println("Circle Perimeter:", Geometry.PI * 5 * 2)

mod Geometry.Shapes as Shapes
println("Aliased area:", Shapes.area(5))
```

---

## 13. Additional Features

### 13.1. Input Handling

The `scanln` function reads a line of user input as a string, enabling interactive programs by capturing dynamic input.

```z
println("Enter a sentence:")
var input = scanln()
println("You entered:", input)
```

### 13.2. String Formatting

The `sprintf` function formats strings with placeholders (e.g., `%v`), combining values into a single string for output or further processing.

```z
var name = "Alice"
var age = 25
var formatted = sprintf("Name: %v, Age: %v", name, age)
println("Formatted:", formatted)
```

### 13.3. Shadowing

Shadowing in ZScript allows a variable declared with `var` to override a previous variable with the same name, either in the same scope or in an inner scope. All variables are mutable, allowing reassignment and type changes. In the same scope, a new `var` declaration shadows the earlier one, with the last declaration taking precedence. In different scopes, an inner scope variable shadows the outer scope variable without modifying it, leaving the outer variable intact outside the inner scope.

```z
// Shadowing in the same scope
var hue = "Red"
var hue = "Blue"  // Shadows previous 'hue'
println("Same scope hue:", hue)  // Outputs: Blue

var value = 100
var value = "Text"  // Shadows with a different type
println("Shadowed value:", value)  // Outputs: Text

// Multiple shadowing in the same scope
var count = 1
var count = 2
var count = 3
println("Final count:", count)  // Outputs: 3

// Shadowing in different scopes
var color = "Yellow"
{
    var color = "Green"  // Shadows outer 'color' in inner scope
    println("Inner color:", color)  // Outputs: Green
    var color = "Purple"  // Shadows again in the same inner scope
    println("Inner shadowed color:", color)  // Outputs: Purple
}
println("Outer color:", color)  // Outputs: Yellow, unaffected by inner scope
```

---

## 14. Operators

ZScript supports a variety of operators for arithmetic, assignment, comparison, logical operations, and more. Below are the supported operators, organized by category, followed by their precedence rules.

### 14.1. Arithmetic Operators

- `+` (Addition)
- `-` (Subtraction)
- `*` (Multiplication)
- `/` (Division)
- `%` (Modulus)
- `**` (Exponentiation)
- `/_` (Integer division)
- `%%` (Percentage)

**Example**:

```z
var a = 10
var b = 3
println("Addition:", a + b)           // 13
println("Subtraction:", a - b)        // 7
println("Multiplication:", a * b)     // 30
println("Division:", a / b)           // 3.333...
println("Modulus:", a % b)            // 1
println("Exponentiation:", a ** 2)    // 100
println("Integer Division:", a /_ b)  // 3
println("Percentage:", 25 %% 1000)    // 250
```

### 14.2. Assignment Operators

- `=` (Assignment)

**Example**:

```z
var x = 5
x = x + 1
println("x:", x)  // 6
```

### 14.3. Comparison Operators

- `==` (Equal to)
- `!=` (Not equal to)
- `>` (Greater than)
- `<` (Less than)
- `>=` (Greater than or equal to)
- `<=` (Less than or equal to)

**Example**:

```z
var a = 5
var b = 3
println("Equal:", a == b)          // false
println("Not Equal:", a != b)      // true
println("Greater Than:", a > b)    // true
println("Less Than:", a < b)       // false
println("Greater or Equal:", a >= b) // true
println("Less or Equal:", a <= b)    // false
```

### 14.4. Logical Operators

- `and` (Logical AND)
- `or` (Logical OR)
- `!` (Logical NOT)

**Example**:

```z
var t = true
var f = false
println("AND:", t and f)  // false
println("OR:", t or f)    // true
println("NOT:", !t)       // false
```

### 14.5. Unary Operators

- `++` (Increment)
- `--` (Decrement)
- `-` (Unary negation)

**Example**:

```z
var x = 5
println("Increment:", ++x)  // 6
println("Decrement:", --x)  // 5
println("Negation:", -x)    // -5
```

### 14.6. Force Operator

- `!{}` (Force struct instantiation with custom fields)

**Example**:

```z
struct Vec3
var v = Vec3!{x = 1, y = 2, z = 3}  // Forces creation with fields x, y, z
println(v)
```

### 14.7. Operator Precedence

Operator precedence determines the order in which operators are evaluated. The table below lists the precedence levels from highest to lowest, using descriptive categories for clarity.

| Precedence Level | Operators |
|------------------|-----------|
| Literals         | `number`, `string`, `boolean`, `null`, `( )` (grouped expressions) |
| Calls            | `.` (field access), `[]` (subscripting), `()` function calls |
| Unary            | `++`, `--`, `-` (negation), `!` (not) |
| Multiplicative   | `*`, `/`, `%`, `**`, `/_`, `%%` |
| Additive         | `+`, `-` |
| Comparison       | `>`, `<`, `>=`, `<=` |
| Equality         | `==`, `!=` |
| LogicalAnd       | `and` |
| LogicalOr        | `or` |
| Assignment       | `=` |

---

## 15. Unicode Support

ZScript supports Unicode and emoji characters in identifiers and strings, using UTF-8 encoding. This allows multilingual and expressive variable names.

```z
var π = 3.14159
println("Value of π:", π)

var 🔥 = "Fire emoji"
println("Emoji variable 🔥:", 🔥)

var greeting = "こんにちは世界"
println("Greeting in Japanese:", greeting)

struct 数学:
    半径 = 5

var 円 = 数学{}
println("半径 (radius):", 円.半径)
```

---

## 16. Native Functions

Built-in functions, grouped by category (e.g., String, Array), provide system-level operations. They are invoked directly with arguments, handling tasks like debugging or date manipulation.

```z
// === String Functions ===
var str = "  Hello, ZScript!  "
println("Original string:", str)
println("To string:", to_str(42))                    // Convert number to string
var chars = to_chars(str)                           // Convert string to array of characters
println("Characters:", array_to_string(chars))
println("Char at 2:", char_at(str, 2))              // Get character at index
println("Substring:", substring(str, 2, 7))         // Get substring
println("Index of 'ZScript':", str_index_of(str, "ZScript")) // First occurrence
println("Last index of 'l':", str_last_index_of(str, "l")) // Last occurrence
println("Contains 'ZScript':", str_contains(str, "ZScript")) // Check if substring exists
println("Starts with 'Hello':", starts_with(str, "Hello")) // Check prefix
println("Ends with '!':", ends_with(str, "!"))       // Check suffix
println("Uppercase:", to_upper(str))                 // Convert to uppercase
println("Lowercase:", to_lower(str))                 // Convert to lowercase
println("Trimmed:", trim(str))                      // Remove whitespace
var split = split(str, ",")                         // Split by delimiter
println("Split by ',':", array_to_string(split))
println("Replace 'ZScript' with 'World':", replace(str, "ZScript", "World")) // Replace substring
println("String length:", str_length(str))           // Get string length

// === Array Functions ===
var arr = [1, 2, 2, 3]
println("Original array:", array_to_string(arr))
println("Array length:", len(arr))                   // Get array length
push(arr, 4)                                        // Push element
println("After push(4):", array_to_string(arr))
var popped = pop(arr)                               // Pop element
println("Popped:", popped, "Array:", array_to_string(arr))
array_sort(arr)                                     // Sort array
println("Sorted:", array_to_string(arr))
var split_arr = array_split(arr, 2)                  // Split by separator
println("Split by 2:", array_to_string(split_arr))
var arr2 = [5, 6]
var joined = array_join(arr, arr2)                   // Join arrays
println("Joined:", array_to_string(joined))
array_sorted_push(arr, 3)                           // Insert into sorted array
println("Sorted push(3):", array_to_string(arr))
println("Linear search 2:", array_linear_search(arr, 2)) // Linear search
println("Binary search 3:", array_binary_search(arr, 3)) // Binary search
println("Index of 2:", index_of(arr, 2))            // First occurrence
println("Last index of 2:", last_index_of(arr, 2))  // Last occurrence
println("Contains 3:", array_contains(arr, 3))      // Check if element exists
array_reverse(arr)                                  // Reverse array
println("Reversed:", array_to_string(arr))
array_remove(arr, 2)                                // Remove first occurrence
println("After remove(2):", array_to_string(arr))
array_clear(arr)                                    // Clear array
println("Cleared:", array_to_string(arr))

// === Iterator Functions ===
var iter_arr = [10, 20, 30]
var iter = array_iter(iter_arr)                     // Create iterator
while (!iter_done(iter)):
    println("Iterator value:", iter_value(iter))     // Get current value
    iter_next(iter)                                 // Move to next

// === Map Functions ===
var map = {"a": 1, "b": 2}
println("Map:", to_str(map))
map_remove(map, "a")                                // Remove key
println("After remove 'a':", to_str(map))
println("Contains key 'b':", map_contains_key(map, "b")) // Check key
println("Contains value 2:", map_contains_value(map, 2)) // Check value
println("Map size:", map_size(map))                 // Get map size
var keys = map_keys(map)                            // Get all keys
println("Keys:", array_to_string(keys))
var values = map_values(map)                        // Get all values
println("Values:", array_to_string(values))
map_clear(map)                                      // Clear map
println("Cleared map:", to_str(map))

// === Date Functions ===
var date = Date(2023, 10, 15)                       // Create date
println("Date:", to_str(date))
println("Current date:", to_str(date_now()))        // Current date
var parsed_date = date_parse_datetime("2024-01-01") // Parse date
println("Parsed date:", to_str(parsed_date))
println("Formatted date:", date_format_datetime(date, "2006-Jan-02")) // Format
var added_date = date_add_datetime(date, 1, 2, 3)   // Add time
println("Added date:", to_str(added_date))
var sub_date = date_subtract_datetime(date, 1, 2, 3) // Subtract time
println("Subtracted date:", to_str(sub_date))
println("Year:", date_get_component(date, "year"))  // Get component
var set_date = date_set_component(date, "year", 2025) // Set component
println("Set year:", to_str(set_date))
println("Add 5 days:", to_str(date_add_days(date, 5))) // Add days
println("Subtract 5 days:", to_str(date_subtract_days(date, 5))) // Subtract days

// === Time Functions ===
var time = Time(14, 30, 0)                          // Create time
println("Time:", to_str(time))
println("Current time:", to_str(time_now()))        // Current time
var parsed_time = time_parse("15:04:05")            // Parse time
println("Formatted time:", time_format(time, "15:04")) // Format
var added_time = time_add(time, 1, 30, 0)           // Add time
println("Added time:", to_str(added_time))
var sub_time = time_subtract(time, 1, 30, 0)        // Subtract time
println("Subtracted time:", to_str(sub_time))
println("Timezone:", time_get_timezone(time))       // Get timezone
var converted = time_convert_timezone(time, "America/New_York") // Convert timezone
println("Converted timezone:", to_str(converted))

// === DateTime Functions ===
var dt = DateTime(2023, 10, 15, 14, 30, 0)          // Create datetime
println("DateTime:", to_str(dt))
println("Current datetime:", to_str(datetime_now())) // Current datetime
var parsed_dt = datetime_parse("2023-10-15 14:30:00") // Parse datetime
println("Parsed datetime:", to_str(parsed_dt))
println("Formatted datetime:", datetime_format(dt, "2006-Jan-02 15:04")) // Format
var added_dt = datetime_add(dt, 1, 2, 3, 4, 5, 6)   // Add datetime
println("Added datetime:", to_str(added_dt))
var sub_dt = datetime_subtract(dt, 1, 2, 3, 4, 5, 6) // Subtract datetime
println("Subtracted datetime:", to_str(sub_dt))
println("Hour:", datetime_get_component(dt, "hour")) // Get component
var set_dt = datetime_set_component(dt, "hour", 16) // Set component
println("Set hour:", to_str(set_dt))
println("Add 5 days:", to_str(datetime_add_days(dt, 5))) // Add days
println("Subtract 5 days:", to_str(datetime_subtract_days(dt, 5))) // Subtract days

// === Random Functions ===
var colors = ["Red", "Blue", "Green"]
shuffle(colors)                                     // Shuffle array
println("Shuffled colors:", array_to_string(colors))
var rand_num = random_between(1, 10)                // Random number
println("Random number:", rand_num)
var rand_str = random_string(8)                     // Random string
println("Random string:", rand_str)

// === Output Functions ===
print("Hello, ")                                    // Print without newline
println("ZScript!")                                 // Print with newline
printf("Number: %d, String: %s\n", 42, "ZScript")   // Formatted print

// === Input Functions ===
var scanned = scan()                             // Read input as array
println("Scanned:", array_to_string(scanned))
var line = scanln()                             // Read line
println("Line:", line)
var formatted = scanf("%s")                     // Read formatted input
println("Formatted input:", formatted)

// === Formatting Functions ===
var formatted = sprintf("Value: %d", 42)            // Format string
println("Formatted:", formatted)
var err_msg = errorf("Error: %s", "Invalid")        // Format error
println("Error message:", err_msg)

// === File Operations ===
write_file("test.txt", "Hello, ZScript!")           // Write to file
var content = read_file("test.txt")                 // Read from file
println("File content:", content)

// === Utility Functions ===
var num = parse_int("123")                          // Parse string to int
println("Parsed int:", num)

// === Type Functions ===
println("Type of 42:", get_runtype(42))             // Get runtime type

// === Other Functions ===
var time = clock()                                  // Get current time in seconds
println("Current time (seconds):", time)

// === Debug Functions ===
enable_debug()           // Turn on bytecode debug printing
enable_trace()          // Turn on instruction-level execution tracing
println("Debug enabled")
disable_debug()         // Turn off bytecode debug printing
disable_trace()         // Turn off instruction-level execution tracing
println("Debug disabled")
```
