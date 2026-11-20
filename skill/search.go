package skill

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"
)

type SearchMatch struct {
	Path    string `json:"path"`
	Section string `json:"section"`
	Content string `json:"content"`
}

type SearchResult struct {
	Query     string        `json:"query"`
	Matches   []SearchMatch `json:"matches"`
	Truncated bool          `json:"truncated"`
}

// Search uses the same embedded procedures and generated command references
// that MOI prepares in the session. It needs no account or running environment.
func Search(query string, limit int) (SearchResult, error) {
	result := SearchResult{Query: query, Matches: []SearchMatch{}}
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return result, fmt.Errorf("CLI_SKILL_SEARCH_QUERY_REQUIRED: use operation keywords, for example channel create")
	}
	if limit < 1 {
		return result, fmt.Errorf("CLI_SKILL_SEARCH_LIMIT_INVALID: limit must be positive")
	}
	type candidate struct {
		match SearchMatch
		score int
	}
	var candidates []candidate
	err := fs.WalkDir(Files(), "references", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(name, ".md") {
			return err
		}
		data, err := Read(name)
		if err != nil {
			return err
		}
		for index, section := range strings.Split(string(data), "\n## ") {
			if index > 0 {
				section = "## " + section
			}
			section = strings.TrimSpace(section)
			title, _, _ := strings.Cut(section, "\n")
			title = strings.TrimSpace(strings.TrimLeft(title, "#"))
			text, heading := strings.ToLower(section), strings.ToLower(title)
			score := 0
			for _, term := range terms {
				if !strings.Contains(text, term) {
					score = 0
					break
				}
				score++
				if strings.Contains(heading, term) {
					score += 15
				}
			}
			if score == 0 {
				continue
			}
			if name == "references/index.md" {
				score += 4
			}
			candidates = append(candidates, candidate{SearchMatch{name, title, section}, score})
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	slices.SortStableFunc(candidates, func(left, right candidate) int {
		if left.score != right.score {
			return right.score - left.score
		}
		if order := strings.Compare(left.match.Path, right.match.Path); order != 0 {
			return order
		}
		return strings.Compare(left.match.Section, right.match.Section)
	})
	result.Truncated = len(candidates) > limit
	for _, item := range candidates[:min(limit, len(candidates))] {
		result.Matches = append(result.Matches, item.match)
	}
	return result, nil
}
