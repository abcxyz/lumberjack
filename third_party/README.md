**Lumberjack is not an official Google product.**

# Third party dependencies

Locally-defined protos cannot depend on protos in a remote repository because
the `protoc` compiler
[cannot fetch files from remote repositories](https://github.com/golang/protobuf/issues/1072).
To allow our locally-defined protos to re-use open source protos, we vendor them
in our `third_party` folder with git subtrees.

## Instructions

```shell
make update_third_party
```

To update a subset of third_party, see the other targets in the `Makefile`.
