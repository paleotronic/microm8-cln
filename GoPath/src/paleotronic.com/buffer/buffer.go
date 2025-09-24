package buffer

import "sync"

type RingBuffer struct {
    b []interface{}
    head, tail, cap int
    grow bool
    block bool
    m sync.Mutex
}
 
func (q *RingBuffer) Push(x interface{}) bool {
	
	//q.m.Lock()
	//defer q.m.Unlock()
	
    switch {
    // buffer full. reallocate if grow is set.
    case q.tail < 0:
		if q.grow {
			next := len(q.b)
			bigger := make([]interface{}, 2*next)
			copy(bigger[copy(bigger, q.b[q.head:]):], q.b[:q.head])
			bigger[next] = x
			q.b, q.head, q.tail = bigger, 0, next+1
			q.cap = len(bigger)
		} else {
			return false
		}
    // zero object. make initial allocation.
    case len(q.b) == 0:
        q.b, q.head, q.tail = make([]interface{}, q.cap), 0 ,1
        q.b[0] = x
    // normal case
    default:
        q.b[q.tail] = x
        q.tail++
        if q.tail == len(q.b) {
            q.tail = 0
        }
        if q.tail == q.head {
            q.tail = -1
        }
    }
    
    return true
}
 
func (q *RingBuffer) Pop() (interface{}, bool) {
	
	//q.m.Lock()
	//defer q.m.Unlock()
	
    if q.head == q.tail {
        return nil, false
    }
    r := q.b[q.head]
    if q.tail == -1 {
        q.tail = q.head
    }
    q.head++
    if q.head == len(q.b) {
        q.head = 0
    }
    return r, true
}
 
func (q *RingBuffer) Empty() bool {
	//q.m.Lock()
	//defer q.m.Unlock()
	
    return q.head == q.tail
}
 
func NewRingBuffer(cap int, grow bool) *RingBuffer {
	return &RingBuffer{
		b: make([]interface{}, cap),
		head: 0,
		tail: 0,
		cap: cap,
		grow: grow,
	}
}

