package lib

import (
	"runtime/debug"
)

const (
	// ライブラリの名前
	Name = "sabafly-lib"
	// ライブラリのモジュール名
	Module = "github.com/sabafly/sabafly-lib"
	// ライブラリのGitHubリポジトリのリンク
	GitHub = "https://github.com/sabafly/sabafly-lib"
)

var (
	// ライブラリのバージョン
	Version = getVersion()
)

func getVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range bi.Deps {
			if dep.Path == Module {
				return dep.Version
			}
		}
	}
	return "unknown"
}
