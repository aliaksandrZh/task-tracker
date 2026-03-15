---
name: code-standards
description: Coding principles to follow when writing or modifying code — DRY, KISS, semantic naming, SRP, minimal changes
---

## Coding Principles

### 1. DRY (Don't Repeat Yourself)
- Abstract repetitive logic into reusable functions or modules.
- If a pattern or logic is used more than once, refactor it into a shared utility.
- Centralize constants and configuration.

### 2. KISS (Keep It Simple, Stupid)
- Prioritize readability over "clever" or overly-optimized code.
- Use standard language features and avoid unnecessary dependencies.
- Code should be self-explanatory to another developer.

### 3. Semantic Naming
- Use clear, descriptive names for variables, functions, and modules.

### 4. Single Responsibility Principle (SRP)
- Each function, class, or module must do **one thing** well.

### 5. Do Not Touch Existing Code
- Do not refactor or modify existing code beyond what the task requires.
- If you spot a bug in existing code, report it to the user and let them decide.
- Only change existing code if strictly necessary to achieve the task's goal.
