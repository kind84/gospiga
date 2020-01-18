if redis.call("xack", KEYS[1], ARGV[1], ARGV[2]) == 1 then
	return redis.call("xadd", KEYS[2], "*", ARGV[3], ARGV[4])
end
return false
