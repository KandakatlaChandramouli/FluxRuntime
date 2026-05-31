package gradient

type Predictor struct {
	Last float64
}

func (p *Predictor) Update(v float64) {
	p.Last = v
}
