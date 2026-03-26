# promptlint

Score prompt complexity, route to the right model. No LLM required.

## API

### CLI (pipe-friendly)

```bash
# Analyze: full metrics
echo "Fix the bug in server.go" | promptlint analyze

# Route: just model name
echo "Fix the bug in server.go" | promptlint analyze --output-model

# Exit codes for pipelines (0=haiku, 1=sonnet, 2=opus)
echo "Fix typo" | promptlint analyze --exit-code
```

### HTTP

```bash
# Start server
promptlint serve 8090

# POST /analyze - score a prompt
curl -X POST http://localhost:8090/analyze \
  -H "Content-Type: application/json" \
  -d '{"text":"Design a payment gateway with retry logic"}'

# GET /health - server status
curl http://localhost:8090/health
```

## Output Format

```json
{
  "length": 37,
  "words": 8,
  "sentences": 1,
  "paragraphs": 1,
  "has_code_block": false,
  "has_code_ref": true,
  "has_url": false,
  "has_file_path": true,
  "questions": 0,
  "action": "fix",
  "domain": {"code": 0.9, "architecture": 0.1},
  "complexity": "low",
  "suggested_model": "haiku"
}
```

## Complexity Levels

```
complexity: "low"    -> suggested_model: "haiku"   (score 0-1)
complexity: "medium" -> suggested_model: "sonnet"  (score 2)
complexity: "high"   -> suggested_model: "opus"    (score 3+)
```

## Scoring Factors

- More than 200 words: +2
- More than 50 words: +1
- More than 5 sentences: +1
- More than 2 questions: +1
- Code blocks present: +1
- Multiple active domains: +2

## Integration

### With ccproxy (Claude Code Proxy)

```yaml
# ccproxy.yaml
rules:
  - name: "promptlint_scoring"
    endpoint: "http://localhost:8090/analyze"
    target_models:
      low: "haiku"
      medium: "sonnet"
      high: "opus"
```

### With agent pipeline

```bash
# Guard -> Route -> Execute -> Validate -> Track
echo "prompt" | seclint check --max-rating 16 \
  && echo "prompt" | promptlint analyze \
  | extract_model \
  | route_to_agent
```

## Install

```bash
go install github.com/mikeshogin/promptlint/cmd/promptlint@latest
```

## Metrics Analyzed

- **Length/words** - character and word count
- **Sentences** - period/question mark detection
- **Code blocks** - triple backtick detection
- **Code refs** - file path/line number patterns (e.g., "server.go:42")
- **Action type** - verb classification (fix=simple, design=complex)
- **Domain** - keyword classification (code/architecture/infra)
- **Questions** - question mark counting
- **Multi-domain** - cross-domain complexity

## Ecosystem

Part of the AI agent cost optimization ecosystem:

- **[seclint](https://github.com/mikeshogin/seclint)** - security/content classification
- **[archlint](https://github.com/mshogin/archlint)** - code architecture validation
- **[costlint](https://github.com/mikeshogin/costlint)** - token cost tracking

```
prompt -> promptlint (route) -> agent (execute) -> archlint (validate) -> costlint (track)
```

See [ECOSYSTEM.md](ECOSYSTEM.md) for full integration.

## For Humans

Promptlint is a lightweight tool that analyzes the complexity of your prompts and recommends which AI model would be most cost-effective to use. It looks at factors like word count, code snippets, and question count - without needing to call any LLM API - and routes simple tasks to cheaper models (like Haiku) and complex tasks to more capable ones (like Opus). This can save 40-60% on AI token costs when integrated with a proxy like ccproxy.
