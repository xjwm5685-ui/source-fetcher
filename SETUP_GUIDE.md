# GitHub Repository Setup Guide

This guide will help you set up your Source Fetcher repository on GitHub with all the optimizations.

## 📋 Pre-Push Checklist

Before pushing to GitHub, ensure you have:

- [x] All documentation files created
- [x] LICENSE file in place
- [x] README optimized with badges
- [x] GitHub Actions workflows configured
- [x] Issue and PR templates created
- [ ] Replace placeholder values (see below)

## 🔧 Required Replacements

### 1. Update README Badges

In both `README.md` and `README.en.md`, replace:

```markdown
YOUR_USERNAME → jiahe
```

Replace with your actual GitHub username (already set to "jiahe").

### 2. Update Security Contact

In `SECURITY.md`:

✅ Already updated to: ckkhua89@gmail.com

### 3. Update Funding (Optional)

In `.github/FUNDING.yml`, uncomment and configure your preferred funding platforms.

### 4. Add Logo (Optional but Recommended)

Replace the placeholder logo URL in README files:

```markdown
![Source Fetcher Logo](https://via.placeholder.com/200x200?text=SF)
```

With your actual logo URL or local path:

```markdown
![Source Fetcher Logo](./assets/logo.png)
```

## 🚀 Initial Repository Setup

### 1. Create Repository on GitHub

```bash
# Initialize git if not already done
git init

# Add all files
git add .

# Create initial commit
git commit -m "feat: initial commit with full documentation and CI/CD"

# Add remote
git remote add origin https://github.com/jiahe/source-fetcher.git

# Push to GitHub
git branch -M main
git push -u origin main
```

### 2. Configure Repository Settings

Go to your repository settings on GitHub:

#### General Settings

1. **Description**: "Unified package download tool - No native clients required"
2. **Website**: Add your documentation site (if any)
3. **Topics**: Add the following topics (see GITHUB_TOPICS.md):
   - package-manager
   - download-manager
   - npm
   - chocolatey
   - winget
   - golang
   - cli
   - tui
   - pip
   - cargo
   - maven
   - dependency-management
   - offline-installer
   - mirror
   - batch-download
   - windows
   - cross-platform

#### Features

- [x] Wikis (optional)
- [x] Issues
- [x] Sponsorships (if using GitHub Sponsors)
- [x] Discussions (recommended)
- [ ] Projects (optional)

#### Pull Requests

- [x] Allow squash merging
- [x] Allow merge commits
- [x] Allow rebase merging
- [x] Automatically delete head branches

#### Security

- [x] Enable Dependabot alerts
- [x] Enable Dependabot security updates
- [x] Enable CodeQL analysis (already configured in workflows)

### 3. Enable GitHub Actions

1. Go to **Actions** tab
2. Enable workflows if prompted
3. Verify that workflows are listed:
   - Tests
   - Release
   - CodeQL

### 4. Configure Branch Protection (Recommended)

For the `main` branch:

1. Go to **Settings** → **Branches** → **Add rule**
2. Branch name pattern: `main`
3. Enable:
   - [x] Require a pull request before merging
   - [x] Require status checks to pass before merging
     - Select: Tests, Lint, Build
   - [x] Require conversation resolution before merging
   - [x] Require linear history (optional)
   - [x] Include administrators (optional)

### 5. Set Up GitHub Discussions

1. Go to **Settings** → **Features**
2. Enable **Discussions**
3. Create categories:
   - 💬 General
   - 💡 Ideas
   - 🙏 Q&A
   - 📣 Announcements
   - 🎉 Show and Tell

### 6. Configure Issue Labels

Add custom labels for better organization:

```
Type:
- bug (red)
- enhancement (blue)
- documentation (green)
- question (purple)
- security (orange)

Priority:
- priority: critical (red)
- priority: high (orange)
- priority: medium (yellow)
- priority: low (green)

Status:
- status: needs-triage (gray)
- status: in-progress (yellow)
- status: blocked (red)
- status: ready-for-review (green)

Area:
- area: npm (blue)
- area: choco (brown)
- area: winget (cyan)
- area: tui (purple)
- area: cli (green)
```

## 📦 First Release

### 1. Create a Git Tag

```bash
# Create and push a version tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

### 2. GitHub Actions Will Automatically

- Run tests
- Build binaries for all platforms
- Create a GitHub Release
- Upload binaries and checksums

### 3. Edit Release Notes

After the release is created:

1. Go to **Releases**
2. Edit the latest release
3. Add detailed release notes from CHANGELOG.md
4. Mark as "Latest release"

## 🎨 Optional Enhancements

### 1. Create a Logo

Design a logo for your project:
- Size: 200x200px minimum
- Format: PNG with transparency
- Place in: `assets/logo.png`
- Update README files

### 2. Record Demo GIF

Use tools like:
- **Windows**: ScreenToGif, LICEcap
- **macOS**: Kap, Gifox
- **Linux**: Peek, SimpleScreenRecorder

Record:
- TUI interface navigation
- Search and download process
- Installation workflow

Add to README:
```markdown
## 🎬 Demo

![Demo](./assets/demo.gif)
```

### 3. Create Documentation Site

Consider using:
- **GitHub Pages** - Free hosting
- **GitBook** - Beautiful documentation
- **Docusaurus** - React-based docs
- **MkDocs** - Python-based docs

### 4. Set Up Social Media

Create accounts for:
- **Twitter/X** - Announcements and updates
- **Discord** - Community chat
- **Reddit** - r/golang, r/programming
- **Dev.to** - Technical articles

## 📊 Analytics and Monitoring

### GitHub Insights

Monitor:
- **Traffic** - Views and clones
- **Community** - Issues, PRs, discussions
- **Commits** - Activity over time
- **Dependency graph** - Dependencies and dependents

### External Tools

Consider integrating:
- **Codecov** - Code coverage reporting
- **Snyk** - Security vulnerability scanning
- **Better Uptime** - Monitor download availability
- **Google Analytics** - Documentation site analytics

## 🎯 Post-Launch Checklist

### Week 1
- [ ] Monitor GitHub Actions for failures
- [ ] Respond to initial issues
- [ ] Share on social media
- [ ] Post on Reddit (r/golang, r/programming)
- [ ] Submit to Hacker News

### Week 2
- [ ] Write introductory blog post
- [ ] Create video tutorial
- [ ] Engage with community feedback
- [ ] Fix critical bugs

### Month 1
- [ ] Analyze usage patterns
- [ ] Plan next features based on feedback
- [ ] Update documentation based on questions
- [ ] Celebrate first 100 stars! 🎉

## 🆘 Troubleshooting

### GitHub Actions Failing

1. Check workflow logs
2. Verify Go version compatibility
3. Ensure all tests pass locally
4. Check for missing secrets/tokens

### Badges Not Showing

1. Verify repository is public
2. Check badge URLs are correct
3. Wait a few minutes for cache refresh
4. Try hard refresh (Ctrl+F5)

### Release Build Failing

1. Verify tag format (v1.0.0)
2. Check build script syntax
3. Ensure all platforms are supported
4. Review release workflow logs

## 📚 Additional Resources

- [GitHub Docs](https://docs.github.com/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)

## 🎉 You're Ready!

Your repository is now fully optimized for success. Good luck with your project!

---

**Questions?** Open an issue or discussion on GitHub.
