// Package examples provides base and fork paths to compare upstream source
// with fork projects. Base paths come from Go module cache, fork paths from
// the current workspace.
//
// examples 提供基准路径和 fork 路径，用于比较上游源项目和 fork 项目的代码差异
// 基准路径来自 Go 模块缓存，fork 路径来自当前工作目录
package examples

import (
	"github.com/yylego/kratos-examples/demo1kratos"
	"github.com/yylego/kratos-examples/demo2kratos"
	"github.com/yylego/runpath"
)

// GetDemo1BasePath returns demo1kratos base path from Go module cache
//
// GetDemo1BasePath 返回模块缓存中 demo1kratos 的基准路径（上游源项目）
func GetDemo1BasePath() string {
	return demo1kratos.SourceRoot()
}

// GetDemo1ForkPath returns demo1kratos fork path in current workspace
//
// GetDemo1ForkPath 返回本地 fork 项目中的 demo1kratos 目录路径
func GetDemo1ForkPath() string {
	return runpath.PARENT.Join("demo1kratos")
}

// GetDemo2BasePath returns demo2kratos base path from Go module cache
//
// GetDemo2BasePath 返回模块缓存中 demo2kratos 的基准路径（上游源项目）
func GetDemo2BasePath() string {
	return demo2kratos.SourceRoot()
}

// GetDemo2ForkPath returns demo2kratos fork path in current workspace
//
// GetDemo2ForkPath 返回本地 fork 项目中的 demo2kratos 目录路径
func GetDemo2ForkPath() string {
	return runpath.PARENT.Join("demo2kratos")
}
