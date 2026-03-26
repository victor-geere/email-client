---
description: "Build the project: sync SPEC.md to prompt-tree.json, implement pending tasks, build, test, and iterate on errors. Maintains state to avoid duplicating work across restarts."
agent: "agent"
---

You are executing the `/build` command for the Email Thread Linearizer project.

## Context Files — Read These First

Before doing anything, read these files to understand the current state:

1. [SPEC.md](../../SPEC.md) — the editable spec (source of truth for what to build)
2. [prompt-tree.json](../../prompt-tree.json) — the implementation plan with task statuses
3. `.build-state.json` — persistent build state (create it if it doesn't exist)
4. [context/ARCHITECTURE.md](../../context/ARCHITECTURE.md) — module layout and data flow
5. [context/CONVENTIONS.md](../../context/CONVENTIONS.md) — code style rules
6. [context/SECURITY.md](../../context/SECURITY.md) — security requirements
7. [context/TESTING.md](../../context/TESTING.md) — test strategy

## Build State Management

Maintain `.build-state.json` at the project root to track progress across restarts:

```json
{
  "lastRun": "ISO-8601 timestamp",
  "completedTasks": ["task.id", ...],
  "failedTasks": {"task.id": "error message", ...},
  "buildIterations": 0,
  "lastPhaseCompleted": "phase.id or null",
  "status": "in-progress | completed | failed-with-rca"
}
```

**On startup**: read `.build-state.json`. If it exists, skip tasks listed in `completedTasks`. Resume from where the last run stopped. This is how you avoid duplicating work across restarts.

## Execution Steps

### Step 1: Sync SPEC → Prompt Tree

Compare `SPEC.md` checkboxes against `prompt-tree.json` task statuses:
- If a SPEC checkbox is `[x]` but the corresponding prompt-tree task is `pending`, update the task to `done`
- If SPEC has new unchecked items not in the prompt tree, note them but do not auto-add (flag for the user)
- Write the updated `prompt-tree.json` back to disk

### Step 2: Identify Next Tasks

Walk the prompt tree in dependency order. Find the next batch of tasks where:
- Status is `pending`
- All `dependsOn` tasks are `done`
- The task is NOT in `.build-state.json` `completedTasks`

### Step 3: Implement Each Task

For each task in the batch:

1. **Mark in-progress**: Update the task status in `prompt-tree.json` to `in-progress`
2. **Read the task prompt**: The `prompt` field contains detailed implementation instructions
3. **Implement**: Create or edit files as specified. Follow:
   - `context/CONVENTIONS.md` for code style
   - `context/SECURITY.md` for security rules (never log tokens/email bodies)
   - `context/TESTING.md` for test patterns
4. **Mark done**: Update task status to `done` in `prompt-tree.json`
5. **Update state**: Add task ID to `.build-state.json` `completedTasks`
6. **Update SPEC**: Check off corresponding `[ ]` items in `SPEC.md` → `[x]`

### Step 4: Build and Test

After implementing a phase's tasks:

```sh
go build ./cmd/email-linearize
go test ./...
go vet ./...
```

### Step 5: Iterate on Errors (max 5 times)

If build or tests fail:

1. Increment `buildIterations` in `.build-state.json`
2. Read the error output carefully
3. Fix the root cause in the relevant source files
4. Re-run `go build ./cmd/email-linearize && go test ./... && go vet ./...`
5. If fixed, continue to next phase
6. If still failing after 5 iterations, go to Step 6

### Step 6: Root Cause Analysis (if stuck)

If errors persist after 5 iterations, generate `RCA.md` at the project root:

```markdown
# Root Cause Analysis

## Build/Test Failures

### Error 1
- **Location**: file:line
- **Error message**: ...
- **Attempted fixes**: ...
- **Root cause hypothesis**: ...
- **Suggested resolution**: ...

### Error 2
...

## State at Time of Failure
- Last completed phase: ...
- Failed tasks: ...
- Build iteration count: 5
```

Update `.build-state.json` with `"status": "failed-with-rca"`.

## Rules

- **Never skip tests**: Every task with tests must have passing tests before marking done.
- **Never log tokens or email bodies**: Per security guidelines.
- **One phase at a time**: Complete all tasks in a phase before moving to the next.
- **Update state files after every task**: This enables clean restarts.
- **If a dependency is broken**: Fix it before proceeding. Do not work around it.
- **Keep SPEC.md, prompt-tree.json, and .build-state.json in sync**: These three files are the single source of truth.

## Completion

When all tasks are `done`:
1. Run final `go build ./cmd/email-linearize && go test ./... && go vet ./...`
2. Update `.build-state.json` with `"status": "completed"`
3. Print summary: phases completed, tests passed, files created
