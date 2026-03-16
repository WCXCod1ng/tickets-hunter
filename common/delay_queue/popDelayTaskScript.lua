--popDelayTaskScript: 从 ZSet 中原子性地弹出一个超时的任务，保证只有一个消费者能拿到该任务

local queue = KEYS[1]
local now = ARGV[1]

-- 查找 score 在 0 到 now 之间的 1 个元素
local tasks = redis.call('ZRANGEBYSCORE', queue, 0, now, 'LIMIT', 0, 1)

-- 查到后立即从 ZSet 中移除，保证只有一个消费者能拿到该任务！
if #tasks > 0 then
    local order_sn = tasks[1]
    redis.call('ZREM', queue, order_sn)
    return order_sn
end

return nil