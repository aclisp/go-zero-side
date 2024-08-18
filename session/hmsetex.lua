local key = KEYS[1]
local expire_time = ARGV[1]

-- Remove the first argument from ARGV which is the expire time
table.remove(ARGV, 1)

-- Execute HMSET command
redis.call('HMSET', key, unpack(ARGV))

-- Set expiration time
redis.call('EXPIRE', key, expire_time)

return "OK"
