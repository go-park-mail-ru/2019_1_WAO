package game

type Command struct {
	IdP       int     `json:"idP"`
	Direction string  `json:"direction"`
	Delay     float64 `json:"delay"`
}

// func () RemoveCommand(command *Command) error {

// }
