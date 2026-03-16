-- 用于实现Redis层次的出座操作：
-- 1. 从合法座位集合中移除座位，防止其他人再锁定这个座位，表示座位已经被卖出了
-- 2. 将指定座位的锁删除
-- 3. 从延迟队列中移除对应的订单号，优化性能，防止过期订单号占用资源

---- KEYS[1] - 合法座位集合的key
--local valid_seat_key = KEYS[1]
---- KEYS[2] - 锁定的key
--local lock_key = KEYS[2]
---- KEYS[3] - 延迟队列的key
--local delay_queue_key = KEYS[3]
---- ARGV[1] - 座位ID，用于从合法座位集合中移除
--local seat_id = ARGV[1]
---- ARGV[2] - 订单号，用于验证锁定的订单号是否匹配，防止误操作
--local order_sn = ARGV[2]
--
---- 1. 将指定座位的锁删除
--if redis.call("GET", lock_key) == order_sn then
--    -- 返回值1表示成功删除锁，0表示锁不存在或订单号不匹配
--    redis.call("DEL", lock_key)
--    -- 2. 只有是自己的锁才会从合法座位集合中移除座位，防止其他人再锁定这个座位，表示座位已经被卖出了
--    redis.call("SREM", valid_seat_key, seat_id)
--    -- 3. 从延迟队列中移除对应的订单号，优化性能，防止过期订单号占用资源
--    redis.call("ZREM", delay_queue_key, order_sn)
--    -- 这三个操作无论哪个失败都不影响最终结果，因为数据库的最终状态都是座位被卖出
--    -- 集合删除失败只会导致无效请求传入，不影响逻辑的正确性；锁删除失败还有TTL保护；延迟队列删除失败则会有定时任务定期清理过期订单号，所以都不影响最终结果
--    return 1 -- 返回1表示成功出座
--else
--    -- 如果锁已经不存在，或者已经被别人抢走了(订单号不匹配)，千万不要乱删锁！
--    return 0 -- 返回0表示删除锁失败，订单号不匹配或锁不存在
--end

-- KEYS[1] - 合法座位集合的bitmap_key
local bitmap_key = KEYS[1]
-- KEYS[2] - 锁定的key
local lock_key = KEYS[2]
-- KEYS[3] - 延迟队列的key
local delay_queue_key = KEYS[3]
-- ARGV[1] - 座位索引，用于从bitmap中清除标记
local seat_index = ARGV[1]
-- ARGV[2] - 订单号，用于验证锁定的订单号是否匹配，防止误操作
local order_sn = ARGV[2]

-- 1. 检查锁是否属于当前订单
if redis.call("GET", lock_key) == order_sn then

    -- 2. 删除锁
    redis.call("DEL", lock_key)

    -- 3. bitmap保持为1 (表示已售出)
    -- 注意：这里其实不用再操作bitmap，因为锁座时已经SETBIT=1
    -- redis.call("SETBIT", bitmap_key, seat_index, 1)
    

    -- 4. 删除延迟队列
    redis.call("ZREM", delay_queue_key, order_sn)

    return 1

else
    return 0
end