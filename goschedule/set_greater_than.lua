local key = KEYS[1]
local value = tonumber(ARGV[1])

if redis.call("EXISTS", key) == 1 then
  local oldValue = tonumber(redis.call('GET', key))

  if value <= oldValue then
    return false
  end
end

redis.call('SET', key, value)

return true
