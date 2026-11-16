# open-iga

Open-source identity governance and administration (IGA). A serious, long-lived project, not a prototype.

- `core/`: Go service. Connectors run as WASM modules; workflows run on DBOS Transact Go.
- `console/`: TypeScript web app. The UI must always follow the designer's design. If the design is not clear, tell me and do not decide on your own.
- `openapi-spec/`: the API contract. Go and console types are generated from it.

## How to work

- One small step at a time. Keep prose short.
- Write in Simplified Technical English.
- Do not change anything outside the task scope. Report out-of-scope problems only if they are major.
- Do not make design decisions for me. When a choice is mine, list the options with a recommendation and wait.
- Review findings (CodeRabbit, code review): check the finding against the code first, and say whether it is valid and what it costs before changing anything.
- I review all the work, so do not commit or push unless I ask.
- Commits: `feat:`, `fix:`, `refactor:` or `chore:` prefix, one line with a precise summary. Leave PR descriptions empty; CodeRabbit fills them in.
- Never create huge PRs. Use stacked PRs when a change has more than 1000 lines (not counting generated code) or more than 25 files.

## Code style

- Comments: one line, two at most, simple English, say what the code does.
- Descriptive names. No `ret0`, `tp`, `mp`. Generated code is exempt.
- Function names start with a verb: `toX` maps, `buildX` builds, `getX` gets.
- Blank line before `return`, and between consecutive `if` blocks.
- Favor the standard library over a new dependency.

## core (Go)

Run from `core/`:

- Lint: `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...` (cyclop <= 10, exhaustruct, wrapcheck)
- Format: `go run mvdan.cc/gofumpt@v0.12.0 -w internal`
- Test: `go test ./internal/...` (repository and engine tests need Docker for Postgres)
- Generate the API: `go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.0 -config oapi-codegen.yaml ../openapi-spec/openapi.yml`, then `bun run generate:openapi-fetch` in `console/`
- Generate SQL: `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate`
- Generate mocks: `go run go.uber.org/mock/mockgen@v0.6.0 -source=internal/contract/runtime_{application,remote,repository}.go -destination=internal/testutil/runtime_*_mock.go -package=testutil`

Rules:

- Tests cover a use case of the feature. No tests for technical detail, no test fatigue.
- Every connector operation must be idempotent: DBOS steps run at least once.
- No users yet, so edit existing migrations directly.

## console (TypeScript)

- Typecheck: `bunx tsc --noEmit -p .`
- Test hooks, not views. Put logic in a hook and test the hook; `vitest.setup.ts` mocks the page wrapper, i18n and sonner.
- Forms use `<form action={fn}>`, not `onSubmit`.
- No nested ternaries. Use `ts-pattern` `match`.
