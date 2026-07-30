# Changesets

Add a changeset to every pull request that changes the published
`iris-analytics` package:

```bash
pnpm changeset
```

Choose the SemVer impact and write a user-facing summary. Use:

- `patch` for backwards-compatible fixes and optimizations.
- `minor` for backwards-compatible features.
- `major` for breaking API or behavior changes.

Changes that do not affect the published SDK, such as dashboard, marketing,
backend, or documentation-only work, do not need a changeset.

After changesets reach `main`, the release workflow maintains a version PR.
Merging that PR starts the approval-gated npm publish.
