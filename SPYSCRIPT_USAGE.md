# SPYScript Usage Guide

Welcome to the SPYScript Usage Guide! SPYScript is a versatile scripting language that supports modern programming constructs, Unicode and emoji identifiers, and a variety of built-in features. This README provides practical examples and explanations of SPYScript's core features, making it an ideal starting point for developers. Each section includes a code snippet from a corresponding `.spy` file, followed by a detailed explanation.

---

## Table of Contents

1. [Variables](#1-variables)
2. [Closures](#2-closures)
3. [Fibonacci Recursive](#3-fibonacci-recursive)
4. [Fibonacci Iterative](#4-fibonacci-iterative)
5. [Structs and Control Flow](#5-structs-and-control-flow)
6. [Arrays](#6-arrays)
7. [File Operations](#7-file-operations)
8. [Native Functions](#8-native-functions)
9. [Modules](#9-modules)
10. [Additional Features](#10-additional-features)

---

## 1. Variables (`variables.spy`)

SPYScript supports variables of various types, including numbers, strings, booleans, and null, with flexible naming using Unicode and emojis.

```spy
// Number
var num = 42;
println("Number:", num);  // Outputs: Number: 42

// String
var message = "Hello, SPYScript!";
println("Message:", message);  // Outputs: Message: Hello, SPYScript!

// Boolean
var isTrue = true;
println("Boolean:", isTrue);  // Outputs: Boolean: true

// Null
var nothing = null;
println("Null:", nothing);  // Outputs: Null: null

// Unicode and Emoji Names
var π = 3.14;
println("Pi (π):", π);  // Outputs: Pi (π): 3.14

var 挨拶 = "こんにちは";
println("Greeting (挨拶):", 挨拶);  // Outputs: Greeting (挨拶): こんにちは

var 🔢 = 100;
println("Number (🔢):", 🔢);  // Outputs: Number (🔢): 100
```

**Explanation**:  

- Use `var` to declare variables with any value type: numbers, strings, booleans, or `null`.
- SPYScript’s unique feature is its support for Unicode (e.g., `π`, `挨拶`) and emoji (e.g., `🔢`) variable names, making it highly expressive.
- The `println` function outputs values with optional labels for clarity.

---

## 2. Closures (`closures.spy`)

Closures in SPYScript allow inner functions to access variables from outer scopes, enabling powerful functional programming patterns.

```spy
fn outer() {
    var a = 1;
    var b = 2;
    fn middle() {
        var c = 3;
        var d = 4;
        fn inner() {
            println("Sum:", a + c + b + d);  // Accesses variables from outer scopes
        }
        inner();
    }
    middle();
}
outer();  // Outputs: Sum: 10
```

**Explanation**:  

- The `inner` function captures `a` and `b` from `outer` and `c` and `d` from `middle`, demonstrating closure scope.
- Calling `outer()` executes the nested functions, summing the variables to output `10` (1 + 3 + 2 + 4).
- SPYScript’s `fn` keyword defines functions, which can be nested arbitrarily.

---

## 3. Fibonacci Recursive (`fibonacci_recursive.spy`)

This example computes the Fibonacci sequence recursively and measures execution time.

```spy
fn fib(n) {
    if (n < 2) return n;
    return fib(n - 2) + fib(n - 1);
}

var start = clock();
println("Fibonacci(16):", fib(16));  // Outputs: Fibonacci(16): 987
printf("Time taken: %v seconds\n", clock() - start);  // Outputs: Time taken: <seconds>
```

**Explanation**:  

- The `fib` function uses recursion: `fib(n) = fib(n-2) + fib(n-1)`, with a base case of `n < 2`.
- `clock()` returns the current time in seconds, used here to measure performance.
- `printf` provides formatted output, showing the time difference.

---

## 4. Fibonacci Iterative (`fibonacci_iterative.spy`)

An iterative approach to Fibonacci calculation, optimized for performance.

```spy
fn fib(n) {
    if (n < 2) return n;
    var a = 0;
    var b = 1;
    for (var i = 2; i <= n; i = i + 1) {
        var temp = a + b;
        a = b;
        b = temp;
    }
    return b;
}

var start = clock();
println("Fibonacci(16):", fib(16));  // Outputs: Fibonacci(16): 987
printf("Time taken: %v seconds\n", clock() - start);  // Outputs: Time taken: <seconds>
```

**Explanation**:  

- This version uses a `for` loop to iteratively compute Fibonacci numbers, updating `a` and `b` in each step.
- It’s more efficient than recursion for large `n`, as shown by the shorter execution time.
- The syntax `i = i + 1` is used for incrementing, though SPYScript also supports `i++`.

---

## 5. Structs and Control Flow (`cat_example.spy`)

This script combines structs, functions, conditionals, and loops, with ASCII and Unicode examples.

```spy
// ASCII Struct
struct Animal {
    species = "Unknown";
    length = 50;  // Average cat length in cm
    height = 25;  // Average cat height in cm
}

fn describeAnimal(animal) {
    return animal.species + ": " + to_str(animal.length) + "x" + to_str(animal.height) + " cm";
}

var favorite = Animal();
favorite.species = "Cat";
if (favorite.length <= 50) {
    println("Animal is average or shorter.");
} else {
    println("Animal is longer than average.");
}
println("Description:", describeAnimal(favorite));  // Outputs: Description: Cat: 50x25 cm

var count = 0;
while (count < 2) {
    println("Meow #", to_str(count));  // Outputs: Meow #0, Meow #1
    count = count + 1;
}

// Unicode Struct
struct 猫 {
    種類 = "不明";
    年 = 2;
}
var 子猫 = 猫();
子猫.種類 = "たまちゃん";
println("Cat name (子猫.種類):", 子猫.種類);  // Outputs: Cat name (子猫.種類): たまちゃん
```

**Explanation**:  

- Structs are defined with `struct`, allowing fields with default values (e.g., `species = "Unknown"`).
- The `if` statement checks conditions, and `while` loops iterate based on a condition.
- Unicode structs (e.g., `猫`) show SPYScript’s multilingual support, with fields accessed via dot notation.

---

## 6. Arrays (`arrays.spy`)

SPYScript provides robust array support with built-in functions for manipulation.

```spy
var arr = [1, 2, 3, 4, 5];
println("Original array:", array_to_string(arr));  // Outputs: Original array: [1, 2, 3, 4, 5]
push(arr, 6);
println("After push(6):", array_to_string(arr));  // Outputs: After push(6): [1, 2, 3, 4, 5, 6]
println("Popped:", pop(arr));  // Outputs: Popped: 6

for (var i = 0; i < len(arr); i = i + 1) {
    println("Element", i, ":", arr[i]);  // Outputs: Element 0: 1, Element 1: 2, etc.
}
```

**Explanation**:  

- Arrays are created with square brackets (e.g., `[1, 2, 3]`).
- Functions like `push`, `pop`, `len`, and `array_to_string` manipulate and inspect arrays.
- The `for` loop iterates over indices, accessing elements with `arr[i]`.

---

## 7. File Operations (`file.spy`)

SPYScript supports basic file I/O operations for reading and writing text files.

```spy
var filename = "test.txt";
var content = "Hello from SPYScript!\nWritten on April 09, 2025.";
write_file(filename, content);
println("Wrote to file:", filename);

var readContent = read_file(filename);
println("Read from file:", readContent);  // Outputs: Read from file: Hello from SPYScript!...
```

**Explanation**:  

- `write_file(filename, content)` writes a string to a file, overwriting if it exists.
- `read_file(filename)` reads the entire file content as a string.
- Useful for simple file-based persistence or logging.

---

## 8. Native Functions (`native_functions.spy`)

SPYScript includes native functions for tasks like timing, randomization, and I/O.

```spy
var time = clock();
println("Current time:", time);  // Outputs: Current time: <timestamp>

var numbers = [1, 2, 3, 4, 5];
shuffle(numbers);
println("Shuffled array:", array_to_string(numbers));  // Outputs: Shuffled array: [e.g., 3, 1, 5, 2, 4]

var randNum = random_between(1, 10);
println("Random number (1-10):", randNum);  // Outputs: Random number (1-10): <random>
```

**Explanation**:  

- `clock()` returns the current time in seconds, useful for timing code.
- `shuffle(array)` randomizes array elements in place.
- `random_between(min, max)` generates a random integer between `min` and `max` (inclusive).

---

## 9. Modules (`modules.spy`)

SPYScript supports modular programming with `import` and `mod` for code organization.

```spy
// math.spy
fn add(a, b) {
    return a + b;
}
var PI = 3.14159;

// main.spy
import "math.spy";
println("Imported add(2, 3):", add(2, 3));  // Outputs: Imported add(2, 3): 5
println("Imported PI:", PI);  // Outputs: Imported PI: 3.14159
```

**Explanation**:  

- Use `import "filename.spy"` to include external scripts.
- Functions and variables from the imported module (e.g., `add`, `PI`) are directly accessible.
- Modules help organize code into reusable units.

---

## 10. Additional Features

### 10.1. Input Handling (`input.spy`)

```spy
println("Enter a sentence:");
var input = scanln();
println("You entered:", input);  // Outputs: You entered: <user input>
```

**Explanation**:  

- `scanln()` reads a line of user input from the console, returning it as a string.

### 10.2. String Formatting (`format.spy`)

```spy
var name = "Alice";
var age = 25;
var formatted = sprintf("Name: %v, Age: %v", name, age);
println("Formatted:", formatted);  // Outputs: Formatted: Name: Alice, Age: 25
```

**Explanation**:  

- `sprintf(format, ...args)` formats strings with placeholders (e.g., `%v` for values).

### 10.3. Control Flow Enhancements (`control.spy`)

```spy
var x = 5;
if (x > 0) {
    println("Positive");
} else if (x < 0) {
    println("Negative");
} else {
    println("Zero");
}

for (var i = 0; i < 3; i = i + 1) {
    if (i == 1) continue;  // Skip 1
    println("i:", i);  // Outputs: i: 0, i: 2
}
```

**Explanation**:  

- Supports `if`, `else if`, `else` for conditionals, and `continue`/`break` in loops.

### 10.4. Operators (`operators.spy`)

```spy
var a = 10;
var b = 3;
println("Add:", a + b);  // Outputs: Add: 13
println("Exponent:", a ** 2);  // Outputs: Exponent: 100
println("Integer Div:", a /_ b);  // Outputs: Integer Div: 3
println("25 percent of 1000:", 25 %% 1000);  // Outputs: 250
```

**Explanation**:  

- Includes arithmetic (`+`, `-`, `*`, `/`), exponentiation (`**`), integer division (`/_`), and percentage (%%).
