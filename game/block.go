package game

type Block struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Vy float64
	w  float64
	h  float64
}
