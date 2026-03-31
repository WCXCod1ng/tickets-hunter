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

// 存放静态座位信息（优化createOrder查MySQL的过程），每个座位的信息都是一个Hash结构
func SeatStaticInfoRedisKey(seatId int64) string {
	return fmt.Sprintf("ticket:seat:static_info:{%d}", seatId)
}

// 以区域为单位的静态座位信息集合，Key格式为 "ticket:seat:section_static_info:{eventId}:{section}"
func SectionSeatStaticInfoRedisKey(eventId int64, section string) string {
	return fmt.Sprintf("ticket:seat:section_static_info:{%d:%s}", eventId, section)
}

// 订单延迟队列的 Redis Key
const OrderDelayQueueKey = "order:delay_queue"
