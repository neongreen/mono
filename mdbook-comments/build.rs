use std::env;
use std::path::Path;
use std::process::Command;

const ASSET_FILES: &[&str] = &[
    "js/comments-base.js",
    "js/comments-json-server-adapter.js",
    "js/comments-supabase-adapter.js",
    "js/comments-googlesheets-adapter.js",
    "js/comments-custom-adapter.js",
];

const WATCHED_SOURCES: &[&str] = &[
    "package.json",
    "pnpm-lock.yaml",
    "tsconfig.json",
    "vitest.config.ts",
    "playwright.config.ts",
    "src",
];

fn main() {
    for path in WATCHED_SOURCES {
        println!("cargo:rerun-if-changed={path}");
    }

    // Always rebuild the frontend assets when the build script runs so the embedded
    // bundles stay in sync with the TypeScript sources.
    build_frontend_assets();

    // Ensure Cargo also reruns this build script when any generated asset changes.
    // This gives us predictable rebuilds when the generated files are removed.
    for asset in ASSET_FILES {
        println!("cargo:rerun-if-changed={asset}");
    }
}

fn build_frontend_assets() {
    let manifest_dir =
        env::var("CARGO_MANIFEST_DIR").expect("CARGO_MANIFEST_DIR must be set by Cargo");

    let pnpm_status = Command::new("pnpm")
        .arg("build")
        .current_dir(&manifest_dir)
        .status()
        .unwrap_or_else(|err| {
            panic!(
                "failed to run `pnpm build` while generating frontend assets: {err}\n\
                 Ensure pnpm is installed (try `mise install` or `mise run //mdbook-comments:dev:setup`)."
            )
        });

    if !pnpm_status.success() {
        panic!(
            "`pnpm build` exited with status {pnpm_status}. \
             Install dependencies with `pnpm install` and try again."
        );
    }

    // Sanity check that the bundles were produced.
    let missing_assets: Vec<&str> = ASSET_FILES
        .iter()
        .copied()
        .filter(|asset| !Path::new(&manifest_dir).join(asset).exists())
        .collect();

    if !missing_assets.is_empty() {
        panic!(
            "frontend build did not produce expected bundles: {missing_assets:?}. \
             Verify the build scripts write to the `js/` directory."
        );
    }
}
