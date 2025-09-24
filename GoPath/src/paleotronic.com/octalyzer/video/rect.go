package video

type Rectangle struct {
	X1, Y1, X2, Y2 int
}

func (r *Rectangle) Contains(x, y int) bool {

	return (x >= r.X1 && x <= r.X2 && y >= r.Y1 && y <= r.Y2)

}
