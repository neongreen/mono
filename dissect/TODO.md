# TODOs

- if there are types and functions defined in the same file, we don't add an import to get the type and as the result the extracted function will not compile
- can we extract mutually recursive functions? need a test for that
- should we do anything with empty leftover files? maybe not if they have comments or idk.
  but otherwise we should remove them.
- a file with just a single function should be renamed, not extracted
- bring back the "does it build" check in tests after extraction