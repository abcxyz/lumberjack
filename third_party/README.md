# Third party dependencies

Locally-defined protos cannot depend on protos in a remote repository because
the `protoc` compiler
[cannot fetch files from remote repositories](https://github.com/golang/protobuf/issues/1072).
To allow our locally-defined protos to re-use open source protos, we vendor them
in our `third_party` folder with git subtrees.

## Prerequisites

Change into the lumberjack root directory. This directory contains the `.git`
file. For example,

```
cd ~/go/lumberjack
```

If this is the first time you add or update a third party repo, you need to add
it as a remote. The examples in this file use the third party repo
`https://github.com/googleapis/googleapis` and we name our remote `googleapis`.

```
git remote add -f googleapis https://github.com/googleapis/googleapis
```

## Instructions

The instructions below are based on
[Github's documentation for Git subtree merges](https://docs.github.com/en/get-started/using-git/about-git-subtree-merges#synchronizing-with-updates-and-changes).
Notice that updating a third party subtree has a different flow than adding a
third party subtree for the first time.

### Add third party repo

Merge the remote repo `googleapis` into your current branch. We use the `squash`
flag to avoid pulling in upstream commits.

```
git merge -s ours --squash --allow-unrelated-histories googleapis/master
```

Create a new directory called `third_party/googleapis` and copy the `googleapis`
repo into it.

```
git read-tree --prefix=third_party/googleapis -u googleapis/master
```

Commit the changes.

```
git commit -m "Add the repo googleapis as a subtree"
```

### Update an existing third party repo

Note that we want to
[automate this process](b/199776038) in the
future.

When we add a remote repo as a subtree, the subtree does not automatically sync
with the upstream changes. We need to manually sync with the commands below. The
`-X theirs` flag overwrites the local subtree because we assume that we never
introduce local changes.

```
git pull -s subtree -X theirs --squash --allow-unrelated-histories --no-rebase googleapis master
```

Commit the changes.

```
git commit -m "Update subtree googleapis"
```
