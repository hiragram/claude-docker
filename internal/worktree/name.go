package worktree

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

const (
	dictPath   = "/usr/share/dict/words"
	maxWordLen = 6
	wordCount  = 3
)

// GenerateName creates a random branch name by picking 3 short words
// from the system dictionary, joined by hyphens.
func GenerateName() (string, error) {
	f, err := os.Open(dictPath)
	if err != nil {
		return "", fmt.Errorf("opening dictionary %s: %w", dictPath, err)
	}
	defer func() { _ = f.Close() }()

	var words []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		w := strings.TrimSpace(scanner.Text())
		if len(w) > 0 && len(w) <= maxWordLen {
			words = append(words, strings.ToLower(w))
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading dictionary: %w", err)
	}
	if len(words) < wordCount {
		return "", fmt.Errorf("not enough words in dictionary (found %d, need %d)", len(words), wordCount)
	}

	picked := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		picked[i] = words[rand.Intn(len(words))]
	}
	return strings.Join(picked, "-"), nil
}
