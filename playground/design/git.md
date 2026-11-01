# Git in the playgrounds: design

Status: a design for the user to decide on. Nothing here is implemented.
Push is always to a new branch, followed by a pull request.

## 1. One git layer for the three playgrounds

**What exists.** go-tex's playground already has an in-browser git client,
`playground/internal/browsergit`, built on go-git v5 in wasm with memory
storage, memfs and the HTTP transport over Fetch. It does Clone, Open,
Snapshot, Status, Stage, Commit, Pull, Push and Log, with sideband progress,
fault-injection tests, and one entry point (`SplitCredential`) that keeps a
token out of `.git/config`.

Around it:
- `gitrpc` with `cmd/git-worker`: git runs in a lazily loaded worker module;
- `go-browserhttp/gitcorsproxy` v0.2.0: deployed and hardened (github.com
  allowlist, rate limit, response cap, token never logged);
- loom's `api_git.go`: the server-side operation set that browsergit mirrors.

go-slides has nothing yet. A third copy is not wanted.

**Where it goes.** A new repository, `github.com/go-browserhttp/browsergit`.
The org already holds `browserhttp` and `gitcorsproxy`, so this is the lowest
layer that the go-tex, go-slides and go-notebook playgrounds can all import.

- `browsergit`: moved from go-tex unchanged. Smart HTTP, against a forge's
  own CORS or `gitcorsproxy`.
- `gitrpc` and `worker`: moved from go-tex. One lazily loaded
  `git-worker.wasm` per page.
- `forge`: new. REST adapters, which need no proxy:
  - `github`: git data API (blobs → trees with `base_tree` → commits → refs)
    plus `POST /pulls`;
  - `gitlab`: commits API plus merge requests;
  - `gitea`/`forgejo`: contents API plus pulls.
- `remote`: the one interface the playgrounds use:

```go
type Remote interface {
    Load(ctx, repo, ref, dir string) (files map[string][]byte, base Base, err error)
    Propose(ctx, base Base, files map[string][]byte, branch, title, body string) (prURL string, err error)
    Refresh(ctx, base Base) (theirs map[string][]byte, head Base, err error)
}
```

Two implementations satisfy it, `forge.Remote` (REST) and
`browsergit.Remote` (clone, commit, push to `slides/<name>`). go-tex moves to
the module, and its `internal/browsergit` is deleted.

**In go-slides.**
- *Open from a repository* is a whole-project step. It is undoable, and the
  project as it was is kept first as a version.
- The loaded commit becomes a version of a new kind, `git`, holding the base
  SHA.
- *Propose* commits what differs from that base and opens the PR.
- *Refresh* runs a three-way merge per file over `textdiff` (base, ours,
  theirs). Clean hunks apply; conflicting ones open in the compare view.

## 2. "Can't we authenticate with OIDC?"

OIDC says who the person is: it yields an ID token. Pushing needs the
forge's own access token, so what the forge offers is what matters.

**GitHub** is not an OIDC provider for applications. Its OIDC provider
issues tokens to Actions workflow jobs ("Each job requests an OIDC token").
Sign-in for apps is OAuth 2.0:
- the web flow accepts PKCE (S256), but `client_secret` is still required at
  `login/oauth/access_token`;
- "CORS pre-flight requests are not supported", so a pure SPA cannot finish
  the flow;
- the device flow needs no secret, but CORS on `github.com/login/*` is not
  documented, so it would need our proxy;
- GitHub App user tokens expire after 8 h, with a refresh token that lasts
  6 months.

**GitLab** is an OIDC provider (`openid` scope, discovery URL). Its OAuth
supports PKCE for SPAs with no secret, `/oauth/token` answers CORS preflight,
tokens last 2 h with a refresh token, and `write_repository` covers git over
HTTPS. A pure SPA works.

**Gitea** and **Forgejo** are OAuth2 and OIDC providers with PKCE and public
clients.
- Gitea has a `write:repository` scope.
- Forgejo: "Scopes are not implemented for OAuth2 tokens", so a token has
  full access.
- CORS on their token endpoints is not documented, and must be checked per
  instance.

**The self-hosted case:** the Caddy + oauth2-proxy guide already
authenticates people with OIDC. Add a backend-for-frontend (RFC 10017 §6.1:
"tokens are only available to the BFF").

## 3. "With OIDC, does token storage still matter?"

**Public github.io SPA: yes.** Whatever the flow (a PAT, or PKCE with
GitLab), the access token lives in the page. RFC 10017: "malicious
JavaScript code has the same privileges as the legitimate application
code". OWASP: one XSS reads all of IndexedDB.

What follows:
- keep the token in memory by default;
- prefer short lifetimes (GitLab 2 h);
- never persist a refresh token;
- offer "remember on this device" by explicit choice only, through the
  Credential Management API (as go-tex does);
- send the token only to the forge's API, or to our own proxy.

Already in our favour: the page's CSP (scripts only from itself), and the
deck rendered in iframes sandboxed without the page's origin, so a deck
cannot reach the token.

**Self-hosted BFF: no, not in the browser.** The browser holds only
oauth2-proxy's HttpOnly, Secure, SameSite=Lax session cookie, so the "store
the token?" question disappears. Storage still matters on the server: secrets
come from files placed by the secret store (the guide's rule), tokens are
encrypted at rest, and nothing is logged.

## 4. The BFF: `slides-gitd`

A small Go service, cobra CLI, on 127.0.0.1 behind Caddy, behind the same
`forward_auth`. It reads the user from `X-Auth-Request-Email`, which only
Caddy can set (`trusted_proxy_ips`).

**API:**
- `POST /api/git/load`
- `POST /api/git/propose`
- `POST /api/git/refresh`

The same `remote.Remote` runs server-side, so smart HTTP needs no CORS proxy
there.

**Credentials:**
- **GitHub:** a GitHub App installed on the chosen repositories, with
  permissions `contents:write`, `pull_requests:write` and `metadata:read`.
  The server signs a JWT with the App key (a file) and gets 1 h installation
  tokens. Commits carry `Co-authored-by: <the signed-in user>`.
  Alternatively, per-user App user tokens through the server-side web flow,
  where the client secret stays on the server.
- **GitLab, Gitea, Forgejo:** a per-user confidential OAuth client (the
  server-side code flow), or a service token per repository.

**Threat model:**
- **CSRF:**
  - SameSite=Lax cookie;
  - POST with `Content-Type: application/json` only;
  - a required `X-Requested-With` header (custom headers force a preflight);
  - `Origin` checked against the site.
- **Per-user authorisation:**
  - deny by default;
  - a configured map of user or group to repositories;
  - the server writes only branches `slides/<user>/<name>`;
  - never the default branch or a protected one;
  - one pull request per branch.
- **Scopes:** the App's minimal permissions on selected repositories; the
  narrowest OAuth scopes (`write_repository` on GitLab, `write:repository`
  on Gitea; Forgejo has none, so prefer a service token there).
- **Bounds:**
  - the project limits on what is loaded (files, sizes, paths cleaned);
  - a rate limit per user;
  - forges allowlisted (SSRF).
- **Audit:** an append-only log of user, forge, repository, base and new
  SHA, branch, PR URL and time, with no content and no token.

## 5. Recommendation, one path per case

| Deployment | Forge | Path |
|---|---|---|
| go-slides.github.io | GitHub | REST (`forge/github`) with a **fine-grained PAT**: one repository, Contents and Pull requests read/write, with an expiry. Kept in memory; remember only by explicit opt-in. No proxy. |
| go-slides.github.io | GitLab.com | REST with **OAuth PKCE**: public client, 2 h token, in memory. |
| go-slides.github.io | Gitea/Forgejo/other | **PAT plus smart HTTP**, through the forge's CORS or our `gitcorsproxy`; in memory. |
| Self-hosted (Caddy guide) | any | **BFF `slides-gitd`**: GitHub App installation tokens, server-side OAuth for the others. The browser holds the session cookie only. |

Sources opened:
- GitHub: *Authorizing OAuth apps*; *About security hardening with OpenID
  Connect*; *Refreshing user access tokens*; *Using CORS and JSONP*; *Git
  trees*; *Managing your personal access tokens*.
- GitLab: *OAuth 2.0 identity provider API*; *GitLab as an OpenID Connect
  identity provider*.
- Gitea: *OAuth2 provider*.
- Forgejo: *OAuth2 provider*.
- RFC 10017 (draft-ietf-oauth-browser-based-apps-27).
- OWASP: *HTML5 Security Cheat Sheet*.
- git-scm: *http-protocol*.
- go-tex's `browsergit.go`.
- Memories: browser-git-playground, gotex-playground-ui-architecture.
