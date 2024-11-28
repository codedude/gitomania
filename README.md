# tig

Git clone in Go

## Steps

1. init and track files: **DONE**
    - [x] init: create a directory .tig with data X
    - [x] status: show tracked and untracked files X
    - [x] add: track file X
    - [x] rm: untrack file X
2. commit:
    - [x] Add modified/created files to the commit X
    - [x] Remove staged files X
    - [x] Commit changes
    - [ ] List commit
3. revert, head:
    - [ ] Revert to a specific commit
    - [ ] Delete a commit
    - [ ] Reset head

4. branch:
    - [ ] Create a branch
    - [ ] Switch branch

## Bugs
- add X, edit X, add X, back to 1st, add X => it uses the first snapshot (same hash)
  - If rm a staged file = snapshot can be linked to many commits
