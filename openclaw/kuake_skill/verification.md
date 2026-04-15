# Verification Script for Kuake Skill

This script verifies that the kuake skill can be loaded and used with OpenClaw agents.

## Prerequisites

1. Install **`kuake`** from the project [Releases](https://github.com/zhangjingwei/kuake_cli/releases) for your platform, add it to `PATH`, then confirm:
   ```bash
   kuake version
   ```
   (Same as `kuake -v` / `kuake --version`.)

2. OpenClaw is installed and can run agents.

## Verification Steps

1. **Load the skill**: in OpenClaw config, set `skills.load.extraDirs` (or your product’s equivalent) to the **folder that contains this package’s `SKILL.md`**—for example a folder you downloaded or unzipped named `kuake_skill`. You only need that folder on disk; nothing else from the project tree is required.

2. **Restart OpenClaw gateway** or start a new session.

3. **Check skill loading**:
   ```bash
   openclaw skills list
   ```
   Verify that `kuake_skill` appears in the list.

4. **Test skill activation**:
   Send a message to the agent like: "List my Quark Cloud Drive root directory"
   The agent should respond by executing `kuake list "/"` and returning the results.

5. **Test other commands**:
   - "Upload file.txt to my Quark drive"
   - "Download /file.txt"
   - "Create a share for /file.txt"

## Expected Behavior

- Agent recognizes Quark-related requests
- Agent executes kuake commands safely
- Results are returned in a user-friendly format
- No arbitrary shell commands are executed

## Fallback

If `kuake` is not available, prompt the user to install it from **Releases**, extract if needed, and add it to `PATH` (or use the full path to the executable).