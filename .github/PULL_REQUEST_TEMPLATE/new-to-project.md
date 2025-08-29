<!--
Use this if you're new to the repo, the tooling, or the app's domain.
It's a teaching checklist: follow it step-by-step.
Title: <type>(<scope>): <summary>
-->

## Summary

<!-- What does this PR do and why? 1–3 sentences -->

## Linked Issues

<!-- If any -->

Closes #

## How to Test

<!-- How to test this PR for reviewers to check -->

---

## Required (must pass)

- [ ] Lint & vet pass locally (`make lint`).
- [ ] Unit tests pass (`make test`) and run with race detector (`make test-race`).
- [ ] Integration tests pass (`make test-integration`) **or** marked N/A with reason.
- [ ] Project builds with no warnings (`make build`).
- [ ] I did a self-review and removed dead/debug code.
- [ ] No secrets/keys/tokens introduced.
- [ ] CLI remains backward compatible (commands/flags/help), **or** I documented the breaking change in this PR and updated docs/examples.

## Correctness & Design

- [ ] Clear naming; avoided duplicated logic by extracting helpers where sensible.
- [ ] Errors include context, don't just `return err`.
- [ ] Edge cases handled (empty inputs, timeouts, missing binaries/tools, invalid paths).
- [ ] Long operations have cancellation or timeouts; network/IO has retries/backoff.

## Security & Dependencies

- [ ] No secrets/keys/tokens introduced.
- [ ] Inputs validated; outputs escaped where appropriate.

## Performance & Reliability (as applicable)

- [ ] Concurrency is safe (tests pass with `-race`).

## Documentation & Maintainability

- [ ] README/help/examples updated for UX-visible changes.
- [ ] Feature flags/toggles documented (default state + removal plan).

## Reviewer Notes

<!-- Notes for the reviewer -->
