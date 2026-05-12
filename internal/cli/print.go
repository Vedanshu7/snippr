package cli

import (
	"fmt"
	"strings"
	"time"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
)

func PrintSnippetTable(snippets []Snippet) {
	if len(snippets) == 0 {
		fmt.Println(dim + "  No snippets found." + reset)
		return
	}

	fmt.Printf("%s  %-6s  %-36s  %-12s  %-20s  %s%s\n",
		bold, "ID", "TITLE", "LANGUAGE", "TAGS", "UPDATED", reset)
	fmt.Println(dim + "  " + strings.Repeat("─", 90) + reset)

	for _, s := range snippets {
		tags := strings.Join(s.Tags, ", ")
		if len(tags) > 20 {
			tags = tags[:17] + "..."
		}
		title := s.Title
		if len(title) > 36 {
			title = title[:33] + "..."
		}
		fmt.Printf("  %s%-6d%s  %-36s  %s%-12s%s  %-20s  %s\n",
			cyan, s.ID, reset,
			title,
			yellow, s.Language, reset,
			tags,
			humanTime(s.UpdatedAt),
		)
	}
}

func PrintSnippet(s *Snippet) {
	fmt.Printf("\n%s%s%s\n", bold, s.Title, reset)
	fmt.Printf("%s#%d  %s  %s%s\n",
		dim, s.ID, s.Language, strings.Join(s.Tags, " "), reset)
	if s.IsPublic && s.ShareSlug != "" {
		fmt.Printf("%sshare: /s/%s%s\n", dim, s.ShareSlug, reset)
	}
	fmt.Println()
	fmt.Println(s.Content)
	fmt.Println()
}

func Success(msg string) {
	fmt.Printf("%s✓%s %s\n", green, reset, msg)
}

func Fatal(msg string) {
	fmt.Printf("%s✗%s %s\n", red, reset, msg)
}

func humanTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 2 2006")
	}
}
