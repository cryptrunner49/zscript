# 🕵️‍♂️ SPYScript

**SPYScript** is a lightweight.

---

## 👩‍💻 Hello World

```spy
hello = "Hello, World!";
println(hello)  // Outputs: Hello, World!
```

📖 Explore variables, structs, loops, and more in the [SPYScript Usage Guide](SPYSCRIPT_USAGE.md).

---

## ✨ Features

- **🌍 Unicode & Emoji Identifiers** — Use `π` or even `🐱` as variable names.
- **🧠 Simple Syntax** — Easy-to-learn keywords like `var`, `func`, `if`, and `for`.
- **⚙️ Native Functions** — Built-ins such as `clock()`, `shuffle()`, and `random_between()`.
- **🧱 Structs & Closures** — Create custom types and use powerful functional constructs.
- **📁 File I/O** — Read and write files with `read_file()` and `write_file()`.
- **🖥 Cross-Platform** — Works on any system with the required dependencies.

📚 Dive deeper in the [SPYScript Usage Guide →](SPYSCRIPT_USAGE.md)

---

## 📦 Installation

### ✅ Requirements

- Linux (Debian, Ubuntu, Fedora, Arch, etc.) or macOS
- [Go (Golang)](https://golang.org)
- Dependencies: `gcc`, `pkg-config`, `make`, `libffi`, `readline`

### 🧰 System Setup

#### Option 1: Install via Script

**System-wide installation (requires `sudo`)**:

```bash
curl -sL https://github.com/cryptrunner49/spy/raw/refs/heads/main/install.sh | bash -s -- install --system
```

**User-only installation (`$HOME/.local/bin`)**:

```bash
curl -sL https://github.com/cryptrunner49/spy/raw/refs/heads/main/install.sh | bash -s -- install --user
```

#### Option 2: Manual Download

```bash
curl -LO https://github.com/cryptrunner49/spy/releases/latest/download/spyvm
chmod +x spyvm
```

---

### 🛠 Build From Source

1. Clone the repository:

   ```bash
   git clone https://github.com/cryptrunner49/spy.git
   cd spy
   ```

2. Build the interpreter:

   ```bash
   make
   ```

3. Run a script:

   ```bash
   ./bin/spyvm sample/game.spy
   ```

---

## 🧪 Platform-Specific Setup

### Ubuntu/Debian

```bash
sudo apt update
sudo apt install gcc pkg-config make golang libffi-dev libreadline-dev
```

### macOS

```bash
brew install go pkg-config gcc make libffi readline
```

---

## 🗺 Roadmap

Coming soon to SPYScript:

- [ ] **Pattern Matching** — More expressive conditionals.
- [ ] **Switch Statement** — Cleaner multi-branch logic.
- [ ] **Elif Support** — Less nesting, more clarity.
- [ ] **Enums** — Organize data like a pro.
- [ ] **Error Handling** — Try-catch or similar constructs.
- [ ] **Standard Library** — More built-in power.

🎯 Track progress or suggest features via [Issues →](https://github.com/cryptrunner49/spy/issues)

---

## 🤝 Contributing

We’d love your help! Whether it's fixing bugs, improving docs, or proposing features—your contributions matter.

📘 See the [Contributing Guide →](CONTRIBUTING.md) to get started.

---

## 📄 License

SPYScript is licensed under the **GNU General Public License v3.0 (GPL-3.0)**.  
See the [LICENSE](LICENSE) file for full details.
