subinclude("//third_party/go:go_defs")

go_binary(
    name = "prrun",
    srcs = glob(["*.go"]),
    deps = [
        "//lib/ghrelease",
    ],
    visibility = ["PUBLIC"],
)
