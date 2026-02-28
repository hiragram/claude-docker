package worktree

import _ "embed"

//go:embed embed/plans-watcher.sh
var plansWatcherSh []byte

//go:embed embed/git-diff-picker.sh
var gitDiffPickerSh []byte

//go:embed embed/pr-status.sh
var prStatusSh []byte

//go:embed embed/layout.kdl.tmpl
var layoutKdlTmpl []byte
