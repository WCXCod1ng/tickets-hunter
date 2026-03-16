package ticket_seat

// 座位状态常量
const (
	SeatStatusAvailable = 0 // 可售
	SeatStatusLocked    = 1 // 锁定（已被订单锁定，等待支付结果）
	SeatStatusSold      = 2 // 已售（支付成功后更新为已售）
)
