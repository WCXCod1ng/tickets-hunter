package define

const (
	SeatLockTTL = 15*60 + 2*60 // 座位锁定的过期时间，单位为秒 (15分钟)，加上2分钟的缓冲期
	//SeatLockTTL = 1 * 60 + 2 * 60 // 测试用
)
