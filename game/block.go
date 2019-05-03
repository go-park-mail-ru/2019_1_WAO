package game

type Block struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	w float64
	h float64
}

var blocks []*Block
