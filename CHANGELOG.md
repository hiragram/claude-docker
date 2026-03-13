# Changelog

## [0.9.0](https://github.com/hiragram/agent-workspace/compare/v0.8.1...v0.9.0) (2026-03-13)


### Features

* enable sudo for claude user in Docker container ([e0fb33d](https://github.com/hiragram/agent-workspace/commit/e0fb33d406def5606fe252102663800827afb440))
* update builtin profiles and enable sudo for claude user ([9a68530](https://github.com/hiragram/agent-workspace/commit/9a68530793827b4b92703922651572652c4b050c))

## [0.8.1](https://github.com/hiragram/agent-workspace/compare/v0.8.0...v0.8.1) (2026-03-04)


### Bug Fixes

* use already-built Docker image in zellij claude pane ([26f5f40](https://github.com/hiragram/agent-workspace/commit/26f5f40b69bec67385b75a0b3a2ec6b6e3e32c66))
* use already-built Docker image in zellij claude pane ([513d63f](https://github.com/hiragram/agent-workspace/commit/513d63f5fdf5c652ddd259158a2c9d2d43aec4d5))

## [0.8.0](https://github.com/hiragram/agent-workspace/compare/v0.7.1...v0.8.0) (2026-03-04)


### Features

* add default-dockerfile command and custom Dockerfile support ([4e3308c](https://github.com/hiragram/agent-workspace/commit/4e3308caa250e07ec7b3dadde8f81193edcceab8))
* add default-dockerfile command and custom Dockerfile support ([227b61d](https://github.com/hiragram/agent-workspace/commit/227b61df6067a40ca72af24916d8f62c668f39a9))

## [0.7.1](https://github.com/hiragram/agent-workspace/compare/v0.7.0...v0.7.1) (2026-03-01)


### Bug Fixes

* pass profile env vars to child processes via .aw-profile-env ([9380e3d](https://github.com/hiragram/agent-workspace/commit/9380e3d9e3988d569fe74b28f43e8d5a27a33b1c))
* pass profile env vars to child processes via .aw-profile-env ([a520685](https://github.com/hiragram/agent-workspace/commit/a52068584de02bb63a78054fd1ce20d9ad81f70a))

## [0.7.0](https://github.com/hiragram/agent-workspace/compare/v0.6.0...v0.7.0) (2026-03-01)


### Features

* add custom environment variable passthrough to Docker containers ([b9e67b4](https://github.com/hiragram/agent-workspace/commit/b9e67b40f04b16af4e591cc4316e166dcd8817a9))


### Bug Fixes

* handle errcheck lint for file Close in envfile parser ([cf2863a](https://github.com/hiragram/agent-workspace/commit/cf2863a5549f3547c56d8787a6899704ab62f7b8))

## [0.6.0](https://github.com/hiragram/agent-workspace/compare/v0.5.0...v0.6.0) (2026-03-01)


### Features

* add on-end lifecycle hook ([e59a776](https://github.com/hiragram/agent-workspace/commit/e59a7766b97ffb61bd3bfb0bb97c7b9e2bc5dbbb))
* add on-end lifecycle hook for worktree profiles ([9e64201](https://github.com/hiragram/agent-workspace/commit/9e642016c548b307ccdcebb4382b7690a3e20c94))

## [0.5.0](https://github.com/hiragram/agent-workspace/compare/v0.4.0...v0.5.0) (2026-03-01)


### Features

* add on-create hook for worktree profiles ([0092cc6](https://github.com/hiragram/agent-workspace/commit/0092cc6c871a169033554fc33c0495653a60325d))
* add on-create hook for worktree profiles ([3d7293e](https://github.com/hiragram/agent-workspace/commit/3d7293ebf912d7a644b93ad474596adecb3b0007))
* install Go 1.23.6 in Docker container ([ba59007](https://github.com/hiragram/agent-workspace/commit/ba590071391ce7a88fbec789b7404a2d06128d83))
* install Go 1.23.6 in Docker container ([4d01510](https://github.com/hiragram/agent-workspace/commit/4d01510deba918fdc52c5b1f59eaad63b3e1fd4c))
* merge builtin profiles with user config ([94b57d4](https://github.com/hiragram/agent-workspace/commit/94b57d44de1ead4cfe1c60f15d07ea704c871c19))
* merge builtin profiles with user config ([c97dfc7](https://github.com/hiragram/agent-workspace/commit/c97dfc7f3edcffa5e4878973638251ad82d0384e))

## [0.4.0](https://github.com/hiragram/agent-workspace/compare/v0.3.0...v0.4.0) (2026-03-01)


### Features

* add `aw profiles` subcommand ([1c4964d](https://github.com/hiragram/agent-workspace/commit/1c4964d1a66aaeaeb9f5a2152ea59530e2972640))
* add `aw profiles` subcommand to list available profiles ([f7ac122](https://github.com/hiragram/agent-workspace/commit/f7ac122c28addbafabdd388720fb97a9ce7ed082))
* rename docker-claude to claude and add worktree-zellij as default ([8c713cf](https://github.com/hiragram/agent-workspace/commit/8c713cf5a1dfc3c02673a195add47aaad2ff4d09))
* rename docker-claude to claude and add worktree-zellij as default builtin ([c20b69a](https://github.com/hiragram/agent-workspace/commit/c20b69aadc7f915cd153e9a0c7422be472bda4e3))

## [0.3.0](https://github.com/hiragram/agent-workspace/compare/v0.2.0...v0.3.0) (2026-02-28)


### Features

* refactor claude-docker into agent-workspace with YAML profile-based configuration ([2608e4b](https://github.com/hiragram/agent-workspace/commit/2608e4b52fc3b47522929690fdca1b61b10ffe32))
* refactor to agent-workspace with YAML profile config ([ca0855b](https://github.com/hiragram/agent-workspace/commit/ca0855b5a83f39812e9b1c5870d63f362b3b647b))


### Bug Fixes

* goreleaser archive name to match binary name ([3b86a4b](https://github.com/hiragram/agent-workspace/commit/3b86a4b70b042b51842c67a18fd67afaaf00ec22))
* use binary name in goreleaser archive name_template ([1a0f30f](https://github.com/hiragram/agent-workspace/commit/1a0f30fe9b03681c6d825317436d1bc7b967b6a1))

## [0.2.0](https://github.com/hiragram/claude-docker/compare/v0.1.0...v0.2.0) (2026-02-28)


### Features

* add --version flag and release-please integration ([70cddbe](https://github.com/hiragram/claude-docker/commit/70cddbe61f53136b8d64efb48a1884b44266a389))
* add --worktree flag for zellij-based dev environment ([cd0b591](https://github.com/hiragram/claude-docker/commit/cd0b5910b9667c8c5f7818e085717924a0cc1f03))
* add --worktree flag for zellij-based dev environment ([4e86d56](https://github.com/hiragram/claude-docker/commit/4e86d56863ebc5e309a94e5a79b1e1dff696e2d9))
* add `update` subcommand for self-updating ([6d0bdfb](https://github.com/hiragram/claude-docker/commit/6d0bdfb1ec5725db7ec818c74a89e29e9fb676e0))
* add `update` subcommand for self-updating ([70ed382](https://github.com/hiragram/claude-docker/commit/70ed3829bd3f816d5523249c765465658dcaf6fc))
* add curl|bash installer and rename to claude-docker ([0780317](https://github.com/hiragram/claude-docker/commit/078031742b1de7a3ff49ed03d410fcc4b60a9bd0))
* add python3, node, pnpm to container image ([5e7dbd0](https://github.com/hiragram/claude-docker/commit/5e7dbd07467b410352493c5c6da5f69c21037754))
* add single-script Docker launcher for Claude Code ([ff66d5e](https://github.com/hiragram/claude-docker/commit/ff66d5e8afc5cb342c2f8434b8ce6daf3339a237))
* mount workspace with host path and add --rebuild option ([188c106](https://github.com/hiragram/claude-docker/commit/188c106582fe5a68976304d7751e4b0877f6c7af))
* persist Claude Code installation in named volume ([404080b](https://github.com/hiragram/claude-docker/commit/404080ba27e5869e5ac80bc8e03f63621c026029))
* share host git/gh/ssh config with container ([71a58ff](https://github.com/hiragram/claude-docker/commit/71a58ff32a9082f6a019f8c76e11c21804357bdf))


### Bug Fixes

* add openssh-client for git SSH operations ([34e584c](https://github.com/hiragram/claude-docker/commit/34e584c29fb440ef69005f77994ab002a8bf4498))
* check error return values in tests to satisfy errcheck lint ([fad1a2a](https://github.com/hiragram/claude-docker/commit/fad1a2a6387f4ea055dadf13775b88f72f136937))
* handle errcheck lint warnings in update package ([6e7b35d](https://github.com/hiragram/claude-docker/commit/6e7b35dbb3ed0489aa5df1fcead3e88b999503c2))
* resolve all golangci-lint issues ([fad510d](https://github.com/hiragram/claude-docker/commit/fad510d67cf5ec3ba730381b0657c4b86c9d65d5))
* run as non-root user to allow --dangerously-skip-permissions ([5f67c8b](https://github.com/hiragram/claude-docker/commit/5f67c8bbe4a684e99d7721a8afe0bbb5d6f63870))
* symlink host claude path for plugin resolution ([23fa128](https://github.com/hiragram/claude-docker/commit/23fa1284c8c7cb3f34e969e2db40ff48ef0cc8c4))
* use golangci-lint-action v7 with golangci-lint v2 ([178c9c9](https://github.com/hiragram/claude-docker/commit/178c9c9ae257fbdcef360286f2269a4ee6241383))
* use separate config dir and persist onboarding state ([bf9d312](https://github.com/hiragram/claude-docker/commit/bf9d312db00828ac0877cf557bbbc1ac78e21b29))
