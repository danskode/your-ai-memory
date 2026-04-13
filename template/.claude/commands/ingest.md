Run the INGEST protocol on the file provided as the argument.

Usage: /ingest <filename>

The filename should be a path relative to the `raw/` directory, e.g.:
  /ingest articles/my-article.md
  /ingest papers/attention-is-all-you-need.pdf
  /ingest videos/lex-fridman-ep-400.md

Steps (follow the INGEST Protocol defined in CLAUDE.md exactly):

1. Read the full content of `raw/$ARGUMENTS`
2. Extract all distinct knowledge units (concepts, patterns, people, questions)
3. For each unit — create a new page or update an existing one in the correct
   `wiki/` subdirectory, with proper frontmatter
4. Update `wiki/index.md` — add/update one row per affected page
5. Append a row to `wiki/log.md` with today's date, source filename, counts
6. Print an ingest summary: source, pages created, pages updated, questions raised

If `$ARGUMENTS` is empty, list all files in `raw/` that are not yet referenced
by any page in the wiki and ask which one to ingest.
