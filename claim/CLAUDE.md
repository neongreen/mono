# Claude Agent Instructions for claim

## Budget Authorization

**AUTHORIZED: Up to $100 per session for Claude API usage**

When working on the `claim` project, you are explicitly authorized to spend up to $100 per session on Claude API calls. This includes:

- Testing end-to-end functionality
- Running `claim check` on fixtures
- Running `claim golden` test suite
- Debugging and verification
- Any development or testing needs

**Do not hesitate to use real Claude API calls.** This is a claim-checking tool that requires Claude to function - testing with real API calls is not optional, it's mandatory.

## Expected Usage Patterns

### During Development
- Run `claim check` freely to test changes
- Use `--debug-prompt` to verify prompt construction
- Test multiple fixtures to ensure robustness
- Each check costs ~$0.01-0.10

### During Testing
- Run `claim golden` to verify all fixtures (~$0.10-1.00 for 9 checks)
- Test edge cases and complex claims
- Verify post-rule enforcement

### Cost Optimization Tips
- Use `--model haiku` for faster/cheaper testing if appropriate
- But default to sonnet for quality - cost is not the primary concern
- Don't batch checks unless it makes sense - individual checks are fine

## When NOT to Worry About Cost

- Testing a single fixture: ~$0.01-0.10 per check
- Running golden suite: ~$0.10-1.00 per run
- Debugging issues: Test as many times as needed
- Iterating on prompts: Each iteration is cheap

You have $100/session budget. A typical session might use $1-10 even with extensive testing.

## Development Workflow

1. Make changes to code
2. **Test with real Claude immediately** - don't use mocks for end-to-end verification
3. Run `claim check --claim <id> --root fixtures --lens-file lenses.md`
4. If it works, test another fixture
5. Run `claim golden` before marking tasks done
6. Document findings

## Remember

This tool's entire purpose is to use Claude to check claims. Testing without real Claude is like testing a database without a database - it's incomplete. Use real API calls liberally.
