package inventory

import "strconv"

// ticketsKey constructs the inventory counter key for an event.
// Format: {event:<id>}:tickets
// Hash tag {event:<id>} guarantees Redis Cluster slot colocation.
//
// Stack allocation: result escapes to caller; no fmt.Sprintf.
func ticketsKey(eventID uint64, buf []byte) string {
	buf = buf[:0]
	buf = append(buf, '{', 'e', 'v', 'e', 'n', 't', ':')
	buf = strconv.AppendUint(buf, eventID, 10)
	buf = append(buf, '}', ':', 't', 'i', 'c', 'k', 'e', 't', 's')
	return string(buf)
}

// ordersKey constructs the reservation audit key for an event.
// Format: {event:<id>}:orders
func ordersKey(eventID uint64, buf []byte) string {
	buf = buf[:0]
	buf = append(buf, '{', 'e', 'v', 'e', 'n', 't', ':')
	buf = strconv.AppendUint(buf, eventID, 10)
	buf = append(buf, '}', ':', 'o', 'r', 'd', 'e', 'r', 's')
	return string(buf)
}

// auditPayload constructs the pipe-delimited audit string.
// Format: reservation_id|timestamp_ns|quantity|shard_id
// JSON is forbidden on the hot path per spec §6.7.
func auditPayload(reservationID uint64, timestampNs int64, quantity uint32, shardID int, buf []byte) string {
	buf = buf[:0]
	buf = strconv.AppendUint(buf, reservationID, 10)
	buf = append(buf, '|')
	buf = strconv.AppendInt(buf, timestampNs, 10)
	buf = append(buf, '|')
	buf = strconv.AppendUint(buf, uint64(quantity), 10)
	buf = append(buf, '|')
	buf = strconv.AppendInt(buf, int64(shardID), 10)
	return string(buf)
}
