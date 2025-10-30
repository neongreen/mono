"""Dagger module for tk project."""

import dagger
from dagger import dag, function, object_type


@object_type
class Tk:
    """Tk project CI/CD functions."""

    @function
    async def test(self) -> str:
        """Run tk tests with coverage.
        
        Returns:
            Test output as a string.
        """
        return await (
            dag.container()
            .from_("golang:1.23")
            .with_directory(
                "/src",
                dag.host().directory(".", exclude=[".git", "*.db", ".dagger"]),
            )
            .with_workdir("/src")
            .with_exec(["go", "test", "./...", "-coverprofile=coverage.out", "-covermode=atomic"])
            .stdout()
        )

    @function
    async def build(self) -> dagger.File:
        """Build tk binary.
        
        Returns:
            The compiled tk binary.
        """
        return (
            dag.container()
            .from_("golang:1.23")
            .with_directory(
                "/src",
                dag.host().directory(".", exclude=[".git", "*.db", ".dagger"]),
            )
            .with_workdir("/src")
            .with_exec(["go", "build", "-o", "tk", "."])
            .file("/src/tk")
        )

    @function
    async def fmt_check(self) -> str:
        """Check Go code formatting.
        
        Returns:
            Output from gofmt check.
        """
        return await (
            dag.container()
            .from_("golang:1.23")
            .with_directory(
                "/src",
                dag.host().directory(".", exclude=[".git", "*.db", ".dagger"]),
            )
            .with_workdir("/src")
            .with_exec(["sh", "-c", "gofmt -l . | tee /tmp/fmt-issues.txt && test ! -s /tmp/fmt-issues.txt"])
            .stdout()
        )
