---
name: summary
description: 'Generate a topical summary of email threads. Use when asked to create a summary, topic index, or overview of emails about a specific topic. Runs the email-linearize summary subcommand.'
argument-hint: 'Topic name, e.g. "axiom" or "roadmap"'
---

# Topic Summary

Generate a topic index page and summary page for email threads matching a given topic.

## When to Use

- User asks to summarize emails about a specific topic
- User wants an index of all emails related to a subject
- User says `/summary <topic>` or `/summary "<topic>"`

## Procedure

1. Extract the topic from the user's input. The topic is the text after `/summary`.
2. Run the summary subcommand:

```sh
./email-linearize summary --topic "<topic>" --output ./output
```

3. Report the result: two files are generated in `./output/`:
   - `<slug>-index.html` — flat list of all matching email threads with dates
   - `<slug>.html` — summary page with threads grouped by month, including dates, snippets, and hyperlinks

4. If the binary is not built, build it first:

```sh
./scripts/build.sh
```

5. If no threads are found, let the user know and suggest trying a broader or different topic.

## Notes

- The topic is matched case-insensitively against email thread titles and filenames
- The output directory must already contain HTML files (run `./email-linearize --format html` first if needed)
- Both generated files include the navigation banner and use the shared stylesheet
