This command is for performing the requests related to Mise and especially Mise tasks. 

For example, if the user asks to add a new task:

> Add cloc --by-file --include-lang=Go .

You should add the task to the `mise.toml` file in the repository root.

You should intelligently guess which mise.toml file the task should be added to.
It can be the repository root, it can be an existing mise.toml file in the project directory, or it can even be a new mise file if a project doesn't have it yet. 

You should guess the name of the task, or propose several options if not obvious.
If you propose options:
- Don't add the task until the user picks the name.
- After the user picks the name, update this file (mise.md) with the command, proposed names, and user's decision.
This way you can become smarter over time.
