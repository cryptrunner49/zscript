# 💤📜 ZScript

**ZScript** is a lightweight, expressive scripting language inspired by Python — but with its own minimalist and expressive style. It supports Unicode and emoji identifiers, functional programming constructs, file I/O, and can be embedded as a shared C library. Whether you're automating tasks, scripting games, or exploring creative code, ZScript makes it simple and fun.

---

## 🚀 Quick Start

```z
var hello = "Hello, World!";
println(hello);  // Outputs: Hello, World!
```

📖 Explore language features in the [Usage Guide →](ZSCRIPT_USAGE.md)

---

## ✨ Features

- 🌍 **Unicode & Emoji Identifiers** — Name variables `π`, `🐱`, or any emoji you like.
- 🧠 **Familiar Python-like Syntax** — Use `var`, `const`, `func`, `if`, `for`, and more.
- ⚙️ **Built‑in Functions** — `clock()`, `shuffle()`, `random_between()` and others built-in.
- 📁 **File I/O** — `read_file()` and `write_file()` to handle files natively.
- 🧱 **Structs & Closures** — Define custom types and encapsulate behavior.
- 🐧 **Linux‑Only** — Runs out of the box on most Linux distributions.
- 🧬 **Embeddable VM** — Integrate ZScript with C, Go, Rust, or C++ applications.

---

## 🔽 Download

Get the latest prebuilt binaries and development files from the [Releases →](https://github.com/cryptrunner49/zscript/releases/latest):

- **💤 VM Executable**: [Download `zvm`](https://github.com/cryptrunner49/zscript/releases/latest/download/zvm)
- **🔧 Development Files**:
  - [libzscript.h](https://github.com/cryptrunner49/zscript/releases/latest/download/libzscript.h)
  - [libzscript.so](https://github.com/cryptrunner49/zscript/releases/latest/download/libzscript.so)
- **📦 Full Release Bundle** (VM + Libs + Headers):  
  [zscript-release.zip](https://github.com/cryptrunner49/zscript/releases/latest/download/zscript-release.zip)

---

## ⚙️ Installation

### ✅ Requirements

- OS: Linux (Debian, Ubuntu, Fedora, Arch, etc.)
- Tools: `gcc`, `make`, `pkg-config`, `libffi`, `readline`
- [Go (Golang)](https://golang.org)

### 🧰 Install with Script

**System-wide:**

```bash
curl -sL https://github.com/cryptrunner49/zscript/raw/refs/heads/main/install.sh | bash -s -- install --system
```

**User-only:**

```bash
curl -sL https://github.com/cryptrunner49/zscript/raw/refs/heads/main/install.sh | bash -s -- install --user
```

### 🏗 Build from Source

```bash
git clone https://github.com/cryptrunner49/zscript.git
cd zscript
make vm
./bin/zvm samples/scripts/rpg_game.z
```

---

## 🧠 Embedding ZScript

ZScript is easy to embed in other languages like **C, Go, C++**, and **Rust**.

### ✅ Example in C

```c
#include "libzscript.h"
#include <stdio.h>
#include <stdlib.h>

int main(int argc, char** argv) {
    ZScript_Init(argc, argv);

    if (argc > 1) {
        ZScript_RunFile(argv[1]);
    } else {
        int exitCode;
        char* result = ZScript_InterpretWithResult("1 + 2;", "<script>", &exitCode);
        if (exitCode == 0) printf("Last value: %s\n", result);
        else printf("Execution failed with code %d\n", exitCode);
        free(result);
    }

    ZScript_Free();
    return 0;
}
```

### 🛠 Build & Run

```bash
make lib
gcc -o run_sample samples/lib/c/sample.c -Lbin -lzscript -Ibin
LD_LIBRARY_PATH=bin ./run_sample
```

---

## 🔍 More Embedding Examples

Find ready-to-run embedding samples in:

- [📄 C](samples/lib/c/sample.c)
- [📄 C++](samples/lib/c/sample.cpp)
- [📄 Go](samples/lib/c/sample.go)
- [📄 Rust](samples/lib/c/sample.rust)

These show how to use ZScript with FFI across different ecosystems.

---

## 🧪 Platform‑Specific Setup

### Ubuntu / Debian

```bash
sudo apt update
sudo apt install gcc pkg-config make golang libffi-dev libreadline-dev
```

---

## 🗺 Roadmap

What’s next for ZScript?

- [x] **Elif**
- [ ] **Pattern Matching**
- [ ] **Switch Statement**
- [ ] **Enums**
- [ ] **Error Handling**
- [ ] **Standard Library**

✨ Track progress or suggest features via [Issues →](https://github.com/cryptrunner49/zscript/issues)

---

## 🤝 Contributing

We’d love your help! Whether it's fixing bugs, improving docs, or proposing features—your contributions matter.

👉 See the [Contributing Guide →](CONTRIBUTING.md) to get started.

---

## 📄 License

**ZScript** is licensed under the **GNU GPL-3.0**.  
See the [LICENSE](LICENSE) for details.
