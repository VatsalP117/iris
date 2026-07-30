# Releasing `iris-analytics` to npm

This is the canonical release runbook for the `iris-analytics` package in
`web/`. Normal releases are made from `main`; do not run `npm publish` from a
developer machine.

Iris uses:

- Changesets to record version intent and build the changelog.
- The **Release SDK** GitHub Actions workflow to plan and publish releases.
- The `npm-production` GitHub environment for a manual deployment approval.
- npm Trusted Publishing (OIDC) for short-lived authentication without an
  `NPM_TOKEN`.
- GitHub Releases for a matching source tag and generated release notes.

## Normal release flow

### 1. Add a changeset to an SDK pull request

Any pull request that changes the package shipped from `web/` must include a
changeset:

```bash
pnpm changeset
```

Choose the SemVer impact:

- `patch`: backwards-compatible fixes, performance improvements, or internal
  changes that alter the shipped package.
- `minor`: backwards-compatible features or new public APIs.
- `major`: breaking API or behavior changes.

Write the summary for package users; it becomes part of the changelog.
Backend, dashboard, marketing, CI, and documentation-only changes do not need a
changeset.

Commit the generated `.changeset/*.md` file with the implementation. Before
merging, verify the SDK locally:

```bash
pnpm install --frozen-lockfile
pnpm --filter iris-analytics build
pnpm pack:sdk
```

### 2. Merge the SDK pull request

Merge the reviewed SDK pull request into `main`. The
`.github/workflows/release.yml` workflow detects the pending changeset and
creates or updates a pull request titled:

```text
chore: release packages
```

That pull request consumes the pending changesets, updates
`web/package.json`, and updates the package changelog. Multiple SDK pull
requests can accumulate in the same release pull request.

GitHub may initially hold CI runs for a pull request created by
`github-actions[bot]`. If an approval banner appears, select
**Approve workflows to run**.

### 3. Review and merge the release pull request

Confirm that the release pull request has:

- The expected `iris-analytics` version.
- The correct patch, minor, or major bump.
- Clear changelog entries for every included SDK change.
- Passing required checks, or an intentional administrator override while the
  temporary admin bypass remains enabled.

Merge the release pull request into `main`. This merge is the publish trigger.

### 4. Approve the production deployment

Open **GitHub → Actions → Release SDK** for the merge commit. The publish job
targets the `npm-production` environment and waits for its required reviewer.

Review the commit and version, then approve the deployment. The workflow will:

1. Install dependencies from the lockfile.
2. Build `iris-analytics`.
3. run `npm pack --dry-run` to inspect the package contents.
4. Request a short-lived npm credential using GitHub OIDC.
5. Publish the package with the `latest` dist-tag.
6. Create a Git tag and GitHub Release named
   `iris-analytics@<version>`.

The workflow first checks npm and skips publishing when that exact version
already exists. This makes rerunning the planning workflow safe.

### 5. Verify the release

Check npm:

```bash
npm view iris-analytics version dist-tags
npm view iris-analytics@<version> dist.integrity
```

Then confirm:

- The new version appears on
  [npm](https://www.npmjs.com/package/iris-analytics).
- `latest` points at the intended version.
- The matching GitHub Release and tag exist.
- A clean test project can install the version when the release changes
  packaging, exports, or runtime behavior.

Package versions are immutable. If a published release is defective, fix it
and make a new patch release; never attempt to overwrite the existing version.

## One-time configuration

The automated workflow depends on matching values in GitHub and npm.

### GitHub

Create an environment named `npm-production` with:

- Required reviewer: `VatsalP117`.
- Protected branches only.
- Self-approval allowed while Iris has one maintainer.

The workflow's publish job must use that environment and grant
`id-token: write`. Repository Actions permissions must allow the release
workflow to create the version pull request and GitHub Release.

### npm

On **iris-analytics → Settings → Trusted Publisher**, configure:

| Field | Value |
| --- | --- |
| Provider | GitHub Actions |
| Organization or user | `VatsalP117` |
| Repository | `iris` |
| Workflow filename | `release.yml` |
| Environment | `npm-production` |
| Allowed action | `npm publish` |

The workflow filename and environment are exact, case-sensitive trust
boundaries. The repository currently has this connection configured. No
`NPM_TOKEN` repository secret is required.

After the first successful OIDC release, choose npm's most restrictive
token-publishing option and revoke obsolete npm write tokens. Trusted
Publishing continues to work with that setting.

## Troubleshooting

### No release pull request appears

- Confirm that the SDK pull request included a `.changeset/*.md` file.
- Open the **Release SDK** run for the merge into `main`.
- Check that Actions can create pull requests and write repository contents.

### The publish job is waiting

Open the workflow run and approve its `npm-production` environment deployment.
Only the configured reviewer can approve it.

### npm rejects the OIDC identity

Compare all four identity fields with the npm Trusted Publisher settings:
owner, repository, workflow filename, and environment. Also confirm the
publish job has `id-token: write`.

### The version already exists

Do not republish it. Confirm whether the version is already healthy on npm. If
another change is needed, add a patch changeset and run the normal flow again.

## Emergency manual fallback

Use local publishing only during an npm or GitHub Actions outage, with explicit
maintainer approval:

```bash
pnpm release:sdk
```

The local npm account must be authenticated and satisfy the package's 2FA
policy. Manually create the matching Git tag and GitHub Release afterward, and
record why the automated path was unavailable.

## Prereleases

The current workflow publishes stable releases to `latest`. Before publishing
a prerelease, update the workflow deliberately to use Changesets prerelease
mode and a non-default npm dist-tag such as `next` or `beta`. Never place an
experimental build on `latest`.
