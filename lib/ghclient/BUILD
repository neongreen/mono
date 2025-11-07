subinclude("//third_party/go:go_defs")

go_library(
    name = "ghclient",
    srcs = glob(["*.go"], exclude = ["*_test.go"]),
    visibility = ["PUBLIC"],
)

go_test(
    name = "ghclient_test",
    srcs = glob(["*_test.go"]),
    deps = [":ghclient"],
)
