package util

import (
	"log"
	"strings"
)

// RemoveThinkTags removes <think>...</think> tags from the content if present.
// It always looks for the closing </think> tag, regardless of opening tag,
// and removes all content before and including that tag.
func RemoveThinkTags(content string) string {
	// Always look for the closing </think> tag, regardless of opening tag
	endIndex := strings.Index(content, "</think>")
	if endIndex != -1 {
		// Remove everything before and including the </think> tag
		cleanedContent := strings.TrimSpace(content[endIndex+len("</think>"):])
		log.Printf("Removed thinking section, content now starts with: %.100s...", cleanedContent)
		return cleanedContent
	}
	return content
}
