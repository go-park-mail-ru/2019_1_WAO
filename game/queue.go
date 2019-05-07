package game

// NewQueue returns a new queue with the given initial size.
func NewQueue(size int) *Queue {
	return &Queue{
		commands: make([]*Command, size),
		size:     size,
	}
}

// Queue is a basic FIFO queue based on a circular list that resizes as needed.
type Queue struct {
	commands []*Command
	size     int
	head     int
	tail     int
	Count    int
}

// Push adds a Command to the queue.
func (q *Queue) Push(n *Command) {
	if q.head == q.tail && q.Count > 0 {
		commands := make([]*Command, len(q.commands)+q.size)
		copy(commands, q.commands[q.head:])
		copy(commands[len(q.commands)-q.head:], q.commands[:q.head])
		q.head = 0
		q.tail = len(q.commands)
		q.commands = commands
	}
	q.commands[q.tail] = n
	q.tail = (q.tail + 1) % len(q.commands)
	q.Count++
}

// Pop removes and returns a Command from the queue in first to last order.
func (q *Queue) Pop() *Command {
	if q.Count == 0 {
		return nil
	}
	Command := q.commands[q.head]
	q.head = (q.head + 1) % len(q.commands)
	q.Count--
	return Command
}
