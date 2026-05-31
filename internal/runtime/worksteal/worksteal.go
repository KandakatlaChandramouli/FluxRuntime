package worksteal

type Scheduler struct{}

func (s *Scheduler) Steal() bool {
	return false
}
