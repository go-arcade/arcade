package core

import "errors"

var (
	// ErrPluginAlreadyExists 插件已存在错误
	ErrPluginAlreadyExists = errors.New("plugin already exists")
	// ErrPluginNotFound 插件未找到错误
	ErrPluginNotFound = errors.New("plugin not found")
)
