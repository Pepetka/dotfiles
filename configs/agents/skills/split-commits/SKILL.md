---
name: split-commits
description: Splits staged changes into several semantic Conventional Commits after the user approves a commit plan. Use when the user wants staged changes split into multiple well-formed commits.
---

# Split staged changes into commits

1. Read `git status --short`, `git branch --show-current` and `git diff --cached`. If nothing is staged, say so and stop.

2. Propose a plan and wait for approval:
   - group files by meaning; each staged file goes into exactly one commit, none left out;
   - a rename's old and new paths go into the same commit;
   - subjects only, Conventional Commits in English: `feat(ui): add filters panel`;
   - order: refactors and foundations first, tests with or after their code;
   - if the branch name does not fit the changes, suggest `<type>/<short-description>`.

3. After approval, snapshot the staged state and commit group by group. The working tree is never touched, so unstaged changes stay in place:

   ```bash
   base=$(git rev-parse HEAD) tree=$(git write-tree)
   git reset -q
   git restore --source="$tree" --staged -- <files>   # per commit
   git commit -m "<subject>"
   ```

   Do not use `--no-verify`.

4. Verify: `git diff --quiet "$tree" HEAD` must succeed. Report `git log --oneline "$base"..HEAD`.

   On any failure, restore the original state and ask the user: `git reset -q --soft "$base" && git read-tree "$tree"`.
