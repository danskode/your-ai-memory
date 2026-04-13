package wiki

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Stats holds page counts and recent activity for a wiki.
type Stats struct {
	// PageCounts maps category subdirectory name to number of .md files.
	PageCounts map[string]int
	// TotalPages is the sum of all PageCounts.
	TotalPages int
	// RecentLogEntries holds the last few rows from wiki/log.md (raw markdown).
	RecentLogEntries []string
}

var categories = []string{
	"concepts", "patterns", "papers", "people", "connections", "questions",
}

// ReadWikiStats scans a wiki directory and returns Stats.
func ReadWikiStats(wikiRoot string) Stats {
	s := Stats{PageCounts: make(map[string]int)}

	for _, cat := range categories {
		dir := filepath.Join(wikiRoot, "wiki", cat)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		count := 0
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				count++
			}
		}
		s.PageCounts[cat] = count
		s.TotalPages += count
	}

	s.RecentLogEntries = readRecentLogEntries(filepath.Join(wikiRoot, "wiki", "log.md"), 5)
	return s
}

func readRecentLogEntries(logPath string, n int) []string {
	f, err := os.Open(logPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Table data rows start with | followed by a date-like value
		if strings.HasPrefix(line, "| 20") {
			lines = append(lines, line)
		}
	}

	if len(lines) <= n {
		return lines
	}
	return lines[len(lines)-n:]
}

// ReadFile reads a file from the wiki and returns its content, or empty string on error.
func ReadFile(wikiRoot, relPath string) string {
	data, err := os.ReadFile(filepath.Join(wikiRoot, relPath))
	if err != nil {
		return ""
	}
	return string(data)
}
