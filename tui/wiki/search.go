package wiki

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/sahilm/fuzzy"
)

// SearchResult is one match from a wiki search.
type SearchResult struct {
	WikiName string
	File     string // relative path within wiki root
	Line     int
	Excerpt  string
}

// Search performs fuzzy full-text search over all .md files in wikiRoot/wiki/.
func Search(wikiRoot, wikiName, query string) []SearchResult {
	var lines []indexedLine
	wikiDir := filepath.Join(wikiRoot, "wiki")

	filepath.Walk(wikiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, _ := filepath.Rel(wikiRoot, path)
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			lines = append(lines, indexedLine{
				wikiName: wikiName,
				file:     rel,
				line:     lineNum,
				text:     scanner.Text(),
			})
		}
		return nil
	})

	if len(lines) == 0 {
		return nil
	}

	// Build a string slice for fuzzy matching
	texts := make([]string, len(lines))
	for i, l := range lines {
		texts[i] = l.text
	}

	matches := fuzzy.Find(query, texts)
	results := make([]SearchResult, 0, len(matches))
	for _, m := range matches {
		l := lines[m.Index]
		results = append(results, SearchResult{
			WikiName: l.wikiName,
			File:     l.file,
			Line:     l.line,
			Excerpt:  excerpt(l.text, 120),
		})
	}
	return results
}

// CrossSearch searches multiple wikis and merges results.
func CrossSearch(wikis []Wiki, query string) []SearchResult {
	var all []SearchResult
	for _, w := range wikis {
		all = append(all, Search(w.Path, w.Name, query)...)
	}
	return all
}

// AssembleCrossWikiContext builds a prompt context from index.md + overview.md
// for all wikis sharing a given topic tag.
func AssembleCrossWikiContext(wikis []Wiki, topic, query string) string {
	var sb strings.Builder
	for _, w := range wikis {
		hasTopic := false
		for _, t := range w.Topics {
			if t == topic {
				hasTopic = true
				break
			}
		}
		if !hasTopic {
			continue
		}
		sb.WriteString("## Wiki: ")
		sb.WriteString(w.Name)
		sb.WriteString(" (")
		sb.WriteString(w.Domain)
		sb.WriteString(")\n\n")

		index := ReadFile(w.Path, "wiki/index.md")
		if index != "" {
			sb.WriteString("### index.md\n\n")
			sb.WriteString(index)
			sb.WriteString("\n\n")
		}
		overview := ReadFile(w.Path, "wiki/overview.md")
		if overview != "" {
			sb.WriteString("### overview.md\n\n")
			sb.WriteString(overview)
			sb.WriteString("\n\n")
		}
	}
	if query != "" {
		sb.WriteString("---\n\n**Query:** ")
		sb.WriteString(query)
		sb.WriteString("\n")
	}
	return sb.String()
}

type indexedLine struct {
	wikiName string
	file     string
	line     int
	text     string
}

func excerpt(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
