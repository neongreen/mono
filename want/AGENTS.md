# Agent Guidelines for Want

## Automation vs Manual Steps

### Core Principle

**want should automate as much as possible.** The tool should perform actions automatically whenever it can do so safely, rather than instructing the user to do them manually.

### Guidelines

1. **Automatic Actions**
   - File modifications (e.g., adding shell configuration lines)
   - Tool installations
   - Environment setup
   - Any action that can be safely performed programmatically

2. **Manual Steps - Explicitly Mark Them**
   - When a step MUST be performed by the user (e.g., because it requires sudo, or needs interactive input)
   - Mark these steps clearly with "(MANUAL)" or similar indication
   - Always collect and print all manual steps at the end of execution

3. **End-of-Process Summary**
   - Always print a summary at the end showing:
     - What was done automatically
     - What steps (if any) require manual action
   - This helps users understand what they need to do next

### Example: mise Installation

When installing mise:
- ✅ Automatically add mise activation to shell config
- ✅ Create shell config file if it doesn't exist
- ❌ Don't just print instructions asking the user to add it manually

The shell activation step should be:
- Performed automatically during installation
- Only marked as manual if the automatic addition fails

### Plan Display

When showing fulfillment plans:
- Only mark manual steps with "(MANUAL)" label
- Automatic steps have no label (cleaner output)
- This highlights what requires user action without cluttering the display

### Safe Commands and Confirmation

Some commands are read-only/safe and don't modify the system:
- Commands like `ps`, `ls`, `git status`, etc.
- When a plan contains ONLY safe steps, no confirmation is needed
- Installation steps and other modifications still require confirmation
- Safe commands are identified by the `isCommandSafe()` function

**Current Limitations:**
- The safe command list is conservative and based on command names only
- Commands that CAN modify state with certain flags (e.g., `find -delete`, `date -s`) are excluded entirely for safety
- Future improvement: Implement argument inspection to allow safe uses of potentially dangerous commands

### History

This guideline was created in response to issue where `want --dry-run mise` showed Step 2 (adding shell activation) but didn't actually perform it automatically during real execution. The user expectation was that want should:
1. Do this step automatically
2. Only mark steps as manual when they truly need user action
3. Provide a summary of manual steps at the end

------------------------------------------------------------

## Code Formatting

**All Go code must be formatted with `go fmt` before work is considered complete.**

Before submitting any changes:
- Run `go fmt ./...` in the want directory
- Ensure all Go files are properly formatted
- This applies to both new and modified Go code
