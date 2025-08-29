<!--
Use this if you're new to the repo, the tooling, or the app's domain.
It's a teaching checklist: follow it step-by-step.
Title: <type>(<scope>): <summary>
-->

## Summary

<!-- What does this PR do and why? 1–3 sentences -->

## Linked Issues

<!-- If any -->

## How to Test

<!--
Step-by-step instructions for a reviewer to verify the changes.
Be explicit: commands to run, expected output, etc.
-->

---

## Quality Checklist

### Pre-commit & CI (automatic)

_(No need to re-run manually — pre-commit hook + CI pipeline handle this.)_

- [ ] Pre-commit hook passed (lint + unit tests).
- [ ] CI pipeline is green (lint, tests, build).

### Code & Design

- [ ] Code is self-reviewed, clean, and free of debug/log/test junk.
- [ ] Clear, consistent naming; no duplicated logic (helpers extracted when useful).
- [ ] Errors include context (`return fmt.Errorf("...: %w", err)`) rather than plain `return err`.
- [ ] Edge cases handled (empty inputs, timeouts, invalid paths, missing tools).
- [ ] Long operations have cancellation/timeouts; network/IO has retries/backoff.
- [ ] No secrets/keys/tokens introduced.
- [ ] **If a new dependency was added, I explained why it is necessary and why alternatives were not chosen.**

### Testing

- [ ] Added or updated at least one test for new/changed code.
- [ ] Verified tests pass locally (pre-commit hook covers this).
- [ ] Verified main use case manually (document steps in **How to Test**).

### Documentation & UX

- [ ] README/help/examples updated if user-visible behavior changed.
- [ ] Feature flags/toggles documented.
- [ ] CLI remains backward compatible, **or** breaking changes are documented in this PR.

### Final Self-Check

- [ ] I personally reviewed all changes before submitting.
- [ ] Removed leftover debug code, logs, TODOs, experimental bits.

---

## Reviewer Notes

<!-- Notes for the reviewer -->
