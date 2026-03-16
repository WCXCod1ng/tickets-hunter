package redis

import "fmt"

// 有效座位集合的 Redis Key，格式为 "ticket:event:{eventId}:valid_seats"
func ValidSeatRedisKey(eventId int64) string {
	return fmt.Sprintf("ticket:seat:%d:valid_seats", eventId)
}

// 针对有效座位集合的优化：使用BitMap
func SeatBitMapRedisKey(eventId int64, section string) string {
	return fmt.Sprintf("ticket:seat:bitmap:{%d:%s}", eventId, section)
}

// 座位锁定的 Redis Key，格式为 "ticket:seat:{seatId}:lock"
func SeatLockRedisKey(seatId int64) string {
	return fmt.Sprintf("ticket:seat:%d:lock", seatId)
}

// 订单延迟队列的 Redis Key
const OrderDelayQueueKey = "order:delay_queue"
