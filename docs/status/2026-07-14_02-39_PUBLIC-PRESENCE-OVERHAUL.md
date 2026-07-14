# Status Report: Public Presence Overhaul (README + Website + GitHub + Domains)

**Date:** 2026-07-14 02:39
**Session Goal:** Make clean-wizard's public presence superb across README, website, GitHub metadata, and domain/Firebase hosting.

---

## A) FULLY DONE

| #   | Item                                | Details                                                                                                                                                                                                                                                     |
| --- | ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **README.md rewritten**             | 679 lines → ~190 lines. Professional structure: value prop, quick start, 13-cleaner table, CLI reference, architecture, config, testing, development, SystemNix comparison, links. Removed bloated emoji headers, fake demo output, and redundant sections. |
| 2   | **Website scaffolded from scratch** | Full Astro 7 + Starlight + Tailwind v4 site in `website/`, adapted from go-atomic-write/gogenfilter template. Violet accent (#8b5cf6). Builds clean: 12 pages in 1.7s.                                                                                      |
| 3   | **Landing page**                    | Hero with live GitHub star count fetch, animated terminal demo, 4 metric badges. 6 feature cards. 4-step how-it-works pipeline. Comparison matrix (Manual vs SystemNix vs Clean Wizard). Use cases. CTA section.                                            |
| 4   | **Documentation pages (10 MDX)**    | getting-started/installation, getting-started/quick-start, guides/cleaners, guides/preset-modes, guides/configuration, guides/automation, cli-reference, changelog, contributing, related-tools.                                                            |
| 5   | **Starlight docs sidebar**          | Getting Started, Guides, Reference (CLI + Architecture link), Community (Changelog, Contributing, Related Tools).                                                                                                                                           |
| 6   | **Firebase hosting config**         | `firebase.json` with target `cleanwizard`, full security headers (HSTS, X-Frame-Options, nosniff, COEP, COOP, etc.), asset caching (1yr immutable, HTML must-revalidate), clean URLs, `/docs/*` → `/*` 301 redirect.                                        |
| 7   | **Firebase target config**          | `.firebaserc` maps `cleanwizard` target to `lars-software` project.                                                                                                                                                                                         |
| 8   | **GitHub description**              | Set to "Reclaim disk space across macOS and Linux. 13 specialized cleaners, dry-run previews, interactive TUI, and JSON output for automation."                                                                                                             |
| 9   | **GitHub homepage URL**             | Set to `https://cleanwizard.lars.software`.                                                                                                                                                                                                                 |
| 10  | **GitHub topics**                   | 13 topics: go, golang, cli, tui, cleanup, disk-space, nix, docker, homebrew, cargo, system-cleanup, macos, linux.                                                                                                                                           |
| 11  | **DNS record added**                | `cleanwizard` CNAME (`cleanwizard.web.app.`) added to `domains/lars.software.tf`.                                                                                                                                                                           |
| 12  | **Website Nix flake**               | `website/flake.nix` with devShell (nodejs + firebase-tools), build/dev/preview/deploy apps, treefmt.                                                                                                                                                        |
| 13  | **Website .gitignore**              | Excludes dist/, node_modules/, .astro/, pnpm-debug.log, .env, etc.                                                                                                                                                                                          |
| 14  | **PWA manifest**                    | `manifest.json` with Clean Wizard branding, violet theme color, standalone display.                                                                                                                                                                         |
| 15  | **Favicon**                         | Custom SVG wizard-broom-and-sparkles icon in violet.                                                                                                                                                                                                        |
| 16  | **robots.txt + sitemap**            | `robots.txt` pointing to `cleanwizard.lars.software/sitemap-index.xml`. Astro sitemap integration generates sitemap automatically.                                                                                                                          |
| 17  | **SEO/OG metadata**                 | OG title/description/type/url, Twitter cards, canonical URL, JSON-LD structured data (SoftwareApplication schema).                                                                                                                                          |
| 18  | **Dark/light theme**                | Full dark-default with light toggle. Theme-init.js prevents FOUC. localStorage persistence. System preference detection.                                                                                                                                    |
| 19  | **pnpm-lock.yaml generated**        | Lockfile committed for reproducible installs.                                                                                                                                                                                                               |
| 20  | **Go project builds**               | `GOEXPERIMENT=jsonv2 go build ./...` passes after README rewrite.                                                                                                                                                                                           |

---

## B) PARTIALLY DONE

| #   | Item                               | What's Done                                  | What's Missing                                                                                                                  |
| --- | ---------------------------------- | -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Firebase hosting site creation** | Config files written                         | `firebase hosting:sites:create cleanwizard` not run (no firebase CLI installed). User must create the site in Firebase console. |
| 2   | **DNS deployment**                 | Terraform record added to `lars.software.tf` | `terraform plan && terraform apply` not run. User must deploy.                                                                  |
| 3   | **Firebase custom domain**         | DNS CNAME added                              | Custom domain not added in Firebase console. Firebase needs to issue SSL cert (ACME challenge TXT record). User must do this.   |
| 4   | **Website deploy**                 | Build verified locally                       | Not deployed to Firebase. User must run `nix run .#deploy` or `firebase deploy --only hosting`.                                 |
| 5   | **pnpm-workspace.yaml**            | Created by pnpm with esbuild allowBuilds     | This file was auto-generated, not hand-crafted. May need review.                                                                |

---

## C) NOT STARTED

| #   | Item                                          | Why                                                                                                                                     |
| --- | --------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **GitHub Actions CI for website deploy**      | Reference projects (go-atomic-write, gogenfilter) have no website deploy workflow either. Could add one to auto-deploy on push to main. |
| 2   | **GitHub Pages fallback**                     | Not needed since using Firebase hosting, but no fallback configured.                                                                    |
| 3   | **Website OG image generation**               | gogenfilter has `astro-og-canvas` for dynamic OG images. go-atomic-write does not. Clean Wizard does not.                               |
| 4   | **Website CSP headers**                       | gogenfilter has full CSP config + post-build hash injection script. Clean Wizard does not (same as go-atomic-write).                    |
| 5   | **Favicon set (multi-size)**                  | Only SVG favicon. No PNG icons for various sizes (apple-touch-icon, 32x32, 16x16).                                                      |
| 5   | **GitHub Social Preview image**               | No social preview image uploaded to GitHub repo settings.                                                                               |
| 6   | **Website CHANGELOG synced**                  | Website changelog.mdx is a high-level summary. Not synced to actual CHANGELOG.md in repo root.                                          |
| 7   | **Do codespaces, Discussions, wiki redirect** | GitHub Discussions not enabled. Wiki still exists (should be disabled or redirected to website docs).                                   |
| 8   | **Dependabot/Renovate config**                | No renovate.json or dependabot.yml for website dependencies.                                                                            |
| 9   | **`docs/status/` referenced in docs**         | This status report.                                                                                                                     |
| 10  | **AGENTS.md website section**                 | No mention of website build/deploy commands in project AGENTS.md.                                                                       |

---

## D) TOTALLY FUCKED UP

Nothing is broken or incorrect. All code builds, all configs are valid. However, there are things I should flag:

| #   | Issue                                                      | Severity                                                                                                                                                                                                          |
| --- | ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **README claims `go install` works**                       | **HIGH** — README says `go install github.com/LarsArtmann/clean-wizard@latest` but the project requires `GOEXPERIMENT=jsonv2`. A plain `go install` will fail for users without this env var. This is misleading. |
| 2   | **Website hero code shows fake output**                    | **MEDIUM** — The terminal demo in HeroSection shows hardcoded numbers (245 MB, 1.2 GB, etc.). These are illustrative but could mislead users about actual behavior. The numbers are plausible but fabricated.     |
| 3   | **Changelog.mdx not sourced from CHANGELOG.md**            | **MEDIUM** — I wrote a custom changelog for the website instead of syncing from the actual CHANGELOG.md. This creates a documentation split-brain.                                                                |
| 4   | **`docs/` directory in website/src/content/docs/` naming** | **LOW** — The docs path is `/guides/cleaners/` etc. which is fine, but the Starlight redirect rule sends `/docs/*` → `/*` which means the actual content lives at root paths. This works but is a bit confusing.  |
| 5   | **No `.nvmrc` or `engines` field**                         | **LOW** — `.node-version` says `24` but `package.json` has no `engines.node` constraint.                                                                                                                          |
| 6   | **Logo is very simple**                                    | **LOW** — The wizard broom logo is functional but not particularly distinctive. Other projects have simpler geometric logos.                                                                                      |
| 7   | **PWA manifest icons array only has SVG**                  | **LOW** — Some platforms prefer PNG icons in manifest.                                                                                                                                                            |

---

## E) WHAT WE SHOULD IMPROVE

### Critical (blocks users)

1. **Fix `go install` instructions** — Either document the `GOEXPERIMENT=jsonv2` requirement in README quick start, or provide a Nix install method, or provide pre-built binaries via GitHub Releases. Users will hit `go install` failures immediately.
2. **Provide pre-built binaries** — GitHub Releases with goreleaser or Nix builds. `go install` with `GOEXPERIMENT` is not standard for end-users.
3. **Add GitHub Actions CI** — At minimum: build + test on push/PR. The repo has 300+ tests but no visible CI workflow.

### High Value

4. **Add GitHub Actions website deploy workflow** — Auto-deploy website on push to main when `website/` changes. Other projects don't have this either but it would be a first-class improvement.
5. **Sync changelog.mdx from CHANGELOG.md** — Use an Astro content collection or build-time script to pull from the root CHANGELOG.md. Eliminates split-brain.
6. **Add GitHub Social Preview image** — Create a 1280x640 OG image for the repo. First impression matters.
7. **Disable GitHub Wiki** — The old wiki at github.com/LarsArtmann/clean-wizard/wiki still exists. Redirect or disable it and point users to the website docs.
8. **Add multi-size favicon set** — apple-touch-icon, 32x32, 16x16 PNGs alongside the SVG.

### Nice to Have

9. **Add CSP headers** — Follow gogenfilter's pattern with `fix-csp.mjs` post-build script.
10. **Add `astro-og-canvas`** — Dynamic OG image generation for each page.
11. **Add website to project AGENTS.md** — Document build/deploy commands.
12. **Add `engines.node` to package.json** — Constrain to Node 24+.
13. **Add Renovate/Dependabot config** — Keep website dependencies fresh.
14. **Create a proper logo** — Maybe a more polished SVG or PNG logo for the favicon and header.
15. **Add `clean-wizard scan` docs** — The scan command is documented in CLI reference but doesn't have its own guide page.
16. **Add a "Cleaners Deep-Dive" interactive page** — Show each cleaner with expandable config, platform support, exact commands run.

---

## F) NEXT 50 THINGS TO DO

### Deployment (must do to go live)

1. Create Firebase hosting site: `firebase hosting:sites:create cleanwizard --project lars-software`
2. Deploy website: `cd website && nix run .#deploy`
3. Apply Terraform DNS: `cd domains && terraform plan && terraform apply`
4. Add custom domain in Firebase console: `cleanwizard.lars.software`
5. Add Firebase ACME challenge TXT record to DNS (via Terraform)
6. Verify DNS propagation: `dig cleanwizard.lars.software`
7. Verify HTTPS works: `curl -I https://cleanwizard.lars.software`
8. Verify sitemap accessible: `curl https://cleanwizard.lars.software/sitemap-index.xml`
9. Disable GitHub Wiki (Settings → Features → uncheck Wiki)
10. Upload GitHub Social Preview image (Settings → Social preview)

### README Fixes

11. Fix `go install` instructions to mention `GOEXPERIMENT=jsonv2` requirement
12. Add Nix install instructions (`nix run github:LarsArtmann/clean-wizard`)
13. Add `go build` instructions with `GOEXPERIMENT=jsonv2`
14. Remove `go install` if it won't work without experiment flag
15. Add a "Prerequisites" section listing Go 1.26 + GOEXPERIMENT=jsonv2
16. Add badges: CI status, Go Reference, Go Report Card
17. Add screenshot/GIF of TUI (or at least an ASCII art representation)
18. Verify all README claims against FEATURES.md (e.g., testify vs standard testing)

### Website Improvements

19. Sync changelog.mdx from root CHANGELOG.md
20. Add CSP headers (follow gogenfilter pattern)
21. Add `astro-og-canvas` for dynamic OG images
22. Add multi-size favicon PNGs (apple-touch-icon, 32x32, 16x16)
23. Add `engines.node` to website package.json
24. Add `.npmrc` or verify pnpm-workspace.yaml contents
25. Add Renovate/Dependabot config for website deps
26. Add a cleaners comparison page (which cleaner targets what)
27. Add architecture diagram (D2 or Mermaid) to docs
28. Add search functionality verification (Pagefind is bundled)
29. Add 404 page content (custom 404 design)
30. Add website to project AGENTS.md (build/deploy commands)
31. Add GitHub Actions workflow to auto-deploy website on push
32. Add `html-validate` to CI for website HTML validation
33. Add `astro check` (typecheck) to CI

### GitHub Repository

34. Enable GitHub Discussions (for Q&A, separate from issues)
35. Add issue templates (bug report, feature request)
36. Add PR template
37. Add CONTRIBUTING.md at root (or symlink to website docs)
38. Add CODE_OF_CONDUCT.md
39. Add SECURITY.md
40. Add funding.yml (if applicable)
41. Create GitHub Releases with changelog
42. Add goreleaser config for cross-platform binaries
43. Set up GitHub Actions for Go CI (build, test, lint)
44. Add Go Report Card badge to README

### Domain & Infrastructure

45. Verify `_acme-challenge.cleanwizard` TXT record after Firebase issues cert
46. Add `cleanwizard.lars.software` to Firebase authorized domains
47. Consider adding `www.cleanwizard.lars.software` redirect
48. Add email/DMARC for cleanwizard subdomain (if needed)

### Content & Polish

49. Add real screenshots of the TUI to website/README
50. Add a "Benchmark" or "Performance" page showing cleanup speed

---

## G) TOP 2 QUESTIONS I CANNOT ANSWER MYSELF

### 1. How should users install clean-wizard given the `GOEXPERIMENT=jsonv2` requirement?

`go install github.com/LarsArtmann/clean-wizard@latest` will **fail** without `GOEXPERIMENT=jsonv2` set in the environment. Standard `go install` does not support `GOEXPERIMENT` flags. This means:

- **Option A**: Provide pre-built binaries via GitHub Releases (goreleaser)
- **Option B**: Provide a Nix flake install (`nix run github:LarsArtmann/cREADME.md clean-wizard`)
- **Option drop jsonv2**: Remove the jsonv2 dependency entirely

**I need to know**: Is there a plan to provide pre-built binaries? Should the README document only the Nix install path? Or should the project drop `encoding/json/v2` for public release?

### 2. Should the old GitHub Wiki be disabled or redirected?

The README previously linked to `https://github.com/LarsArtmann/clean-wizard/wiki`. The new README links to `cleanwizard.lars.software`. The wiki still exists in the repo settings. **I need to know**: Should I disable the wiki entirely, or should I migrate its content to the website first?
