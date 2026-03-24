--定制一份安全版的强制同步Lua脚本，用于实现如果MySQL座位已经被释放，恢复Redis中的座位状态

local bitmap_key = KEYS[1]
local lock_key = KEYS[2]
local seat_index = ARGV[1]

-- 1. 查看当前有没有人在持有这把锁
local lock_order = redis.call("GET", lock_key)

-- 2. 【核心防御】如果锁存在，说明就在你查完 MySQL 到执行这段 Lua 的微秒级间隙里，
-- 有新的用户刚刚抢到了这个座位！
-- 此时绝对不能碰 BitMap，尊重当前正在发生的交易！
if lock_order then
    return 0 -- 同步被拒绝，因为存在活跃交易
end

-- 3. 如果锁不存在 (nil)：
-- 既然 MySQL 告诉我们这个座位状态是 0，且当前没有人持有抢票锁。
-- 那么不管 BitMap 之前是 1 还是什么，它绝对是一个遗留的死锁幽灵！
-- 安全强制覆盖为 0！
redis.call("BITFIELD", bitmap_key, "SET", "u1", "#" .. seat_index, 0)

return 1 -- 同步成功