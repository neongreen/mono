# Adding a New Test - Example

To add a new test case, simply create two files:

## Step 1: Create input file
Create `testdata/my-new-test.input.md`:
```markdown
# My Test

This is a test. It has multiple sentences.
```

## Step 2: Create expected output file
Create `testdata/my-new-test.output.md`:
```markdown
# My Test

This is a test.
It has multiple sentences.
```

## Step 3: Run the test
```bash
go test -v -run TestFormatMarkdownFromFiles/my-new-test
```

That's it! No code changes needed. The test framework automatically:
- Discovers your new test files
- Runs the formatter on the input
- Compares with expected output
- Shows clear diffs if there's a mismatch

## Example of a failing test output

If the output doesn't match, you'll see:
```
--- FAIL: TestFormatMarkdownFromFiles/my-new-test (0.00s)
    main_test.go:733: formatMarkdown() mismatch (-want +got):
          string(
        - 	"Expected output here\n",
        + 	"Actual output here\n",
          )
```

The `-want` shows what was expected, and `+got` shows what was actually produced.
