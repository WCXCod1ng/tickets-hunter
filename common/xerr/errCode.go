package xerr

// 0-16已经由go-zero内部定义的状态码占用了

// 用户模块
const (
	LoginFailed uint32 = 10000 + iota
)
