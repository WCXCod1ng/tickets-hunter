
-- 解锁座位脚本
-- 该脚本用于解锁座位，确保只有持有正确订单号的请求才能解锁对应的座位，防止误解锁。
-- 返回值：
-- 1 - 解锁成功
-- 0 - 解锁失败，订单号不匹配或锁不存在

---- KEYS[1] - 锁定的key
--local lock_key = KEYS[1]
---- ARGV[1] - 订单号，用于验证锁定的订单号是否匹配，防止误解锁
--local order_sn = ARGV[1]
--
--if redis.call("GET", lock_key) == order_sn then
--    -- 只有当锁定的订单号与传入的订单号匹配时才执行解锁，防止误解锁
--    return redis.call("DEL", lock_key)
--else
--    return 0 -- 解锁失败，订单号不匹配或锁不存在
--end

local bitmap_key = KEYS[1]
local lock_key = KEYS[2]

local seat_index = ARGV[1]
local order_sn = ARGV[2]

-- 1. 校验锁的状态
local lock_order = redis.call("GET", lock_key)

-- 锁存在且属于自己
if lock_order == order_sn then
    -- 2. 恢复BitMap
    redis.call(
            "BITFIELD",
            bitmap_key,
            "SET", "u1", "#" .. seat_index, 0
    )
    -- 3. 删除锁
    redis.call("DEL", lock_key)

    return 1 -- 解锁成功
end

-- 情况3：锁存在，但不属于自己，直接退出

return 0