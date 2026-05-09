# Releasing

This document is for project maintainers cutting a new release.

## One-time setup (first release only)

1. **Enable package permissions** for the repository:
   `Settings → Actions → General → Workflow permissions` → "Read and write permissions".
2. **Create the `gh-pages` branch** (will be done automatically on first
   release by `chart-releaser-action`, or manually with
   `git checkout --orphan gh-pages && git commit --allow-empty -m "init" && git push origin gh-pages`).
3. **Enable GitHub Pages** under `Settings → Pages` → source: `gh-pages` branch, `/` (root).
4. **Connect the repo to Zenodo** at <https://zenodo.org/account/settings/github/>.
   Toggle the `data-gravity-operator` switch to ON. Subsequent tagged releases
   will be archived automatically and assigned a DOI.
5. **Update `CITATION.cff` and `.zenodo.json`** with your real ORCID
   (currently `0000-0000-0000-0000` placeholder).

## Cutting a release

1. Update version metadata on a release branch:
   - `helm/data-gravity-operator/Chart.yaml`: bump `version` and `appVersion`
   - `helm/data-gravity-operator/values.yaml`: bump `image.tag` and `mockRucio.image.tag`
   - `CITATION.cff`: bump `version` and set `date-released`
   - `CHANGELOG.md`: add a section for the new version, list changes, update the
     reference links at the bottom
2. Open a PR titled "Release vX.Y.Z", merge after CI is green.
3. Tag the merge commit:
   ```bash
   git checkout main
   git pull
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```
4. The `Release` workflow fires and:
   - Runs the test suite as a final check
   - Builds and pushes a multi-arch (amd64 + arm64) container image to
     `ghcr.io/<owner>/data-gravity-operator:X.Y.Z` and `:latest`
   - Packages the Helm chart, attaches the `.tgz` to a new GitHub Release,
     and updates `gh-pages` so users can `helm repo add` the chart repository
5. Zenodo will pick up the new GitHub Release and mint a DOI within a few minutes.
6. Add the new DOI badge to `README.md` (the latest-version DOI badge URL stays
   constant; the version-specific DOI is shown on the Zenodo page).

## Verifying a release

```bash
# Image
docker pull ghcr.io/<owner>/data-gravity-operator:X.Y.Z

# Helm chart
helm repo add data-gravity https://<owner>.github.io/data-gravity-operator
helm repo update
helm search repo data-gravity
```
