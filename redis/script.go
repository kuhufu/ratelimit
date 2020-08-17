package redis

const script = `
-- key
local key = KEYS[1]
-- 最大存储的令牌数
local burst = tonumber(KEYS[2])
-- 每秒钟产生的令牌数
local limit = tonumber(KEYS[3])
-- 请求的令牌数
local N = tonumber(ARGV[1])

-- 下次请求可以获取令牌的起始时间
local last = tonumber(redis.call('hget', key, 'last') or 0)

-- 当前存储的令牌数
local tokens = tonumber(redis.call('hget', key, 'tokens') or 0)

-- 当前时间
local time = redis.call('time')
local now = tonumber(time[1]) * 1000000 + tonumber(time[2])

-- 添加令牌的时间间隔
local interval = 1000000 / limit

-- 距离上次获取流逝的时间
local max_elapsed = (burst - tokens)*interval

local elapsed = now - last
if (max_elapsed < elapsed) then
	elapsed = max_elapsed
end


-- 补充令牌
local new_tokens = elapsed / interval
tokens = math.min(burst, tokens + new_tokens)

-- 消耗令牌
local fresh_permits = math.max(N - tokens, 0);
local wait_micros = fresh_permits * interval

if (tokens >= N) then
	redis.call('hset', key, 'tokens', tokens - N)
	redis.call('hset', key, 'last', now)
end

-- redis.replicate_commands()

redis.call('expire', key, 1000)

-- 返回需要等待的时间长度
return wait_micros
`

/*
eval 'lua脚本' 3 '自定义的key' '最大存储的令牌数' '每秒钟产生的令牌数' '请求的令牌数'
redis将返回获取请求成功后，线程需要等待的微秒数
*/

/*
https://www.jianshu.com/p/d322c475f2c1
*/
