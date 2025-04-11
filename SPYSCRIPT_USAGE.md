# SPYScript Usage Guide

Welcome to the SPYScript Usage Guide! SPYScript is a versatile scripting language that supports modern programming constructs, a variety of built-in features, and Unicode/emoji identifiers. This guide walks through the language features with practical examples and explanations.

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

```spy
// Number
var num = 42;
println("Number:", num);

// String
var message = "Hello, World!";
println("Message:", message);

// Boolean
var isTrue = true;
println("Boolean:", isTrue);

// Null
var nothing = null;
println("Null:", nothing);
```

---

## 2. Closures

```spy
fn outer() {
    var a = 1;
    var b = 2;
    fn middle() {
        var c = 3;
        var d = 4;
        fn inner() {
            println("Sum:", a + c + b + d);
        }
        inner();
    }
    middle();
}
outer();
```

---

## 3. Fibonacci Recursive

```spy
fn fib(n) {
    if (n < 2) return n;
    return fib(n - 2) + fib(n - 1);
}

var start = clock();
println("Fibonacci(16):", fib(16));
printf("Time taken: %v seconds\n", clock() - start);
```

---

## 4. Fibonacci Iterative

```spy
fn fib(n) {
    if (n < 2) return n;
    var a = 0;
    var b = 1;
    for (var i = 2; i <= n; i++) {
        var temp = a + b;
        a = b;
        b = temp;
    }
    return b;
}

var start = clock();
println("Fibonacci(16):", fib(16));
printf("Time taken: %v seconds\n", clock() - start);
```

---

## 5. Structs and Control Flow

```spy
struct Animal {
    species = "Unknown";
    length = 50;
    height = 25;
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
println("Description:", describeAnimal(favorite));

var count = 0;
while (count < 2) {
    println("Meow #", to_str(count));
    count = count + 1;
}
```

---

## 6. Arrays

```spy
var arr = [1, 2, 3, 4, 5];
println("Original array:", array_to_string(arr));
push(arr, 6);
println("After push(6):", array_to_string(arr));
println("Popped:", pop(arr));

for (var i = 0; i < len(arr); i++) {
    println("Element", i, ":", arr[i]);
}
```

---

## 7. File Operations

```spy
var filename = "test.txt";
var content = "This is a file handling example.\nDemonstrating how to read and write files.";
write_file(filename, content);
println("Wrote to file:", filename);

var readContent = read_file(filename);
println("Read from file:", readContent);
```

---

## 8. Native Functions

```spy
var time = clock();
println("Current time:", time);

var numbers = [1, 2, 3, 4, 5];
shuffle(numbers);
println("Shuffled array:", array_to_string(numbers));

var randNum = random_between(1, 10);
println("Random number (1-10):", randNum);
```

---

## 9. Modules

```spy
// geometry.spy

mod Geometry {
    mod Shapes {
        fn area_circle(radius) {
            return Geometry.PI * radius * radius;
        }

        fn perimeter_circle(radius) {
            return 2 * Geometry.PI * radius;
        }
    }

    var PI = 3.14159;
}
```

---

## 10. Import

```spy
// main.spy
import "geometry.spy";

println("Circle Area:", Geometry.Shapes.area_circle(5));
println("Circle Perimeter:", Geometry.Shapes.perimeter_circle(5));
```

---

## 11. Additional Features

### 11.1. Input Handling

```spy
println("Enter a sentence:");
var input = scanln();
println("You entered:", input);
```

### 11.2. String Formatting

```spy
var name = "Alice";
var age = 25;
var formatted = sprintf("Name: %v, Age: %v", name, age);
println("Formatted:", formatted);
```

### 11.3. Advanced Control Flow

```spy
var x = 5;
if (x > 0) {
    println("Positive");
} else if (x < 0) {
    println("Negative");
} else {
    println("Zero");
}

for (var i = 0; i < 3; i++) {
    if (i == 1) continue;
    println("i:", i);
}
```

### 11.4. Operators

```spy
var a = 10;
var b = 3;
println("Add:", a + b);
println("Exponent:", a ** 2);
println("Integer Div:", a /_ b);
println("25 percent of 1000:", 25 %% 1000);
```

---

## 12. Unicode Support

SPYScript supports Unicode and emoji characters in identifiers, strings, and output. This can be useful for internationalization, math expressions, or fun scripting.

```spy
// Unicode variable
var π = 3.14159;
println("Value of π:", π);

// Emoji variable
var 🔥 = "Fire emoji";
println("Emoji variable 🔥:", 🔥);

// Unicode string
var greeting = "こんにちは世界"; // Japanese: Hello, World
println("Greeting in Japanese:", greeting);

// Struct with Unicode fields
struct 数学 {
    半径 = 5;
}
var 円 = 数学();
println("半径 (radius):", 円.半径);
```
