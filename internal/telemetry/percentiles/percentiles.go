package percentiles

func Compute(data []uint64, p float64) uint64 {
	if len(data) == 0 {
		return 0
	}

	idx := int(float64(len(data)-1) * p)

	return data[idx]
}
