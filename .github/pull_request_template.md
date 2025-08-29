<!--
Title: <type>(<scope>): <summary>
Examples: fix(auth): handle expired refresh token
          feat(cli): add --dry-run flag
-->

## Summary

<!-- 1–3 sentences: what and why -->

## Linked Issues

<!-- If any -->

Closes #

## How to Test

<!-- Copy/paste steps & example commands (include kube context/namespace if relevant) -->

---

## Checklist (Required)

- [ ] Builds locally with no errors (`make build`).
- [ ] Linter passes (`make lint`) and tests pass (`make test`).
- [ ] No secrets/keys/tokens added.
- [ ] CLI behavior and flags remain backward compatible **or** breaking changes are documented here.

## Checklist (Recommended)

- [ ] Docs/help/examples updated if behavior/flags changed.
- [ ] Edge cases considered (empty inputs, timeouts, invalid paths).

---

## Reviewer Notes

<!-- Files or areas to focus on, tradeoffs, known limitations -->
