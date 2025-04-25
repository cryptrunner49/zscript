# Contributing to the ZScript Project

Thank you for your interest in contributing to the ZScript Project! We welcome contributions from the community to enhance the language, its documentation, and examples. This guide outlines how you can get involved, from reporting issues to submitting code changes.

---

## Table of Contents

1. [How to Contribute](#how-to-contribute)
2. [Reporting Issues](#reporting-issues)
3. [Suggesting Enhancements](#suggesting-enhancements)
4. [Submitting Code Changes](#submitting-code-changes)
   - [Getting Started](#getting-started)
   - [Code Style Guidelines](#code-style-guidelines)
   - [Submitting a Pull Request](#submitting-a-pull-request)
5. [Community Guidelines](#community-guidelines)
6. [Contact](#contact)

---

## How to Contribute

There are several ways to contribute to the ZScript project:

- **Report Bugs**: Identify and report issues with the language or examples.
- **Suggest Features**: Propose new language features, native functions, or example scripts.
- **Improve Documentation**: Enhance the `ZSCRIPT_USAGE.MD`, this file, or add new guides.
- **Submit Code**: Add new `.z` example scripts, fix bugs, or improve the interpreter (if applicable).

All contributions should align with the project’s goal of making ZScript a simple, expressive, and Unicode-friendly scripting language.

---

## Reporting Issues

If you encounter a bug or unexpected behavior:

1. **Check Existing Issues**: Search the [Issues](https://github.com/cryptrunner49/zscript/issues) tab to avoid duplicates.
2. **Open a New Issue**:
   - Provide a clear title (e.g., "Runtime Error in `for` Loop with Unicode Variables").
   - Include:
     - A description of the issue.
     - Steps to reproduce (e.g., a minimal `.z` script).
     - Expected vs. actual behavior.
     - ZScript version (if known) and environment (e.g., OS).
   - Example:

     ```text
     **Title**: Runtime Error: String Comparison in Loop
     **Description**: Using `>` with strings causes an error.
     **Steps**: Run `var a = "1"; println(a > "2");`
     **Expected**: False
     **Actual**: Runtime Error: Both operands must be numbers
     **Environment**: ZScript vX.X, Windows 10
     ```

3. **Label**: Add the `bug` label to help us triage.

---

## Suggesting Enhancements

Have an idea for a new feature or improvement?

1. **Check Existing Suggestions**: Look in [Issues](https://github.com/cryptrunner49/zscript/issues) for similar proposals.
2. **Open a New Issue**:
   - Use the "Feature Request" template (if available).
   - Provide:
     - A clear title (e.g., "Add Array Slicing Syntax").
     - A detailed description of the feature.
     - Use case or example code.
     - Potential benefits (e.g., improved readability).
   - Example:

     ```text
     **Title**: Add Array Slicing Syntax
     **Description**: Support `arr[start:end]` for slicing arrays.
     **Example**: `var arr = [1, 2, 3, 4]; println(arr[1:3]); // Outputs: [2, 3]`
     **Benefits**: Simplifies array manipulation, aligns with other languages.
     ```

3. **Label**: Add the `enhancement` label.

---

## Submitting Code Changes

### Getting Started

1. **Fork the Repository**: Click "Fork" on [GitHub](https://github.com/cryptrunner49/zscript) to create your copy.
2. **Clone Your Fork**: `git clone https://github.com/<your-username>/<repo-name>.git`
3. **Create a Branch**: `git checkout -b <branch-name>` (e.g., `git checkout -b add-array-slicing`)
4. **Make Changes**: Edit files locally (e.g., add `.z` scripts or modify the ZScript VM).
5. **Test**: Run your changes with the ZScript VM to ensure they work as expected.

### Code Style Guidelines

For `.z` scripts:

- **Section Headers**: Use `// --- Section Name ---` and `println("--- Section Name ---");` for organization (see `ZSCRIPT_USAGE.MD`).
- **Print Statements**: Include descriptive labels (e.g., `println("Result:", value);`).
- **Spacing**: Use consistent indentation (2 or 4 spaces) and line breaks for readability.
- **Comments**: Add brief comments for complex logic (e.g., `// Base case for recursion`).
- **Naming**: Use meaningful names; Unicode/emoji names are encouraged where thematic (e.g., `🐱` for cat-related examples).

For Compiler and VM code:

- Follow the Go style conventions (e.g., use gofmt for formatting).
- Document new functions with comments.
- Ensure that comments are clear and provide context for the functionality.

### Submitting a Pull Request

1. **Commit Changes**: `git add . && git commit -m "Add array slicing example in arrays.z"`
   - Use clear, concise commit messages (e.g., "Fix bug in string comparison", "Update README with modules").
2. **Push to Your Fork**: `git push origin <branch-name>`
3. **Open a Pull Request**:
   - Go to the original repository and click "New Pull Request".
   - Select your branch and compare it to the main branch.
   - Provide:
     - A title (e.g., "Add Array Slicing Example").
     - A description:
       - What you changed.
       - Why (e.g., fixes issue #123).
       - How to test (e.g., "Run `arrays.z`").
   - Example:

     ```text
     **Title**: Add Array Slicing Example
     **Description**: Added slicing demo to `arrays.z`. Resolves #45.
     **Changes**: Updated `arrays.z` with `arr[1:3]` example.
     **Test**: Run `seed arrays.z` and check output: `[2, 3]`.
     ```

4. **Label**: Add `enhancement`, `bug`, or `documentation` as appropriate.
5. **Review**: Respond to feedback from maintainers.

---

## Community Guidelines

- **Be Respectful**: Treat everyone with kindness and respect, regardless of experience level.
- **Be Constructive**: Offer helpful feedback and suggestions.
- **Stay On Topic**: Keep discussions and contributions relevant to ZScript.
- **Follow Standards**: Adhere to this guide and the project’s goals.

We aim to foster an inclusive, collaborative community. Violations may result in moderation by maintainers.

---

## Contact

- **Issues**: Use GitHub Issues for bugs or feature requests.
- **Discussions**: Join the [Discussions](https://github.com/cryptrunner49/zscript/discussions) tab for general questions or ideas.
- **Email**: Reach out to `<your-email>` (optional, replace with your contact if desired).

---

## Acknowledgments

Thank you for contributing to ZScript! Your efforts help make this language more powerful, accessible, and fun to use. Every bug report, feature idea, or code submission brings us closer to a better ZScript ecosystem.
