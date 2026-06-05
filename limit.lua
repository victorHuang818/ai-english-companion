-- 引入 OpenResty 自带的 Redis 库
local redis = require "resty.redis"
local red = redis:new()

red:set_timeouts(1000, 1000, 1000) -- 1秒超时

-- 连接 Redis
local ok, err = red:connect("192.168.1.100", 6379)
if not ok then
    ngx.log(ngx.ERR, "failed to connect to redis: ", err)
    return ngx.exit(500)
end

-- 优先获取代理透传的真实客户端 IP
local client_ip = ngx.req.get_headers()["x-real-ip"] or ngx.req.get_headers()["x-forwarded-for"] or ngx.var.remote_addr
if type(client_ip) == "table" then
    client_ip = client_ip[1]
end
if client_ip then
    client_ip = string.match(client_ip, "%s*([^,]+)") -- 若有多个IP，取第一个
end

local limit_key = "rate_limit:" .. (client_ip or "unknown")

-- 定义 Redis 内部执行的 Lua 脚本，保证 INCR 和 EXPIRE 的原子性
-- 这样既防止了并发竞争，也防止了网络中断导致的 TTL 永久失效
local lua_script = [[
    local key = KEYS[1]
    local limit = tonumber(ARGV[1])
    local expire_time = tonumber(ARGV[2])
    
    local current = redis.call('get', key)
    if current and tonumber(current) >= limit then
        return tonumber(current) + 1
    end
    
    local val = redis.call('incr', key)
    if val == 1 then
        redis.call('expire', key, expire_time)
    end
    return val
]]

-- 执行 Redis 内部原子操作：限制每 60 秒最多 5 次请求
local current, err = red:eval(lua_script, 1, limit_key, 5, 60)

if not current then
    ngx.log(ngx.ERR, "failed to run rate limit eval: ", err)
    -- 如果 Redis 评估出错，安全放行，并将连接归还
    red:set_keepalive(10000, 100)
    return
end

-- 只要拿到了 Redis 的执行结果，立刻把连接放回连接池，绝不拖延
red:set_keepalive(10000, 100)

if current > 5 then
    -- 超过阈值，直接拦截！
    -- 【重要】先设置 Header，再 ngx.say 发送数据
    ngx.header.content_type = "application/json; charset=utf-8"
    ngx.status = 429
    ngx.say('{"error": "Too Many Requests. Please slow down."}')
    return ngx.exit(429)
end
