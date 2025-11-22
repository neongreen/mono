# Errors And Validation

## Error handling

Validation and error handling:
- Invalid or empty topics are ignored silently (no explicit validation yet).
- Conflicting topic/section titles within the same comment group fail extraction with file:line.
- Comments that do not attach to package/func/type/const/var doc groups are skipped.
- CLI exits non-zero only when extraction fails (parse/metadata error) or generation fails (write error).
- Bad markers inside otherwise valid files do not stop extraction; they are just skipped.

*Source: `lion/internal/extractor/extractor.go:292`*

