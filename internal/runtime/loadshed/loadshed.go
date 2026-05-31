package loadshed

func Reject(depth uint64, limit uint64) bool {
	return depth > limit
}
