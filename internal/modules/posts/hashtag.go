package posts

import (
	"regexp"
	"strings"
)

var hashtagRegex = regexp.MustCompile(`#([a-zA-Z0-9_]+)`)

func extractHashtags(body string) []string {
	matches := hashtagRegex.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(matches))
	tags := make([]string, 0, len(matches))

	for _, match := range matches {
		tag := strings.ToLower(match[1])
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}

	return tags
}
