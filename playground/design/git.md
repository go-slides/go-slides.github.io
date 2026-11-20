# Git in the playgrounds: as built

Status: as built, 2026-11-15. The first version of this document (2026-11-01)
was a design; most of it has shipped, with one large deviation: on
go-slides.github.io the token is **not** in the page, it is in a frame on a
second origin (§3). Push is always to a new branch, followed by a pull request.

Still open:
- **GitHub App** for server mode: not registered yet. `slides-gitd` needs one
  for its `github-app` forge (gitd's README, "Registering the forges"); that is
  with the maintainer.
- **GitLab application** for "Sign in with GitLab" in the frame: not
  registered yet. Redirect URI `https://go-slides-usercontent.github.io/callback.html`
  (README.md:351-357). Until its Pages workflow names an application, sign-in
  is off and only a personal access token works. Nothing here claims the
  sign-in works.
- **go-tex has not moved** to `browsergit` (§1).
- Not built: the `git` version kind, a three-way merge on Refresh, Gitea and
  Forgejo in the playground, the `gitcorsproxy` path (§1, §5).

Every "implemented" below names the file and line, or the release, it was read
from, at origin/main of go-slides/go-slides.github.io on 2026-11-15.

## 1. One git layer for the three playgrounds

**What was built.** `github.com/go-browserhttp/browsergit`, tags v0.1.0,
v0.2.0, v0.3.0, v0.4.0 (`gh release list -R go-browserhttp/browsergit`). Its README
lists the packages:
- `remote`: the `Remote` interface (`Load`, `Propose`, `Refresh`), `Base`,
  path, branch and repository validation, size limits;
- `forge`: REST adapters for GitHub, GitLab and Gitea/Forgejo;
- `smarthttp`: a `Remote` over git smart HTTP (go-git), a package of its own so
  that a REST-only page does not carry go-git (about 13 MB of wasm);
- `oauth`: authorization code with PKCE for a public client;
- `gitd`: a `Remote` over `slides-gitd`;
- `secret`: `Token`, which prints `[redacted]` and refuses every encoder;
- the root package `browsergit` and `gitrpc`, "moved here from go-tex's
  `playground/internal/browsergit`, history kept".

**Deviations from the sketch.**
- The interface differs from the one sketched: `Propose` takes a
  `remote.Proposal` and returns a `remote.Result` (the `browsergit` README's
  example; `playground/gitframe.go:81`). `Remote` is the one interface the
  playground uses, as designed.
- The REST and smart-HTTP implementations are split into `forge` and
  `smarthttp`; there is no `forge.Remote`/`browsergit.Remote` pair.
- `Propose` does not name the branch on a server: the server does (§4).

**go-slides uses it.** `playground/go.mod:8` requires
`github.com/go-browserhttp/browsergit v0.4.0`. The packages imported by the
playground are `remote` (`playground/gitframe.go:9`), `gitd`, `oauth`, `secret`
(`playground/gitvm.go:10-13`) and, in the wasm module, `forge`
(`playground/cmd/playground-wasm/git.go:15-18`).

**go-tex has NOT moved.** On go-tex/go-tex.github.io main,
`playground/internal/browsergit` still exists (its `browsergit.go` and tests,
last commit 2026-11-08) and `playground/go.mod` does not mention
`browsergit`. So the "go-tex moves to the module, and its
`internal/browsergit` is deleted" step of the design is still to do; the
module has the copy with history, the original is not removed.

**In go-slides.**
- *Open from a repository* (Load) reads a folder holding `deck.md`; it is one
  step of Undo and is kept as a "From GitHub" version (README.md:288-289).
  This is not the `git` version kind with a base SHA that the sketch named; the
  base is kept with the view-model and the file hashes (README.md:298-299).
- *Propose changes…* creates a new branch `slides/<title>-<time>` and a pull
  request, never writing to an existing or default branch
  (README.md:290-291; the prefix is `playground/gitvm.go:78`).
- *Check for changes* compares the forge's newer version with the project, and
  *Take theirs* loads it (README.md:292-293). The explorer marks `M`, `A`, `D`
  and `!` where the forge changed the same file differently
  (README.md:294-299). There is **no three-way merge** per file; `!` is a
  conflict that Take theirs would overwrite. The sketch's compare-view merge is
  not built.
- The pane is `toolkit.ForgePane` (`playground/scene_git.go`) over `forgeModel`
  (`playground/forgemodel.go`), README.md:301-304.

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
tokens last 2 h with a refresh token. A pure SPA works.

**Why GitLab needs `api`, and what that costs.** GitLab's scope table
([access token scopes](https://docs.gitlab.com/security/tokens/access_token_scopes/))
says that `write_repository` "Grants read and write access (pull and push) to
repositories" and "Does not support API authentication". That is git over
HTTP only, which GitLab serves without CORS, so a page cannot use it. The REST
calls a proposal makes (the Commits API with `start_sha`, then a merge request)
need `api`, which "Grants complete read and write access to the API for the
token's scope". For a personal token it also grants Git over HTTP.

A stolen `api` token can do anything its user can do through the API, on every
project the user can reach. Examples: delete branches the user may delete,
change settings the user administers, read private projects, and create
further tokens where allowed. Hence:
- prefer a **project or group access token**, limited to the deck
  repositories (for those, `api` does not include the Git-over-HTTP clause
  and is limited to that project or group);
- otherwise use a **dedicated GitLab account** that is a member only of the
  deck projects, with the Developer role;
- keep tokens short-lived: the OAuth sign-in's tokens last 2 h, rotate on
  refresh, and are revoked on Disconnect.

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

**The sketch said:** on a public github.io SPA, yes, the token lives in the
page; keep it in memory, never persist a refresh token, offer "remember on this
device" by explicit choice through the Credential Management API, and rely on
the page's CSP and the sandboxed deck frames. RFC 10017: "malicious JavaScript
code has the same privileges as the legitimate application code". OWASP: one
XSS reads all of IndexedDB.

**What was built instead: the token is not in the page.** The page cannot lose
what it never held, so the git panel runs in a frame on a second origin,
`https://go-slides-usercontent.github.io` (github.io is on the Public Suffix
List, so it is a different site, with storage of its own).

*The frame holds the token.* It takes it from a password input of its own,
keeps it in its own `localStorage` only when "Remember on this device" is
ticked, does the GitLab sign-in and runs the forge calls
(`playground/gitframe.go:12-21`; README.md:306-314). Its own side is the
repository go-slides-usercontent/go-slides-usercontent.github.io
(`playground/gitframe.go:23`); I did not read that repository, only the page's
side and what the page's repository declares of it.

*The page sends typed commands.* `load`, `propose`, `refresh`, `disconnect`,
`status` (README.md:315-316). The page's side is a `remote.Remote` whose every
call is a command to the frame (`frameRemote`,
`playground/gitframe.go:48-122`), over `FrameHost`
(`playground/gitframe.go:36-46`). The wasm side is
`playground/cmd/playground-wasm/gitframe.go`:
- the iframe is created when the pane first asks for it, so a visit that never
  opens the Git pane downloads nothing of the frame (lines 58-59, 83-112);
- it is **not sandboxed**: the frame needs its own origin's storage, and what
  keeps it from the page is that it is another origin (lines 27-28;
  `playground/browsertest/git.cjs:298` asserts `sandbox === null`);
- the page names the origin once (`goSlidesGitFrameOrigin`) and uses it as the
  iframe address, as the `targetOrigin` of every message, never `"*"` (line
  257), and as the only origin it believes, together with the iframe's own
  window as the source (line 149);
- the page says hello to the frame with a random nonce once the frame has
  loaded (lines 97-109), because `location.ancestorOrigins` is missing in
  Firefox; the nonce also travels in every command (line 246). This is the
  deviation from "no frame": it is a handshake, not authentication of the
  frame to the page;
- the answer to a command is data plus files; a command that never answers
  fails after 5 minutes (lines 30-32, 259-266).

*No command takes or returns a token.* A script injected into the page finds
none, in memory, in storage or in a message (`playground/gitframe.go:18-21`);
`browsertest/git.cjs` searches for it after a connect and a Remember
(layouts/_default/playground.html:28-31).

*`propose` needs a click inside the frame.* The frame shows the proposal and
waits for its own button; before that, nothing is written
(`playground/browsertest/git.cjs:331-334`: no `POST` until `fr.click("#allow")`;
README.md:321-324). The page still chooses nothing about the branch: only the
title and body travel, the frame names the branch under `slides/`
(`playground/gitframe.go:86-88`).

*The page's CSP.* `connect-src 'self'` (no forge) and
`frame-src 'self' blob: <frame origin>`
(`layouts/_default/playground.html:10`, the `{{ $fo }}` origin defaulting to
go-slides-usercontent.github.io at line 8; asserted by `TestPlaygroundPageCSP`,
`playground/gitvm_test.go:518-528`). The frame's own policy is declared in
`.second-origin.json:6` and checked in CI by `secondorigin verify`
(README.md:346-348).

*What it does not close* (README.md:321-327, 336-342): a script injected into
the page can still read the frame's nonce (the iframe's `src`) and send the
typed commands, as before (reading the loaded folder; asking for a proposal,
which the person must now confirm in the frame); it can draw a fake pane over
the frame and ask for a token (phishing of the person, which no page can
prevent inside its own window); and, because the preview frames are `srcdoc`
iframes that inherit the page's policy, it can still send what the page holds
as the URL of an image, a style sheet or a media file
(`layouts/_default/playground.html:20-27`). What the page holds no longer
includes a forge token.

*Migration.* A token an earlier version kept in the page's own `localStorage`
is dropped at first load, unused, and the pane says so
(`playground/gitframe.go:127-146`; README.md:328-329). "Forget my drafts"
disconnects the frame.

**Deviations from the sketch.**
- Memory-only storage in the page, and the Credential Management API: replaced
  by the frame's own `localStorage`, behind the "Remember" box.
- "Send the token only to the forge's API": the page sends it nowhere; the
  frame's CSP is `connect-src 'self' https://api.github.com https://gitlab.com`
  (`.second-origin.json:6`).
- Why a second origin: the page's CSP could not close every route out of the
  page. The comment at `layouts/_default/playground.html:34-38` records that
  moving the preview to a document of its own was measured in Chrome and did
  not close the image route, "That is why the token was removed from the page
  instead."

**Self-hosted BFF: no, not in the browser.** As sketched. In server mode the
browser holds only the session cookie (HttpOnly, Secure, SameSite=Strict, the
`__Host-` prefix, README.md:367-368 and gitd's README), the token, Remember and
OAuth controls are gone, and a token an earlier visit remembered is forgotten
(README.md:364-366). On the server, secrets come from files, credentials are
sealed per user, and nothing is logged (§4).

## 4. The BFF: `slides-gitd`

**Built:** [go-slides/gitd](https://github.com/go-slides/gitd). The latest tag
is v0.2.2 (the first note in this document said v0.1.0); the playground's
`playground/go.mod:13` requires v0.2.2. It is a service on 127.0.0.1:8091
behind Caddy and oauth2-proxy (gitd's README, the diagram and "Caddy and
oauth2-proxy").

**API** (gitd's README, "API"):
- `GET /api/git/session`: creates or returns the user's session, sets
  `__Host-gitd`, returns the CSRF token and the repositories the user may use;
- `POST /api/git/credential` and `/api/git/credential/delete`: store or forget
  the user's own token for a `user-token` forge, sealed, never sent back;
- `POST /api/git/load`, `/api/git/propose`, `/api/git/refresh`.

**Deviations from the sketch** (all from gitd's README, "Security"):
- the subject is `X-Auth-Request-User` (oauth2-proxy's `sub`), not the email;
- the cookie is SameSite=Strict and `__Host-`, not Lax; and a CSRF token is
  required on every mutating call, with the exact configured `Origin` and
  `X-Requested-With: slides-gitd`;
- the branch is `slides/<user>-<hash>/<title>-<time>` and a request field
  naming a branch is refused;
- user tokens are sealed per user (AES-256-GCM, HKDF per subject), bound to the
  subject and the forge;
- the GitHub App's installation token is minted for the one repository of the
  call with `contents:write`, `pull_requests:write`, `metadata:read`, as
  sketched. The `Co-authored-by` trailer of the sketch is not in gitd's README,
  and I did not look for it in the code;
- GitLab on the server is a per-user `api` token stored through
  `/api/git/credential` (the README, "Registering the forges"), not a
  per-user OAuth client; the reasons are those of §2.

**The playground side.** `playground/gitdvm.go` (`GitModeClient` and the server
mode, line 41; `initGitd`, line 67) and `scene_gitd.go`. The page finds the
server by a static, public marker, `GET /.well-known/slides-gitd`
(`playground/cmd/playground-wasm/git.go:150-183`;
`selfhost/Caddyfile.gitd:43-50`), not by a call to the API, so a down server
(502) never makes the pane offer a token field. `browsergit/gitd` speaks to
the page's own origin only. A 401 shows **Sign in**; the sign-in ending never
falls back to a token (README.md:359-370;
`playground/gitdvm.go:186,250,424`). The frame is not shown in this mode
(`playground/gitframe.go:154-156`).

**Self-hosting files** (`selfhost/`):
- `Caddyfile.gitd`: `/oauth2/*` to oauth2-proxy on 4180; `/api/git/*` in a
  `route` (identity headers deleted from the client, then `forward_auth` to
  `/oauth2/auth` copying `X-Auth-Request-User` and `-Email`, a 64MB body limit,
  then `reverse_proxy` to 8091 with `X-Gitd-Proxy-Secret` read from a file);
  the marker; the static site with the playground headers (lines 15-58);
- `headers.caddy:23-29`: `playground_headers_gitd`, whose only change from
  the playground's policy is `connect-src 'self'`;
- `oauth2-proxy.cfg`: any OIDC provider, `code_challenge_method = "S256"`,
  secrets as files, cookie `Secure` and `SameSite=Lax`;
- `selfhost/caddytest/caddy_test.go`: runs the Caddyfile through Caddy with a
  fake oauth2-proxy and a fake slides-gitd and checks that a client-sent
  identity never reaches it (lines 196, 282, 317).

`playground/browsertest/gitd.cjs` drives two people in two browser contexts
against the real slides-gitd code and checks that neither sees the other's
repositories, decks, branches, session or tokens (README.md:372-376).
`playground/browsertest/cookie.cjs` asks WebKit, Firefox and Chromium, over
https and across two sites, whether the `__Host-gitd` cookie is stored and sent,
including on the first request after the sign-in redirect (README, Self-hosting).

Bounds, rate limit, forge allowlist and audit are in gitd's README
("Security": limits, per-user token bucket, audit line per call with no
content or token).

## 5. As built, one path per case

| Deployment | Forge | Path | State |
|---|---|---|---|
| go-slides.github.io | GitHub | the git frame on go-slides-usercontent.github.io; a **fine-grained PAT** (Contents and Pull requests read/write) typed in the frame; REST through `browsergit/forge`, no proxy; "Remember on this device" is the frame's, by the box | built (README.md:306-314) |
| go-slides.github.io | GitLab.com | the frame: a personal access token with `api`, or **Sign in with GitLab** (PKCE, public client, the frame's popup) | the token works; **sign-in is pending**: the GitLab application is not registered (README.md:351-357) |
| go-slides.github.io | Gitea/Forgejo/other | not in the playground: the frame offers GitHub and GitLab only (`playground/cmd/playground-wasm/gitframe.go:164`). `browsergit/forge` and `smarthttp` support them; `gitcorsproxy` is not used by go-slides | not built |
| Self-hosted (`selfhost/Caddyfile.gitd`) | GitHub | **`slides-gitd`** with a GitHub App; the browser holds the session cookie only | built and tested against fakes; the **GitHub App is not registered yet**, so no live forge has been exercised |
| Self-hosted | GitLab | `slides-gitd` with the user's own `api` token, sealed | built |

Sources opened:
- GitHub: *Authorizing OAuth apps*; *About security hardening with OpenID
  Connect*; *Refreshing user access tokens*; *Using CORS and JSONP*; *Git
  trees*; *Managing your personal access tokens*.
- GitLab: *OAuth 2.0 identity provider API*; *GitLab as an OpenID Connect
  identity provider*; *access token scopes*.
- Gitea: *OAuth2 provider*.
- Forgejo: *OAuth2 provider*.
- RFC 10017 (draft-ietf-oauth-browser-based-apps-27).
- OWASP: *HTML5 Security Cheat Sheet*.
- git-scm: *http-protocol*.
- For §1, §3-§5, read on 2026-11-15: go-slides/go-slides.github.io at
  origin/main (README.md, `playground/gitframe.go`,
  `playground/cmd/playground-wasm/gitframe.go` and `git.go`,
  `layouts/_default/playground.html`, `.second-origin.json`,
  `playground/browsertest/git.cjs`, `selfhost/*`, `playground/go.mod`);
  go-slides/gitd README and tags; go-browserhttp/browsergit README and tags;
  go-tex/go-tex.github.io `playground/go.mod` and `playground/internal`.
- Not opened: the frame's own repository, go-slides-usercontent; the
  `gitd` and `browsergit` source.
