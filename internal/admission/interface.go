package admission

type Controller interface {
	Allow(depth uint64) bool
}
