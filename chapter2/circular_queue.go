package main

type Queue struct {
	elements    []int
	front, rear int
	len         int
}

func NewQueue(size int) *Queue {
	return &Queue{
		elements: make([]int, size),
		front:    0,
		rear:     -1,
		len:      size,
	}
}

func (q *Queue) Enqueue(e int) bool {
	if q.len == len(q.elements) {
		return false
	}

	q.rear = (q.rear + 1) % len(q.elements)
	q.elements[q.rear] = e
	q.len++
	return true
}

func (q *Queue) Dequeue() (int, bool) {
	if q.len == 0 {
		return 0, false
	}

	e := q.elements[q.front]
	q.front = (q.front + 1) % len(q.elements)
	q.len--
	return e, true
}
