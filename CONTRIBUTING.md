# Contributing to PrivUtil

First off, thanks for taking the time to contribute! ❤️

All types of contributions are encouraged and valued. See the [Table of Contents](#table-of-contents) for different ways to help and details about how this project handles them. Please make sure to read the relevant section before making your contribution. It will make it a lot easier for us maintainers and smooth out the experience for all involved. The community looks forward to your contributions.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [I Have a Question](#i-have-a-question)
- [I Want To Contribute](#i-want-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Enhancements](#suggesting-enhancements)
  - [Your First Code Contribution](#your-first-code-contribution)
  - [Improving The Documentation](#improving-the-documentation)
- [Styleguides](#styleguides)
  - [Commit Messages](#commit-messages)
- [Join The Project Team](#join-the-project-team)

## Code of Conduct

This project and everyone participating in it is governed by the
[PrivUtil Code of Conduct](CODE_OF_CONDUCT.md).
By participating, you are expected to uphold this code. Please report unacceptable behavior
to our team.

## I Have a Question

If you want to ask a question, we assume that you have read the available [Documentation](link-to-docs).

Before you ask a question, it is best to search for existing [Issues](https://github.com/odinnordico/privutil/issues) that might help you. In case you have found a suitable issue and still need clarification, you can write your question in this issue. It is also advisable to search the internet for answers first.

If you then still feel the need to ask a question and need clarification, we recommend the following:

- Open an [Issue](https://github.com/odinnordico/privutil/issues/new).
- Provide as much context as you can about what you're running into.
- Provide project and platform versions (nodejs, npm, etc), depending on what seems relevant.

We will then take care of the issue as soon as possible.

## I Want To Contribute

### Reporting Bugs

**If you find a security vulnerability, please send an email to security@example.com instead of using the issue tracker.**

Before you submit a bug report, please check these points:

- Make sure that you are using the latest version.
- Determine if your bug is really a bug and not an error on your side e.g. using incompatible environment components/versions (Make sure that you have read the [documentation](link-to-docs). If you are looking for support, you might want to check [this section](#i-have-a-question)).
- To see if other users have experienced (and potentially already solved) the same issue you are having, check if there is not already a bug report existing for your bug or error in the [bug tracker](https://github.com/odinnordico/privutil/issues?q=label%3Abug).

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for PrivUtil, **including completely new features and minor improvements to existing functionality**. Following these guidelines will help maintainers and the community to understand your suggestion and find related suggestions.

- Use a **clear and descriptive title** for the issue to identify the suggestion.
- Provide a **step-by-step description of the suggested enhancement** in as many details as possible.
- **Describe the current behavior** and **explain which behavior you expected to see instead** and why. At this point you can also tell which alternatives do not work for you.
- You may want to include **screenshots and animated GIFs** which help you demonstrate the steps or point out the part which the suggestion is related to.
- **Explain why this enhancement would be useful** to most PrivUtil users. You may also want to point out the other projects that solved it better and which could serve as inspiration.

### Your First Code Contribution

Unsure where to begin contributing to PrivUtil? You can start by looking through these `good-first-issue` and `help-wanted` issues:

- [Good first issues](https://github.com/odinnordico/privutil/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22) - issues which should only require a few lines of code, and a test or two.
- [Help wanted issues](https://github.com/odinnordico/privutil/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22) - issues which should be a bit more involved than `good first issue`s.

### Improving The Documentation

Documentation improvements are always welcome.

## Styleguides

### Commit Messages

- Use [Conventional Commits](https://www.conventionalcommits.org/).
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `style:` for formatting changes
- `refactor:` for code refactoring
- `test:` for adding/updating tests
- `chore:` for tooling/build updates

## Development Workflow

### Building

```bash
make build      # Build everything
make run        # Build and run locally
```

### Testing

```bash
make test           # Run all tests
make test-backend   # Go tests only
make test-frontend  # React tests only
```

### Linting

Before submitting a PR, ensure your code passes all linters:

```bash
make lint           # Run all linters
make lint-backend   # Go: go vet, go fmt
make lint-frontend  # ESLint
```

## Join The Project Team

If you are interested in joining the team, please reach out to us!
