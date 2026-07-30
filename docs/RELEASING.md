# Releasing `iris-analytics`

Iris uses Changesets for version intent and changelogs, GitHub Actions for
release orchestration, npm Trusted Publishing for short-lived OIDC
authentication, and GitHub Releases for source tags and release notes.

## One-time repository setup

The workflow expects these settings:

1. Create a GitHub environment named `npm-production`.
2. Require `VatsalP117` to approve deployments to that environment. Keep
   self-approval allowed while Iris has one maintainer.
3. On the npm package's **Settings → Trusted publishing** page, add a GitHub
   Actions publisher with:
   - Organization or user: `VatsalP117`
   - Repository: `iris`
   - Workflow filename: `release.yml`
   - Environment: `npm-production`
   - Allowed action: `npm publish`
4. After one successful automated release, restrict traditional token-based
   publishing and revoke any obsolete write tokens.

The workflow has `id-token: write`; no `NPM_TOKEN` secret is required.

## Preparing a release

Every pull request that changes the published SDK should include:

```bash
pnpm changeset
```

Select `patch`, `minor`, or `major` and write a user-facing summary. Backend,
dashboard, marketing, and documentation-only changes do not need a changeset.

When the pull request merges into `main`, `.github/workflows/release.yml`
creates or updates a `chore: release packages` pull request. That pull request
contains the accumulated version bump and changelog.

GitHub may initially hold CI runs for a pull request created by
`github-actions[bot]`. If the pull request shows an approval banner, select
**Approve workflows to run** before merging it.

## Publishing

1. Review and merge the `chore: release packages` pull request.
2. Open the **Release SDK** workflow run.
3. Review the package version and approve the `npm-production` deployment.
4. The workflow builds the SDK, inspects the npm tarball, publishes it, and
   creates a matching GitHub Release and Git tag.
5. Verify the version:

   ```bash
   npm view iris-analytics version dist-tags
   ```

Package versions are immutable. If a published release is defective, create a
new patch release. Do not attempt to overwrite the existing version.

## Prereleases

For larger changes, use Changesets prerelease mode and publish under a
non-default npm dist-tag such as `next` or `beta`. Do not move experimental
builds onto `latest`.

## Manual fallback

Use the local fallback only when GitHub Actions or trusted publishing is
unavailable:

```bash
pnpm release:sdk
```

The npm account must be authenticated and satisfy the package's 2FA policy.
