package game

type Block struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Dy float64 `json:"dy"`
	w  float64
	h  float64
}

func NewBlock(x float64, y float64, dy float64, width float64, height float64) *Block {
	return &Block{
		X:  x,
		Y:  y,
		Dy: dy,
		w:  width,
		h:  height,
	}
}
