---
description: "Generate a topical summary and index of email threads about a specific topic"
agent: "agent"
argument-hint: "Topic name, e.g. client or roadmap"
---

Generate a topic summary for the provided topic argument.

1. Make sure the binary is built: `./scripts/build.sh`
2. Run: `./email-linearize summary --topic "${input}" --output ./output`
3. Report the generated files and how many threads matched.

If the output directory has no HTML files yet, first run `./email-linearize --format html` and then run the summary command.
