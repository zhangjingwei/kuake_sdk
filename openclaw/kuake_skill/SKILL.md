---
name: kuake_skill
description: Enables OpenClaw agents to interact with Quark Cloud Drive using the kuake CLI tool for file operations like listing, uploading, downloading, and sharing.
metadata:
  openclaw:
    requires:
      bins: ['kuake']
---

# Kuake Skill for OpenClaw

This skill allows OpenClaw agents to perform operations on Quark Cloud Drive using the `kuake` CLI tool. It translates user requests into appropriate `kuake` commands and executes them safely.

## Trigger Conditions

Activate this skill when users request operations related to Quark Cloud Drive, including:
- Listing directories and files
- Uploading files to the cloud
- Downloading files from the cloud
- Creating or managing shares
- Any file management operations on Quark Cloud Drive

## Target Behavior

When triggered:
1. Parse the user's intent to determine the appropriate `kuake` command
2. Validate parameters to ensure safety
3. Execute the `kuake` command using the `exec` tool
4. Return the results to the user

## Security Boundaries

- Only execute `kuake` commands with validated parameters
- Do not execute arbitrary shell commands
- Ensure all file paths and parameters are properly sanitized
- Require `kuake` binary to be available in PATH

## Example Command Patterns

- List root directory: `kuake list "/"`
- Upload a file: `kuake upload "local_file.txt" "/remote_file.txt"`
- Download a file: `kuake download "/remote_file.txt" "local_file.txt"`
- Create a share: `kuake share create "/file.txt"`

## Prerequisites

For typical OpenClaw use:

1. Install the **published `kuake` binary** for your OS from the project [Releases](https://github.com/zhangjingwei/kuake_cli/releases) page, and ensure it is on the **same machine’s `PATH`** where OpenClaw runs shell/`exec` for the agent (check with `kuake version` or `kuake -v`).
2. Configure Quark credentials the way your OpenClaw setup expects (commonly environment variable `KUAKE_COOKIE`; details in the project CLI docs).
3. Load this skill by pointing OpenClaw at the folder that contains this `SKILL.md` (see your OpenClaw `skills.load.extraDirs` or equivalent). No project checkout or compiler is required on that machine.

If `kuake` is not found, help the user add the install folder to `PATH` or call `kuake` by full path. The binary does not read `KUAKE_PATH`; only `PATH` or an explicit path to `kuake` applies.