package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mikeshogin/promptlint/pkg/abtest"
	"github.com/mikeshogin/promptlint/pkg/analyzer"
	"github.com/mikeshogin/promptlint/pkg/config"
	"github.com/mikeshogin/promptlint/pkg/perf"
	"github.com/mikeshogin/promptlint/pkg/router"
	"github.com/mikeshogin/promptlint/pkg/server"
	"github.com/mikeshogin/promptlint/pkg/template"
	"github.com/mikeshogin/promptlint/pkg/trend"
	"github.com/mikeshogin/promptlint/pkg/validator"
)

// Exit codes for pipeline integration.
const (
	ExitHaiku  = 0
	ExitSonnet = 1
	ExitOpus   = 2
	ExitError  = 3
)

func modelExitCode(model string) int {
	switch model {
	case "haiku":
		return ExitHaiku
	case "sonnet":
		return ExitSonnet
	case "opus":
		return ExitOpus
	default:
		return ExitError
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: promptlint {analyze|validate|route|serve|ab|perf|trend|template}\n")
		fmt.Fprintf(os.Stderr, "\nanalyze flags:\n")
		fmt.Fprintf(os.Stderr, "  --output-model   print only model name\n")
		fmt.Fprintf(os.Stderr, "  --format=json    output format: json (default), brief\n")
		fmt.Fprintf(os.Stderr, "  --exit-code      use model-based exit codes (0=haiku,1=sonnet,2=opus)\n")
		fmt.Fprintf(os.Stderr, "\nroute:\n")
		fmt.Fprintf(os.Stderr, "  reads prompt from stdin, prints routing decision as JSON\n")
		fmt.Fprintf(os.Stderr, "\nvalidate:\n")
		fmt.Fprintf(os.Stderr, "  reads prompt from stdin, prints JSON array of violations\n")
		fmt.Fprintf(os.Stderr, "\nab flags:\n")
		fmt.Fprintf(os.Stderr, "  --config-b=FILE  path to alternative config YAML (variant B)\n")
		fmt.Fprintf(os.Stderr, "  --format=json    output format: json (default), brief\n")
		fmt.Fprintf(os.Stderr, "  reads one prompt per line from stdin\n")
		os.Exit(ExitError)
	}

	cmd := os.Args[1]

	switch cmd {
	case "analyze":
		// Parse flags
		outputModel := false
		exitCode := false
		format := "json"
		for _, arg := range os.Args[2:] {
			switch {
			case arg == "--output-model":
				outputModel = true
			case arg == "--exit-code":
				exitCode = true
			case strings.HasPrefix(arg, "--format="):
				format = strings.TrimPrefix(arg, "--format=")
			}
		}

		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(ExitError)
		}

		result := analyzer.Analyze(string(input))

		if outputModel {
			fmt.Println(result.SuggestedModel)
		} else {
			switch format {
			case "brief":
				fmt.Printf("complexity=%s model=%s words=%d action=%s\n",
					result.Complexity, result.SuggestedModel, result.Words, result.Action)
			default:
				out, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(out))
			}
		}

		if exitCode {
			os.Exit(modelExitCode(result.SuggestedModel))
		}

	case "serve":
		port := "8080"
		for _, arg := range os.Args[2:] {
			if strings.HasPrefix(arg, "--port=") {
				port = strings.TrimPrefix(arg, "--port=")
			} else if arg == "--port" {
				// handled by next iteration below
			}
		}
		// Support --port 8080 (space-separated)
		args := os.Args[2:]
		for i, arg := range args {
			if arg == "--port" && i+1 < len(args) {
				port = args[i+1]
			}
		}

		fmt.Fprintf(os.Stderr, "promptlint server on :%s\n", port)

		if err := http.ListenAndServe(":"+port, server.New()); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}

	case "route":
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(ExitError)
		}

		r := router.NewDefault()
		result := r.Route(strings.TrimSpace(string(input)))

		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))

	case "validate":
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(ExitError)
		}

		v := validator.New()
		results := v.Validate(string(input))

		// Always output a JSON array (empty array when no violations).
		if results == nil {
			results = []validator.ValidationResult{}
		}
		out, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(out))

		// Exit 1 if any errors were found.
		for _, r := range results {
			if r.Severity == validator.SeverityError {
				os.Exit(1)
			}
		}

	case "ab":
		configBPath := ""
		format := "json"
		for _, arg := range os.Args[2:] {
			switch {
			case strings.HasPrefix(arg, "--config-b="):
				configBPath = strings.TrimPrefix(arg, "--config-b=")
			case strings.HasPrefix(arg, "--format="):
				format = strings.TrimPrefix(arg, "--format=")
			}
		}

		cfgA := config.LoadOrDefault()

		var cfgB *config.Config
		if configBPath != "" {
			var err error
			cfgB, err = config.Load(configBPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config-b: %v\n", err)
				os.Exit(ExitError)
			}
		} else {
			// Default variant B: more aggressive routing - push medium to haiku, high to sonnet
			cfgB = &config.Config{
				Tiers: []config.Tier{
					{Name: "haiku", MaxComplexity: "medium", MaxTokens: 5000, CostPer1k: 0.80},
					{Name: "sonnet", MaxComplexity: "high", MaxTokens: 100000, CostPer1k: 3.00},
				},
				DefaultTier: "sonnet",
			}
		}

		test := abtest.New("routing-comparison", cfgA, cfgB)

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			test.Run(line)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(ExitError)
		}

		summary := test.Summary()

		switch format {
		case "brief":
			fmt.Printf("prompts=%d differs=%d same=%d cost_a=%.2f cost_b=%.2f cost_diff=%.1f%% cheaper=%s\n",
				summary.TotalPrompts,
				summary.DifferentRoutes,
				summary.SameRoutes,
				summary.AvgCostA,
				summary.AvgCostB,
				summary.CostDiffPct,
				summary.CheaperVariant(),
			)
			if summary.DifferentRoutes > 0 {
				fmt.Println("\nDiffering routes:")
				for _, cmp := range test.Results() {
					if cmp.Differs {
						fmt.Printf("  [%s] A->%s B->%s (score=%.2f)\n",
							cmp.PromptHash, cmp.A.RoutedModel, cmp.B.RoutedModel, cmp.A.Score)
					}
				}
			}
		default:
			type fullOutput struct {
				Summary     abtest.Summary      `json:"summary"`
				Comparisons []abtest.ABComparison `json:"comparisons"`
			}
			out, _ := json.MarshalIndent(fullOutput{
				Summary:     summary,
				Comparisons: test.Results(),
			}, "", "  ")
			fmt.Println(string(out))
		}

	case "perf":
		iterations := 100
		for _, arg := range os.Args[2:] {
			if strings.HasPrefix(arg, "--iterations=") {
				val := strings.TrimPrefix(arg, "--iterations=")
				if n, err := strconv.Atoi(val); err == nil && n > 0 {
					iterations = n
				}
			}
		}

		results := perf.RunAll(iterations)
		out, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(out))

	case "trend":
		trendCmd(os.Args[2:])

	case "template":
		templateCmd(os.Args[2:])

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: promptlint {analyze|validate|route|serve|ab|perf|trend|template}\n", cmd)
		os.Exit(1)
	}
}

// trendCmd implements the "trend" subcommand.
func trendCmd(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: promptlint trend {record|summary} [--format=json|text]\n")
		os.Exit(ExitError)
	}

	format := "text"
	subArgs := args[1:]
	for _, a := range subArgs {
		if strings.HasPrefix(a, "--format=") {
			format = strings.TrimPrefix(a, "--format=")
		}
	}

	log := trend.NewDefault()

	switch args[0] {
	case "record":
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(ExitError)
		}
		prompt := string(input)
		result := analyzer.Analyze(prompt)

		if err := log.Record(prompt, result.Complexity, result.PromptScore.Total, result.SuggestedModel); err != nil {
			fmt.Fprintf(os.Stderr, "Error recording trend: %v\n", err)
			os.Exit(ExitError)
		}

		switch format {
		case "json":
			type recordOutput struct {
				Complexity  string `json:"complexity"`
				Score       int    `json:"score"`
				ModelRouted string `json:"model_routed"`
				Recorded    bool   `json:"recorded"`
			}
			out, _ := json.MarshalIndent(recordOutput{
				Complexity:  result.Complexity,
				Score:       result.PromptScore.Total,
				ModelRouted: result.SuggestedModel,
				Recorded:    true,
			}, "", "  ")
			fmt.Println(string(out))
		default:
			fmt.Printf("recorded complexity=%s score=%d model=%s\n",
				result.Complexity, result.PromptScore.Total, result.SuggestedModel)
		}

	case "summary":
		summary, err := log.Summary()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading trend log: %v\n", err)
			os.Exit(ExitError)
		}

		switch format {
		case "json":
			out, _ := json.MarshalIndent(summary, "", "  ")
			fmt.Println(string(out))
		default:
			fmt.Printf("total_entries=%d avg_score=%.2f trend=%s last_7_avg=%.2f previous_7_avg=%.2f\n",
				summary.TotalEntries, summary.AvgScore, summary.Trend,
				summary.Last7Avg, summary.Previous7Avg)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown trend subcommand: %s\nUsage: promptlint trend {record|summary}\n", args[0])
		os.Exit(ExitError)
	}
}

// templateCmd implements the "template" subcommand.
func templateCmd(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: promptlint template score [--format=json|text] < template.txt\n")
		os.Exit(ExitError)
	}

	format := "text"
	for _, a := range args[1:] {
		if strings.HasPrefix(a, "--format=") {
			format = strings.TrimPrefix(a, "--format=")
		}
	}

	switch args[0] {
	case "score":
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(ExitError)
		}

		tmpl := template.ParseTemplate(string(input))
		ts := template.ScoreTemplate(tmpl)

		switch format {
		case "json":
			out, _ := json.MarshalIndent(ts, "", "  ")
			fmt.Println(string(out))
		default:
			fmt.Print(template.FormatScore(ts))
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown template subcommand: %s\nUsage: promptlint template score < template.txt\n", args[0])
		os.Exit(ExitError)
	}
}
