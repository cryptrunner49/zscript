# ZScript Usage Guide

Welcome to the ZScript Usage Guide! ZScript is a versatile scripting language that supports modern programming constructs, a variety of built-in features, and Unicode/emoji identifiers. This guide walks through the language features with practical examples and explanations.

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
10. [Import](#10-import)  
11. [Additional Features](#11-additional-features)  
12. [Unicode Support](#12-unicode-support)

---

## 1. Variables

```z
// Implicitly declaring a Number
num = 42
println("Number:", num)

// Explicitly declaring a Number
var num = 42
println("Number:", num)

// Implicitly declaring a String
message = "Hello, World!"
println("Message:", message)

// Explicitly declaring a String
var message = "Hello, World!"
println("Message:", message)

// Implicitly declaring a Boolean
isTrue = true
println("Boolean:", isTrue)

// Explicitly declaring a Boolean
var isTrue = true
println("Boolean:", isTrue)

// Implicitly declaring a Null
nothing = null
println("Null:", nothing)

// Explicitly declaring a Null
var nothing = null
println("Null:", nothing)
```

---

## 2. Closures

```z
func outer():
    var a = 1
    var b = 2
    func middle():
        var c = 3
        var d = 4
        func inner():
            println("Sum:", a + c + b + d)
        inner()
    middle()
outer()
```

---

## 3. Fibonacci Recursive

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

## 4. Fibonacci Iterative

```z
func fib(n):
    if (n < 2):
        return n
    var a = 0
    var b = 1
    for (var i = 2; i <= n; i++):
        var temp = a + b
        a = b
        b = temp
    return b

var start = clock()
println("Fibonacci(16):", fib(16))
printf("Time taken: %v seconds\n", clock() - start)
```

---

## 5. Structs and Control Flow

```z
struct Animal:
    species = "Unknown"
    length = 50
    height = 25

func describeAnimal(animal):
    return animal.species + ": " + to_str(animal.length) + "x" + to_str(animal.height) + " cm"

var favorite = Animal()
favorite.species = "Cat"
if (favorite.length <= 50):
    println("Animal is average or shorter.")
else:
    println("Animal is longer than average.")

println("Description:", describeAnimal(favorite))

var count = 0
while (count < 2):
    println("Meow #", to_str(count))
    count = count + 1
```

---

## 6. Arrays

```z
var arr = [1, 2, 3, 4, 5]
println("Original array:", array_to_string(arr))
push(arr, 6)
println("After push(6):", array_to_string(arr))
println("Popped:", pop(arr))

for (var i = 0; i < len(arr); i++):
    println("Element", i, ":", arr[i])
```

---

## 7. Maps

```z
// Basic Map Operations
println("--- Map Demo ---")
var map = { "name": "Alice", "age": 30 }
println("Name:", map["name"])  // Outputs: Alice
println("Age:", map["age"])    // Outputs: 30
map["age"] = 31
println("Updated age:", map["age"])  // Outputs: 31

// Map Functions
println("--- Map Functions ---")
var m = {"a": 1, "b": 2}
println("Initial map:", m)
map_remove(m, "a")
println("After removing 'a':", m)
println("Contains key 'b':", map_contains_key(m, "b"))
println("Contains value 2:", map_contains_value(m, 2))
println("Map size:", map_size(m))
println("Keys:", map_keys(m))
println("Values:", map_values(m))
map_clear(m)
println("After clear:", m)

// Map Addition
println("--- Map Addition ---")
var a = {"x": 1, "y": 2}
var b = {"y": 3, "z": 4}
println("a + b:", a + b)

// Map Subtraction
println("--- Map Subtraction ---")
var a = {"x": 1, "y": 2, "z": 3}
var b = {"y": null, "w": null}
println("a - b:", a - b)
```

---

## 8. File Operations

```z
var filename = "test.txt"
var content = "This is a file handling example.\nDemonstrating how to read and write files."
write_file(filename, content)
println("Wrote to file:", filename)

var readContent = read_file(filename)
println("Read from file:", readContent)
```

---

## 9. Native Functions

```z
var time = clock()
println("Current time:", time)

var numbers = [1, 2, 3, 4, 5]
shuffle(numbers)
println("Shuffled array:", array_to_string(numbers))

var randNum = random_between(1, 10)
println("Random number (1-10):", randNum)
```

---

## 10. Modules

```z
// geometry.z

mod Geometry:
    mod Shapes:
        func area_circle(radius):
            return Geometry.PI * radius * radius

        func perimeter_circle(radius):
            return 2 * Geometry.PI * radius

    var PI = 3.14159
```

---

## 11. Import

```z
// main.z
import "geometry.z"

println("Circle Area:", Geometry.Shapes.area_circle(5))
println("Circle Perimeter:", Geometry.Shapes.perimeter_circle(5))
```

---

## 12. Additional Features

### 12.1. Input Handling

```z
println("Enter a sentence:")
var input = scanln()
println("You entered:", input)
```

### 12.2. String Formatting

```z
var name = "Alice"
var age = 25
var formatted = sprintf("Name: %v, Age: %v", name, age)
println("Formatted:", formatted)
```

### 12.3. Advanced Control Flow

```z
var x = 5
if (x > 0):
    println("Positive")
else:
    if (x < 0):
        println("Negative")
    else:
        println("Zero")

for (i = 0; i < 3; i++):
    if (i == 1):
        continue // Skip the iteration when i is 1
    if (i == 2):
        break    // Exit the loop when i is 2
    println("i:", i)
```

### 12.4. Operators

```z
var a = 10
var b = 3
println("Add:", a + b)
println("Exponent:", a ** 2)
println("Integer Div:", a /_ b)
println("25 percent of 1000:", 25 %% 1000)
```

---

## 13. Unicode Support

ZScript supports Unicode and emoji characters in identifiers, strings, and output. This can be useful for internationalization, math expressions, or fun scripting.

```z
// Unicode variable
var π = 3.14159
println("Value of π:", π)

// Emoji variable
var 🔥 = "Fire emoji"
println("Emoji variable 🔥:", 🔥)

// Unicode string
var greeting = "こんにちは世界" // Japanese: Hello, World
println("Greeting in Japanese:", greeting)

// Struct with Unicode fields
struct 数学:
    半径 = 5

var 円 = 数学()
println("半径 (radius):", 円.半径)
```
