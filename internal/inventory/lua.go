package inventory

// luaReservationScript is the atomic reservation script.
// Preloaded at startup via SCRIPT LOAD; SHA1 cached in Store.
// Must never be reloaded on the hot path.
//
// KEYS[1] = {event:<id>}:tickets
// KEYS[2] = {event:<id>}:orders
// ARGV[1] = requested_quantity
// ARGV[2] = reservation_id
// ARGV[3] = audit_payload
const luaReservationScript = `
local stock = tonumber(redis.call('GET', KEYS[1]))

if not stock then
    return -2
end

local quantity = tonumber(ARGV[1])

if stock < quantity then
    return -1
end

redis.call('DECRBY', KEYS[1], quantity)

redis.call(
    'HSET',
    KEYS[2],
    ARGV[2],
    ARGV[3]
)

return 1
`
