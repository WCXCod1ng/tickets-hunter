-- popDelayTaskBatchScript:
-- 该脚本用于获取batch个可执行的任务，并将它们从延迟队列中移除，保证只有一个消费者能拿到这些任务。
-- 它是对popDelayTaskScript的批量版本，适用于需要一次获取多个任务的场景，减小消息处理延迟。

local queue = KEYS[1]
local now = ARGV[1]
local batch_size = ARGV[2]

-- 查找 score 在 0 到 now 之间的 batch_size 个元素
local tasks = redis.call('ZRANGEBYSCORE', queue, 0, now, 'LIMIT', 0, batch_size)

-- 查到后立即从 ZSet 中移除，保证只有一个消费者能拿到该任务！
if #tasks > 0 then
    redis.call('ZREM', queue, unpack(tasks))
    return tasks
end

return {}