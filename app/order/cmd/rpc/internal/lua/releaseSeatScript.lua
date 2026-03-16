-- 实现释放座位的Lua脚本
-- 返回值
-- 0：释放座位失败，订单号不匹配或锁不存在
-- 1：释放座位成功

---- KEYS[1] - 锁定的key
--local lock_key = KEYS[1]
---- KEYS[2] - 合法座位集合的key
--local valid_seat_key = KEYS[2]
---- KEYS[3] - 延迟队列的key
--local delay_queue_key = KEYS[3]
---- ARGV[1] - 订单号，用于验证锁定的订单号是否匹配，防止误解锁
--local order_sn = ARGV[1]
---- ARGV[2] - 座位ID，用于从合法座位集合中移除
--local seat_id = ARGV[2]
--
---- 1. 把它加回合法座位池，防止破坏数据
---- 注意，这里不需要验证座位ID是否合法，因为即使是非法的座位ID加回池子也不会对系统造成影响，反而可以：（1）防止在极端情况下锁丢失导致座位无法被出售的问题，因为我们无论都会把座位加回合法座位集合中；
---- （2）还能防止恶意攻击者通过构造非法座位ID来占用合法座位池的资源，导致合法座位无法被锁定和出售。
--redis.call("SADD", valid_seat_key, seat_id)
--
---- 2. 安全释放锁 (解铃还须系铃人)
--if redis.call("GET", lock_key) == order_sn then
--    redis.call("DEL", lock_key)
--    -- 3. 从延迟队列中移除对应的订单号，优化性能，防止过期订单号占用资源
--    redis.call("ZREM", delay_queue_key, order_sn)
--    return 1
--else
--    -- 如果锁已经不存在，或者已经被别人抢走了(订单号不匹配)，千万不要乱加回池子！
--    return 0
--end


-- KEYS[1] - 锁key
local lock_key = KEYS[1]

-- KEYS[2] - bitmap key
local bitmap_key = KEYS[2]

-- KEYS[3] - 延迟队列key
local delay_queue_key = KEYS[3]

-- ARGV[1] - order_sn
local order_sn = ARGV[1]

-- ARGV[2] - seat index (bitmap中的位置)
local seat_index = ARGV[2]


-- 1. 把座位标记为可售
-- 对应原来的 SADD
redis.call("SETBIT", bitmap_key, seat_index, 0)

-- 2. 安全释放锁
if redis.call("GET", lock_key) == order_sn then

    redis.call("DEL", lock_key)

    -- 3. 删除延迟队列
    redis.call("ZREM", delay_queue_key, order_sn)

    return 1
else
    return 0
end