-- Lua Script
local currKey = KEYS[1]
local prevKey = KEYS[2]

local limit = tonumber(ARGV[1])
local prevWeight = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])

local currCount = tonumber(redis.call('GET', currKey) or "0")
local prevCount = tonumber(redis.call('GET', prevKey) or "0")

local count = currCount + (prevCount * prevWeight)

if count < limit then
    redis.call('INCR', currKey)
    -- FIX: Use math.ceil to ensure Integer for EXPIRE
    redis.call('EXPIRE', currKey, math.ceil(ttl * 2))
    return 1
else
    return 0
end