# Contributing to Source Fetcher

Thank you for your interest in contributing to Source Fetcher! 🎉

## 🌟 Ways to Contribute

- 🐛 Report bugs
- 💡 Suggest new features
- 📝 Improve documentation
- 🔧 Submit bug fixes
- ✨ Add new features
- 🌍 Translate documentation

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Windows (primary platform)

### Setup Development Environment

1. **Fork the repository**
   ```bash
   # Click "Fork" button on GitHub
   ```

2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/source-fetcher.git
   cd source-fetcher
   ```

3. **Install dependencies**
   ```bash
   go mod download
   go mod vendor
   ```

4. **Run tests**
   ```bash
   go test -v
   ```

5. **Build the project**
   ```bash
   go build
   ```

## 📋 Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test improvements

### 2. Make Your Changes

- Write clean, readable code
- Follow Go best practices
- Add tests for new features
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
go test -v

# Run specific tests
go test -v -run TestName

# Check test coverage
go test -cover
```

### 4. Commit Your Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "feat: add support for pip package search"
```

Commit message format:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Test changes
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## 📝 Code Style Guidelines

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` to format code
- Use meaningful variable names
- Add comments for exported functions
- Keep functions small and focused

### Example

```go
// Good
func downloadPackage(ctx context.Context, name string, version string) error {
    // Implementation
}

// Bad
func dp(c context.Context, n string, v string) error {
    // Implementation
}
```

### Testing

- Write tests for new features
- Maintain test coverage above 80%
- Use table-driven tests when appropriate
- Mock external dependencies

```go
func TestDownloadPackage(t *testing.T) {
    tests := []struct {
        name    string
        pkg     string
        version string
        wantErr bool
    }{
        {"valid package", "react", "18.0.0", false},
        {"invalid package", "", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := downloadPackage(context.Background(), tt.pkg, tt.version)
            if (err != nil) != tt.wantErr {
                t.Errorf("downloadPackage() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 📚 Documentation Guidelines

### README Updates

- Keep examples up to date
- Use clear, concise language
- Include both English and Chinese versions
- Add screenshots or GIFs for visual features

### Code Comments

```go
// DownloadPackage downloads a package from the specified source.
// It supports npm, choco, winget, and other package sources.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - source: Package source (npm, choco, winget, etc.)
//   - name: Package name
//   - version: Package version (use "latest" for latest version)
//
// Returns:
//   - error: nil on success, error on failure
func DownloadPackage(ctx context.Context, source, name, version string) error {
    // Implementation
}
```

## 🐛 Reporting Bugs

### Before Reporting

1. Check if the bug has already been reported
2. Try to reproduce the bug
3. Collect relevant information

### Bug Report Template

```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Run command '...'
2. See error

**Expected behavior**
What you expected to happen.

**Actual behavior**
What actually happened.

**Environment**
- OS: [e.g., Windows 11]
- Go version: [e.g., 1.21]
- Source Fetcher version: [e.g., 1.0.0]

**Additional context**
Any other context about the problem.
```

## 💡 Suggesting Features

### Feature Request Template

```markdown
**Is your feature request related to a problem?**
A clear description of the problem.

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
Other solutions you've thought about.

**Additional context**
Any other context or screenshots.
```

## 🔍 Pull Request Guidelines

### Before Submitting

- [ ] Tests pass locally
- [ ] Code follows style guidelines
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] Branch is up to date with main

### PR Description Template

```markdown
**What does this PR do?**
Brief description of changes.

**Related Issue**
Fixes #123

**Type of Change**
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Refactoring
- [ ] Other (please describe)

**Testing**
How has this been tested?

**Screenshots** (if applicable)
Add screenshots to help explain your changes.

**Checklist**
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Code follows style guidelines
- [ ] All tests pass
```

## 🎯 Priority Areas

We especially welcome contributions in these areas:

1. **New Package Sources** - Add support for more package managers
2. **Performance Optimization** - Improve download speed and efficiency
3. **Cross-Platform Support** - Better Linux and macOS support
4. **Documentation** - Tutorials, examples, translations
5. **Testing** - Increase test coverage
6. **UI/UX** - Improve TUI interface

## 🤝 Code Review Process

1. **Automated Checks** - CI/CD runs tests automatically
2. **Maintainer Review** - A maintainer will review your PR
3. **Feedback** - Address any requested changes
4. **Approval** - Once approved, your PR will be merged
5. **Release** - Changes will be included in the next release

## 📞 Getting Help

- **GitHub Issues** - For bugs and feature requests
- **GitHub Discussions** - For questions and discussions
- **Documentation** - Check README and guides first

## 🙏 Recognition

Contributors will be:
- Listed in CHANGELOG.md
- Mentioned in release notes
- Added to contributors list

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Source Fetcher! 🚀
