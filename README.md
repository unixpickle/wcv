# wcv

A version of the `wc` command that shows intermediate output. Short for "`wc` verbose". This is useful for running `wc` in a pipeline after commands that produce a lot of output and take a long time to run.

Unlike `wc`, this program always assumes input is UTF-8, and it ignores the current locale as specified by environment variables.
