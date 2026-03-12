---
name: release
description: Create a new release with git tag and GitHub release
disable-model-invocation: true
allowed-tools: Bash(git *, go *, gh *)
argument-hint: "[version, e.g. v1.2.0]"
---

Release version $ARGUMENTS:

1. Ensure working tree is clean (`git status`). If not, abort and ask user to commit first.
2. Run tests: `go test ./...`
3. If tests fail, abort.
4. Build to verify: `go build -o /dev/null .`
5. Create an annotated git tag:
   ```bash
   git tag -a $ARGUMENTS -m "Release $ARGUMENTS"
   ```
6. Show the user a summary of changes since the last tag:
   ```bash
   git log $(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || git rev-list --max-parents=0 HEAD)..HEAD --oneline
   ```
7. Ask the user to confirm before pushing the tag and creating the release.
8. Push the tag: `git push origin $ARGUMENTS`
9. Create a GitHub release with auto-generated notes:
   ```bash
   gh release create $ARGUMENTS --generate-notes
   ```

Rules:
- NEVER push without user confirmation
- Version must start with "v" (e.g., v1.0.0)
- If no version argument is provided, ask the user for one
