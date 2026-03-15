---
name: dev
description: Develop a feature or fix from a task file, prompt, or description — analyze, implement, test, commit
---

Development workflow for implementing tasks. Input can be a task description, a prompt, or a path to a task file (e.g. a TASKS.md with TODO items).

## Workflow

### 1. Understand the task
- Read the task description or file provided
- Read relevant source files to understand the current code
- If anything is unclear or ambiguous, ask the user before proceeding

### 2. Implement
- Make the required changes
- Keep changes minimal and focused — only what the task requires
- Follow existing code style and patterns

### 3. Run tests
- Run `go test ./...` to check all tests pass
- If tests break:
  - List the broken test names to the user
  - Wait for the user to confirm which tests need updating
  - Only then update the tests as confirmed
- If no tests break, proceed

### 4. Manual test
- Ask the user to manually test the feature
- Wait for their feedback
- If they request adjustments, go back to step 2

### 5. Write tests
- Add unit tests covering the new or changed functionality
- Follow existing test patterns in the codebase
- Run `go test ./...` to check all tests pass

### 6. Commit
- Use the /commit skill to create a commit

### 7. Mark complete
- If the task came from a file (e.g. TASKS.md), update the task status to DONE
