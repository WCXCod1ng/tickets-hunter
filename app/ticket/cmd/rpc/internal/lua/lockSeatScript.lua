--锁座 Lua 脚本
--返回值
--0：锁座成功
--1：锁座失败，由于其他原因（如座位不存在）
--2：座位已被锁定导致锁座失败

--local valid_seat_key = KEYS[1]
--local lock_key = KEYS[2]
--local seat_id = ARGV[1]
--local order_sn = ARGV[2]
--local ttl = ARGV[3]
--
---- 检查座位是否存在于有效座位集合中（防恶意越权攻击、刷单）
--if redis.call("SISMEMBER", valid_seat_key, seat_id) == 0 then
--    return 1 -- 座位不存在，锁座失败
--end
--
---- 尝试使用SETNX命令锁定座位
--local locked = redis.call("SET", lock_key, order_sn, "NX", "EX", ttl)
--if not locked then
--    return 2 -- 座位已被锁定，锁座失败
--end
--
--return 0 -- 锁座成功

local bitmap_key = KEYS[1]
local lock_key = KEYS[2]
local seat_index = ARGV[1] -- 座位的索引，注意不是ID
local order_sn = ARGV[2]
local ttl = ARGV[3]

-- 1. 从bitmap中检查当前座位的真实状态
local current_status = redis.call('BITFIELD', bitmap_key, 'GET', 'u1', '#'..seat_index)[1]
-- 2. 状态校验
if current_status ~= 0 then
    return 1 -- 座位不可用，已经被锁定或出售
end
-- 3. 尝试加锁
local locked = redis.call('SET', lock_key, order_sn, 'NX', 'EX', ttl)
if not locked then
    -- 座位已经被锁定，锁座失败
    return 2
end

-- 4. 加锁成功，设置bitmap_key
redis.call('BITFIELD', bitmap_key, 'SET', 'u1', '#'..seat_index, 1)

return 0 -- 锁座成功