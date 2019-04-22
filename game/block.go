package game

type Block struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	w float32
	h float32
}

var blocks []*Block
