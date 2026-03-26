---
description: "Use when writing or modifying quote detection, linearization, or reference-generation logic. Covers how quoted email content is identified, attributed, and replaced with references."
applyTo: "**/quote/**,**/linearize/**"
---
# Quote Detection and Linearization Guidelines

- Detect quoted content using multiple strategies combined, not one exclusively:
  - `<blockquote>` elements and known email client wrapper classes in HTML bodies
  - Line-prefix patterns (`>`, `> >`) in plain-text bodies
  - Header patterns ("On {date}, {name} wrote:", "From: … Sent: … Subject: …")
  - Content hashing — match quoted text against earlier messages in the same thread
- Replace each detected quote with a reference: `[→ see msg #N from Author, DD Mon]`.
- Trim trailing whitespace, blank lines, and signatures left behind after quote removal.
- Handle multi-level nesting: a reply quoting a reply that quotes the original.
- Preserve non-quoted content surrounding inline quotes — do not discard the author's own words.
- When a quote cannot be attributed to a specific earlier message, use a generic reference: `[→ quoted text omitted]`.
