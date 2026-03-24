package metrics

import (
	"regexp"
	"strings"
)

// CountSentences counts the number of sentences in text.
func CountSentences(text string) int {
	if len(strings.TrimSpace(text)) == 0 {
		return 0
	}
	// Split by sentence-ending punctuation
	re := regexp.MustCompile(`[.!?]+\s`)
	parts := re.Split(text, -1)
	count := len(parts)
	if count == 0 {
		return 1
	}
	return count
}

// CountParagraphs counts paragraphs separated by blank lines.
func CountParagraphs(text string) int {
	if len(strings.TrimSpace(text)) == 0 {
		return 0
	}
	paragraphs := 0
	inParagraph := false
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if inParagraph {
				inParagraph = false
			}
		} else {
			if !inParagraph {
				paragraphs++
				inParagraph = true
			}
		}
	}
	return paragraphs
}

// CountQuestions counts question marks in text.
func CountQuestions(text string) int {
	return strings.Count(text, "?")
}

// HasCodeBlock detects markdown code blocks.
func HasCodeBlock(text string) bool {
	return strings.Contains(text, "```") || strings.Contains(text, "~~~")
}

// HasCodeRef detects references to code (file:line, function names, etc).
func HasCodeRef(text string) bool {
	patterns := []string{
		`\w+\.\w+:\d+`,          // file.go:42
		`func\s+\w+`,            // func name
		`\w+\(\)`,               // function()
		`package\s+\w+`,         // package name
		`import\s+`,             // import
		`class\s+\w+`,           // class name
		`def\s+\w+`,             // python def
	}
	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, text); matched {
			return true
		}
	}
	return false
}

// HasURL detects URLs in text.
func HasURL(text string) bool {
	re := regexp.MustCompile(`https?://\S+`)
	return re.MatchString(text)
}

// CountTechnicalTerms counts technical terms density signals.
func CountTechnicalTerms(text string) int {
	lower := strings.ToLower(text)
	terms := []string{
		"api", "database", "schema", "endpoint", "middleware",
		"authentication", "authorization", "encryption", "protocol",
		"algorithm", "cache", "queue", "thread", "mutex", "semaphore",
		"latency", "throughput", "consistency", "availability",
		"partition", "replication", "sharding", "index",
		"transaction", "rollback", "migration", "webhook",
		"oauth", "jwt", "cors", "ssl", "tls", "dns",
		"grpc", "graphql", "rest", "websocket",
		"kubernetes", "docker", "terraform", "ci/cd",
		"microservice", "monolith", "serverless", "lambda",
	}
	count := 0
	for _, t := range terms {
		count += strings.Count(lower, t)
	}
	return count
}

// HasRolePersona detects role/persona assignment patterns.
func HasRolePersona(text string) bool {
	patterns := []string{
		`(?i)you are (a |an )?(\w+ )*(expert|engineer|architect|developer|analyst|specialist|consultant|senior|lead)`,
		`(?i)act as (a |an )?`,
		`(?i)role:?\s`,
		`(?i)as (a |an )?(senior|expert|principal)`,
	}
	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, text); matched {
			return true
		}
	}
	return false
}

// CountMultiStepIndicators counts multi-step sequence indicators.
func CountMultiStepIndicators(text string) int {
	lower := strings.ToLower(text)
	indicators := []string{
		"first", "then", "next", "finally", "after that",
		"step 1", "step 2", "step 3",
		"1.", "2.", "3.", "4.", "5.",
		"phase 1", "phase 2",
		"before", "afterwards", "subsequently",
	}
	count := 0
	for _, ind := range indicators {
		if strings.Contains(lower, ind) {
			count++
		}
	}
	return count
}

// CountConstraints counts constraint/requirement indicators.
func CountConstraints(text string) int {
	lower := strings.ToLower(text)
	constraints := []string{
		"must", "should", "shall", "require",
		"at least", "at most", "no more than", "no less than",
		"minimum", "maximum", "exactly", "only",
		"do not", "don't", "never", "always",
		"ensure", "guarantee", "constraint",
	}
	count := 0
	for _, c := range constraints {
		count += strings.Count(lower, c)
	}
	return count
}

// HasFilePath detects file paths.
func HasFilePath(text string) bool {
	patterns := []string{
		`[/~]\S+\.\w+`,           // /path/to/file.ext or ~/file.ext
		`\w+/\w+/\w+`,            // dir/subdir/file
		`\w+\.(go|py|js|ts|rs|java|yaml|yml|json|md|txt|sh)`, // file.ext
	}
	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, text); matched {
			return true
		}
	}
	return false
}
