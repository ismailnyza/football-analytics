package layout

func TerminalLayoutSize(termW, termH int) (w, h int) {
	w = termW
	h = termH
	if w <= 0 {
		w = 100
	}
	if h <= 0 {
		h = 28
	}
	const brokenW, brokenH = 24, 10
	if termW > 0 && termW < brokenW {
		w = 80
	}
	if termH > 0 && termH < brokenH {
		h = 24
	}
	return w, h
}
