package interfaces

type Countable interface {
	ImA() string
	Increment(n int)
	Decrement(n int)
	AdjustClock(speed int)
}
