package main

// maxScannerCapacity is the buffer size for scanning large lines in trace files.
// Claude trace files can have very long lines, so we use a 10MB buffer.
const maxScannerCapacity = 10 * 1024 * 1024 // 10MB
