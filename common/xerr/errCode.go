package xerr

// 0-16已经由go-zero内部定义的状态码占用了

// 用户模块
const (
	LoginFailed uint32 = 10000 + iota
)

// 票务模块
const (
	LockSeatFailed uint32 = 20000 + iota
)

// 订单模块
const ()

// 支付模块
const (
	DeductFailed uint32 = 40000 + iota
)
