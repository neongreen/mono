
go_library(
    name = "toml",
    srcs = glob(["*.go"], exclude = ["*_test.go"]),
    visibility = ["PUBLIC"],
)

go_test(
    name = "toml_test",
    srcs = glob(["*_test.go"]),
    deps = [":toml"],
)
