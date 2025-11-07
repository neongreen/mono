
go_library(
    name = "ghrelease",
    srcs = glob(["*.go"], exclude = ["*_test.go"]),
    deps = [
        "//lib/ghclient",
    ],
    visibility = ["PUBLIC"],
)

go_test(
    name = "ghrelease_test",
    srcs = glob(["*_test.go"]),
    deps = [":ghrelease"],
)
