# Verification Script for Kuake Skill

This script verifies that the kuake skill can be loaded and used with OpenClaw agents.

## Prerequisites

1. Ensure kuake executable is available:
   ```bash
   go run cmd/main.go --version
   ```

2. Ensure OpenClaw agent is installed and configured.

## Verification Steps

1. **Load the skill**:
   - Add the workspace as an OpenClaw workspace, or
   - Add `openclaw/kuake_skill/` to `skills.load.extraDirs` in OpenClaw config

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

If kuake is not available, the skill should guide the agent to prompt the user to install/build kuake first.