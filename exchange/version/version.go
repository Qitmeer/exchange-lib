package version

import (
	"fmt"
	"runtime"
)

var (
	// 初始化为 unknown，如果编译时没有传入这些值，则为 unknown
	GitCommitLog   = "unknown_unknown"
	GitStatus      = "unknown_unknown"
	BuildTime      = "unknown_unknown"
	BuildGoVersion = "unknown_unknown"
	Version        = "v0.1.1"
)

// 返回单行格式
func StringifySingleLine() string {
	if GitStatus != "" {
		GitCommitLog = GitCommitLog[0:10] + "-dirty"
	} else {
		GitCommitLog = GitCommitLog[0:10]
	}
	return fmt.Sprintf("Qitmeer-utxo Version=%s. GitCommitLog=%s. GitStatus=%s. BuildTime=%s. GoVersion=%s. runtime=%s/%s.",
		Version, GitCommitLog[0:10], GitStatus, BuildTime, BuildGoVersion, runtime.GOOS, runtime.GOARCH)
}

// 返回多行格式
func StringifyMultiLine() string {
	if GitStatus != "" {
		GitCommitLog = GitCommitLog[0:10] + "-dirty"
	} else {
		GitCommitLog = GitCommitLog[0:10]
	}
	return fmt.Sprintf("Qitmeer-utxo = %s\nGitCommit = %s\nBuildTime = %s\nGoVersion = %s\n",
		Version, GitCommitLog, BuildTime, BuildGoVersion)
}
