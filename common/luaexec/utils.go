package luaexec

import (
	"os"
	"path/filepath"
)

// 加载Lua脚本文件
func LoadLuaFile(path string) (string, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// 必须加载Lua脚本文件，否则程序无法启动
func MustLoadLuaFile(path string) string {
	script, err := LoadLuaFile(path)
	if err != nil {
		panic(err)
	}
	return script
}
