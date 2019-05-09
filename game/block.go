package game

type Block struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Dy float64 `json:"dy"`
	w  float64
	h  float64
}
