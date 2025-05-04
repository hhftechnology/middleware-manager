# Contributing to the Middleware Manager

Want to contribute? Great! First, read this page (including the small print at
the end). By submitting a pull request, you represent that you have the right to
license your contribution to the community, and agree by submitting the patch that 
your contributions are licensed under the [MIT license](./LICENSE). Before 
submitting the pull request, please make sure you have tested your changes and that 
they follow the project guidelines for contributing code.

# Contributing as an Open Source Contributor

As an open source contributor you can report bugs and request features in the
[Issue Tracker](https://github.com/hhftechnology/middleware-manager/issues), as well
as contribute bug fixes and features as a pull request. The requirements to become an 
open source contributor of the [Middleware Manager](https://github.com/hhftechnology/middleware-manager) 
are:

-   Agree to the [License](./LICENSE)


# Bugs

If you find a bug in the source code, you can help us by
[submitting a GitHub Issue](https://github.com/hhftechnology/middleware-manager/issues/new).
The best bug reports provide a detailed description of the issue and
step-by-step instructions for predictably reproducing the issue. Even better,
you can
[submit a Pull Request](https://github.com/hhftechnology/middleware-manager/blob/master/CONTRIBUTING.md#submitting-a-pull-request)
with a fix.

# New Features

You can request a new feature by
[submitting a GitHub Issue](https://github.com/hhftechnology/middleware-manager/issues/new).
If you would like to implement a new feature, please consider the scope of the
new feature:

-   _Large feature_: first
    [submit a GitHub Issue](https://github.com/hhftechnology/middleware-manager/issues/new)
    and communicate your proposal so that the community can review and provide
    feedback. Getting early feedback will help ensure your implementation work
    is accepted by the community. This will also allow us to better coordinate
    our efforts and minimize duplicated effort.
-   _Small feature_: can be implemented and directly
    [submitted as a Pull Request](https://github.com/hhftechnology/middleware-manager/blob/master/CONTRIBUTING.md#submitting-a-pull-request).

# Contributing Code

Middleware Manager follows the "Fork-and-Pull" model for accepting contributions.

### Initial Setup

Setup your GitHub fork and continuous-integration services:

1. Fork the [Middleware manager repository](https://github.com/hhftechnology/middleware-manager)
   by clicking "Fork" on the web UI.

2. All contributions must pass all checks and reviews to be accepted.

Setup your local development environment:

```bash
# Clone your fork
git clone https://github.com/<username>/middleware-manager.git

# Configure upstream alias
cd middleware-manager
git remote add upstream https://github.com/hhftechnology/middleware-manager.git
```

### Submitting a Pull Request

#### Branch

For each new feature, create a working branch:

```bash
# Create a working branch for your new feature
git branch --track <branch-name> origin/main

# Checkout the branch
git checkout <branch-name>
```

#### Create Commits

```bash
# Add each modified file you'd like to include in the commit
git add <file1> <file2>

# Create a commit
git commit
```

This will open up a text editor where you can craft your commit message.

#### Upstream Sync and Clean Up

Prior to submitting your pull request, you might want to do a few things to
clean up your branch and make it as simple as possible for the original
repository's maintainer to test, accept, and merge your work.

If any commits have been made to the upstream master branch, you should rebase
your development branch so that merging it will be a simple fast-forward that
won't require any conflict resolution work.

```bash
# Fetch upstream main and merge with your repository's main branch
git checkout main
git pull upstream main

# If there were any new commits, rebase your development branch
git checkout <branch-name>
git rebase main
```

Now, it may be desirable to squash some of your smaller commits down into a
small number of larger more cohesive commits. You can do this with an
interactive rebase:

```bash
# Rebase all commits on your development branch
git checkout <branch-name>
git rebase -i main
```

This will open up a text editor where you can specify which commits to squash.

#### Push and Test

```bash
# Checkout your branch
git checkout <branch-name>

# Push to your GitHub fork:
git push origin <branch-name>
```

This will trigger the continuous-integration checks. You can view the results in
the respective services. Note that the integration checks will report failures
on occasion.

#### Pull requests

Aim to make pull requests easy to read both when viewed in a list (title only)
as well as clear in content within the description.

##### Title formatting

Describe the change as a one-line in some descriptive manner. Add sufficient
context for a reader to understand what is improved. If platform-specific
consider adding the platform as a prefix, like `[Android]` or any other tags may
be useful for quick filtering like `[TC-ABC-1.2]` to tag test changes.

Examples of descriptive titles:

-   `[Silabs] Fix compile of SiWx917 if LED and BUTTON are disabled`
-   `[Telink] Update build Dockerfile with new Zephyr SHA: c05c4.....`
-   `General Commissioning Cluster: use AttributeAccessInterface/CommandHandlerInterface for processing`
-   `Scenes Management/CopyScene: set access as manage instead of default to match the spec`
-   `Fix build errors due to ChipDeviceEvent default constructor not being available`
-   `Fix crash during DNSSD processing due to malformed packet`
-   `[NRF] Fix crash due to stack overflow during logging for PW-RPC builds`
-   `[TC-ABC-2.3] added new python test case based on test plan`
-   `[TC-ABC] migrate tests from yaml to python`

Examples of titles that are vague (not clear what the change is, one would need
to open the pull request for details or open additional issue in GitHub)

-   `Work on issue 1234`
-   `Fix android JniTypeWrappers`
-   `Fix segfault in BLE`
-   `Fix TC-ABC-1.2`
-   `Update Readme`

##### Summary contents

Ensure that there is sufficient detail in issue summaries to make the content of
the PR clear:

-   a `TLDR` of the change content. This is a judgment call on details,
    generally you should include a what was changed and why. The change is
    trivial/short, this can be very short (i.e. "fixed typos" is perfectly
    acceptable, however if changing 100-1000s of line, the areas of changes
    should be explained)
-   If a crash/error is fixed, explain the root cause and if the fix is not
    obvious (again, judgment call), explain why the given approach was taken.
-   Help the reviewer out with any notable information (specific platform
    issues, extra thoughts or requests for feedback or gotchas on tricky code,
    followup work or PR dependencies)
-   TIP: use the syntax of `Fixes #....` to mark issues completed on PR merge or
    use `#...` to reference issues that are addressed.
-   TIP: prefer adding some brief description (especially about the content of
    the changes) instead of just referencing an issue (helps reviewers get
    context faster without extra clicks).

##### Testing section

All Pull Requests **MUST** contain a `#### Testing` section that describes how
the pull request was tested. Ideally every test should have automated testing,
however for platform specific changes or hardware-specific issues we may not be
able to have such tests. As such, manual testing is acceptable, however the 
description has to be detailed intentionally to avoid a bias towards marking 
pull requests as "manually tested" out of convenience.

-   Automated testing

    **AWESOME**. You can say "unit tests added/updated" or "Integration tests
    updated to cover functionality" or "existing tests already cover this" (make
    sure they do. Integration tests often only cover happy paths).

    Add any notes on not covered things. It is a judgment call on how much can
    be covered as 100% sounds great however not always possible.

-   Manual testing

    Describe why automated testing is impossible in the current CI environment
    or difficult to add. If adding later, reference the issue to add automation
    and a timeline for adding such automation.

    Describe in **DETAIL** how manual testing was done: what environment was used.
    Describe commands ran and physical interaction and what was observed.

-   Trivial/obvious change

    In rare cases the change is trivial (e.g. fixing a typo in a `Readme.md`).
    Scripts still require a `#### Testing` section however you can be brief like
    `N/A` or `checked new URL opens`. Note that these cases are rare - e.g.
    fixing a typo in an ID still requires some description on how you checked
    that the new ID takes effect.

> [!TIP]
>
> When working on a pull request please refrain from using the "Update
> branch" feature in the GitHub UI too often. Updating the PR branch in this way
> triggers the CI workflows cancellation and restart. This feature should be
> used only when a PR has not been worked on for a long time and a lot of
> divergence has accumulated. Your PR branch being out of sync with master is
> not a blocker for merging an approved PR.

### Review Requirements

#### Submit Pull Request

Once you've validated your changes with tests, go to the page for your fork on GitHub,
select your development branch, and click the pull request button. If you need
to make any adjustments to your pull request, just push the updates to GitHub.
Your pull request will automatically track the changes on your development
branch and update.

#### Merge Requirements

-   Github Workflows pass
-   Builds pass
-   Tests pass
-   Linting passes
-   Code style passes

When can I merge? After these have been satisfied, a reviewer will merge the PR
into master

#### Documentation

## Merge Processes

Merges require 1 approval(s) from unique require-reviewers lists, and all tests passing.

### Fast track types

### Trivial changes

Small changes or changes that do not affect the main functionality of the code
can be fast tracked immediately. Examples:

-   Adding/removing documentation (.md files)
-   Adding tests (may include small reorganization/method adding/changes to
    enable testability):
    -   Test scripts
    -   Additional tests following a pattern (e.g. YAML tests)
-   Adding/updating/fixing tooling to aid in development
-   Code readability refactors:
    -   renaming enum/classes/structure members
    -   moving constant header location
    -   Obviously trivial build rule changes (e.g. adding missing files to build
        rules)
    -   Changing comments
    -   Adding/removing includes (include what you need and only what you need
        rules)
-   Pulling new third-party repo files
-   Most changes to existing docker files (pulling new versions, reorganizing)
-   Most changes to new dockerfile version in workflows

#### Fast track changes

Larger functionality changes are allowed to be fast tracked with these
requirements/restrictions:

-   Require at least 1 day to have passed since the creation of the PR
-   Require at least 1 checkmark from someone familiar with the code or problem
    space
    -   This requirement shall be dropped after a PR is 3 days old with stale or
        no feedback.
-   Code is sufficiently covered by automated tests (or impossible to
    automatically test with a very solid reason for this - e.g. changes to BLE
    parameters cannot be automatically tested, but should have been manually
    verified)

Fast tracking these changes will involve resolving any obviously 'resolved'
comments (judgment call here: were they replied to or addressed) and merging the
change.

Any "request for changes" marker will always be respected unless obviously
resolved (i.e. author marked "requesting changes because of X and X was done in
the PR")

-   This requirement shall be dropped after a PR is 3 days old with stale or no
    feedback.
