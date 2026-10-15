# Chainguard Checks + Workspaces — setup & test record

What we did to stand up and validate **Chainguard Checks** and **Chainguard Workspaces** on
this repo (`rajeevvishal06/cg-checks-workspaces`, public, org `rajeevcorp`). Both launched
2026-10-14 as a Technology Preview.

- **Workspaces** = remote dev: `chainctl develop` opens a microVM session with your repo,
  tools, and Claude Code — holding **none of your long-lived credentials**.
- **Checks** = CI on Chainguard's infra: define checks in `.chainguard/ci.yaml`; each runs in
  its own locked-down microVM on every PR and reports as GitHub checks.
- One command, `chainctl develop init`, turns on **both**.

---

## ⚠️ The gotcha that cost us the most time: there are THREE Chainguard GitHub Apps

| App | Install URL | Role |
|-----|-------------|------|
| **Chainguard** | `github.com/apps/chainguard` | Backs **Checks + Workspaces** events. **MUST be installed on the repo** or no checks ever run. |
| **Chainguard Develop** | `github.com/apps/chainguard-develop` | Lets a Workspaces session **push/act on GitHub as you**. Installed by `chainctl develop init`. |
| Chainguard Guardener | `github.com/apps/chainguard-guardener` | Separate — for Guardener (Hardened Actions, etc.). **NOT** the app Checks uses. |

**We lost time because `cg-checks-workspaces` was added to the *Guardener* app, not the
`chainguard` app.** Symptom: a PR with a valid `ci.yaml` got **zero checks** and was
mergeable. Fix: install `github.com/apps/chainguard/installations/new` on the account **with
access to this repo**, then re-fire the PR event (reopen or push a commit).

---

## 0. Prerequisites
- `chainctl` ≥ 0.2.380 (`chainctl version`; update with `sudo chainctl update` / `brew upgrade chainctl`).
- Claude Code CLI installed + logged in (used by the session agent).
- A github.com repo you can push to + a local clone. **Public** avoids scope questions.
- Owner on the Chainguard org (gets the `sandbox.user` role).

## 1. Enable both — `chainctl develop init`
```bash
chainctl auth login            # if your session expired
chainctl develop init          # interactive; run in your own Terminal (full-screen TUI)
```
- Accept the legal-terms screen **promptly** (press Enter) — lingering drops the stream with
  `[Unavailable] stream timeout`; just re-run `chainctl develop init` if it does.
- It turns on `sandbox_checks` (500 vCPU-hr/mo) and `sandbox_workspaces` (500 vCPU-hr/mo,
  **1 session/user**, 3/org), and offers to **connect your GitHub account** (installs
  `chainguard-develop`) so sessions act as you.

## 2. Workspaces
```bash
chainctl develop --parent rajeevcorp        # open/attach a session (microVM at /work)
```
- With no `.chainguard/develop.yaml`/`ci.yaml` it opens on the **default environment**
  (bash, git, go? no — default has no `go`; see below).
- In the session: `run-ci` runs the repo's checks against the **working tree** (advisory,
  "verify before push"). The same runner the PR gate uses.
- **Session env is fixed at creation.** A session opened before `ci.yaml` existed has no `go`;
  either `apk add go` in it, or open a fresh session once `ci.yaml` is in the repo.
- **Credentials:** the VM holds no long-lived secrets — services are reached via local
  hostnames (`http://github/`, `http://claude/`) and creds are injected outside the VM. Full
  hands-on test plan in [`SESSION-CREDENTIALS-TEST.md`](SESSION-CREDENTIALS-TEST.md).

## 3. Checks
Add `.chainguard/ci.yaml` (this repo's Go example):
```yaml
version: 1
environment:
  packages: [go]
checks:
  vet:
    cmd: go vet ./...
    network:
      egress: [proxy.golang.org, sum.golang.org]
  test:
    cmd: go test ./...
    network:
      egress: [proxy.golang.org, sum.golang.org]
```
- Lint it: `chainctl checks validate .` → `ok: the document lowers`.
- Checks reads the config **from the PR itself**, so the PR that adds `ci.yaml` runs its own
  checks. On the PR you get `presubmit / ci / vet`, `presubmit / ci / test`, and the summary
  **Chainguard Checks (presubmit)**. Each runs in a fresh microVM (Chainguard OS) that
  installs `go`, with network limited to the listed egress hosts.

## 4. Require checks before merge (branch protection)
Require **only the summary** check (per-check results can be skipped and would hang):
```bash
gh api --method PUT repos/rajeevvishal06/cg-checks-workspaces/branches/main/protection --input - <<'JSON'
{ "required_status_checks": { "strict": false, "contexts": ["Chainguard Checks (presubmit)"] },
  "enforce_admins": false, "required_pull_request_reviews": null, "restrictions": null }
JSON
```
Verified: a PR with a deliberately failing test showed `mergeStateStatus = BLOCKED`.

## Gotchas we hit (summary)
1. **No checks at all** → the `chainguard` app wasn't on the repo (we'd only added it to the
   *Guardener* app). See the table at top.
2. **Granting app access doesn't retro-run checks** → reopen the PR or push a new commit.
3. **Terms TUI `stream timeout`** → re-run `chainctl develop init`, accept promptly.
4. **`run-ci` says `go: not found`** → old session's env predates `ci.yaml`; `apk add go` or
   open a fresh session.
