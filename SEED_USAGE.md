# Seed Language Usage Guide

This guide showcases the core features of the Seed language through examples. Each script demonstrates a specific concept, such as variables, functions, closures, structs, and control flow, with explanations for clarity.

---

## 1. Variables (`variables.seed`)

This script shows how to declare and use variables of different types, including Unicode and emoji names.

```seed
// Number
var num = 42;
print(num);  // Outputs: 42

// String
var message = "Hello, Seed!";
print(message);  // Outputs: Hello, Seed!

// Boolean
var isTrue = true;
print(isTrue);  // Outputs: true

// Null
var nothing = null;
print(nothing);  // Outputs: null

// Unicode variable names
var π = 3.14;
print(π);  // Outputs: 3.14

var 挨拶 = "こんにちは";
print(挨拶);  // Outputs: こんにちは

// Emoji variable names
var 🔢 = 100;
print(🔢);  // Outputs: 100

var 💬 = "Emoji!";
print(💬);  // Outputs: Emoji!
```

**Explanation**:  

- Use `var` to declare variables and assign them values like numbers, strings, booleans, or `null`.  
- Seed supports Unicode characters (e.g., `π`, `挨拶`) and emojis (e.g., `🔢`, `💬`) as variable names.

---

## 2. Closures (`closures.seed`)

This script demonstrates nested functions and closures, where inner functions access variables from outer scopes.

```seed
fn outer() {
    var a = 1;
    var b = 2;
    fn middle() {
        var c = 3;
        var d = 4;
        fn inner() {
            print a + c + b + d;  // Accesses variables from outer scopes
        }
        inner();
    }
    middle();
}
outer();  // Outputs: 10 (1 + 3 + 2 + 4)
```

**Explanation**:  

- The `inner` function accesses `a` and `b` from `outer` and `c` and `d` from `middle`, showcasing closure behavior.  
- Calling `outer()` triggers the nested calls, resulting in `10`.

---

## 3. Fibonacci Recursive (`fibonacci_recursive.seed`)

This script calculates the Fibonacci sequence using recursion and measures execution time.

```seed
fn fib(n) {
    if (n < 2) return n;
    return fib(n - 2) + fib(n - 1);
}

var start = clock();
print fib(16);  // Outputs: 987
print clock() - start;  // Outputs: time taken
```

**Explanation**:  

- The `fib` function recursively computes the nth Fibonacci number.  
- `clock()` measures the time before and after execution, showing the performance cost of recursion.

---

## 4. Fibonacci Iterative (`fibonacci_iterative.seed`)

This script calculates the Fibonacci sequence iteratively using a loop, also with timing.

```seed
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
print fib(16);  // Outputs: 987
print clock() - start;  // Outputs: time taken (faster than recursion)
```

**Explanation**:  

- This iterative approach uses a `for` loop to compute Fibonacci numbers, avoiding recursive overhead.  
- It’s faster than the recursive version for larger values of `n`.

---

## 5. Structs and Control Flow (`cat_example.seed`)

This script demonstrates structs, functions, conditionals, and loops with a cat theme, using both ASCII and Unicode identifiers.

```seed
// ASCII Example
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
    print("Animal is average or shorter.");
} else {
    print("Animal is longer than average.");
}

var description = describeAnimal(favorite);
print(description);  // Outputs: Cat: 50x25 cm

var count = 0;
while (count < 2) {
    print("meow #" + to_str(count) + " of the cat");
    count = count + 1;
}

// Unicode Example
struct 動物 {
    種類 = "不明";
    長さ = 50;  // Average cat length in cm
    高さ = 25;  // Average cat height in cm
}

fn 動物を説明する(アニマル) {
    return アニマル.種類 + ": " + to_str(アニマル.長さ) + "x" + to_str(アニマル.高さ) + " cm";
}

var お気に入り = 動物();
お気に入り.種類 = "ネコ";

if (お気に入り.長さ <= 50) {
    print("動物が平均または短い。🐾🐱");
} else {
    print("動物が平均より長い。🐾");
}

var 説明 = 動物を説明する(お気に入り);
print(説明);  // Outputs: ネコ: 50x25 cm

var 回数 = 0;
while (回数 < 2) {
    print("にゃん #" + to_str(回数) + "🐾");
    回数 = 回数 + 1;
}
```

**Explanation**:  

- **Structs**: Define `Animal` (or `動物`) with fields for species, length, and height, initialized with cat-like defaults (50 cm length, 25 cm height).  
- **Functions**: `describeAnimal` (or `動物を説明する`) returns a string describing the cat’s species and dimensions.  
- **Conditionals**: Check if the cat’s length is average or shorter (`<= 50`) vs. longer than average.  
- **Loops**: Use a `while` loop to print "meow" (ASCII) or "にゃん" (Unicode) with iteration numbers, enhanced with cat-themed output and emojis (`🐾`, `🐱`).  
- **Unicode Support**: The language handles ASCII (`Cat`), Unicode (`ネコ`), and emoji (`🐾`) seamlessly in variable names, fields, and strings.
