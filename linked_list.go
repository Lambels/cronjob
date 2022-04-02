package cronjob

import (
	"time"
)

// Node represents a node in the storage system.
type Node struct {
	// The id of the node in the storage system.
	Id int

	// The schedule the node is set to be activated on.
	Schedule Schedule

	// The job which needs to be ran.
	Job *Job

	// The ptr to the next node.
	next *Node
}

// linked list will point to the root node.
type linkedList struct {
	head *Node
	len  int
}

// NextCycle gets the duration of sleeping before activating.
func (l *linkedList) NextCycle(now time.Time) time.Duration {
	if l.head == nil {
		return -1
	}

	// the head node has the shortest duration.
	if v := l.head.Schedule.Calculate(now); v < 0 {
		return 0
	} else {
		return v
	}
}

// RunNow gets all the node that need to be ran now.
//
// This includes any nodes scheduled to run now or in the past.
func (l *linkedList) RunNow(now time.Time) (nodes []*Node) {
	ptr := l.head
	for i := 0; i < l.len; i++ {
		if ptr.Schedule.Calculate(now) <= 0 {
			nodes = append(nodes, ptr)
		}

		ptr = ptr.next
	}

	return
}

// ScheduleJob will add the node in the respective position
// based on the schedule.
//
// no-op if node is nil.
func (l *linkedList) AddNode(now time.Time, node *Node) {
	// return if pointer is nil.
	if node == nil {
		return
	}

	// if inserting a cyclic schedule, move to next activation time.
	if sched, ok := node.Schedule.(CyclicSchedule); ok {
		sched.MoveNextAvtivation(now)
	}

	// if head is nil add node as the head.
	if l.head == nil {
		l.len++
		l.head = node
		return
	}

	durInsertNode := node.Schedule.Calculate(now)

	ptr := l.head
	for i := 0; i < l.len; i++ {
		if durInsertNode <= ptr.Schedule.Calculate(now) {
			// this can only happen for the first node as all the other nodes are already checked
			// for in the next condition.
			l.len++
			l.addFirst(node)
			return

		} else if ptr.next == nil || durInsertNode <= ptr.next.Schedule.Calculate(now) {
			// add node after current node if next ptr is either nill (end of list)
			// or duration of the ptr to the next node is less then desired node.
			l.len++
			ptr.insertAfter(node)
			return
		}

		// advance in list.
		ptr = ptr.next
	}
}

// RemoveJob removes the job with the given id.
//
// no-op if node not found or list empty.
func (l *linkedList) RemoveNode(id int) {
	if l.len == 0 {
		return
	}

	ptr := l.head
	for i := 0; i < l.len; i++ {
		if ptr.Id == id {
			if i > 0 {
				prevNode := l.getAt(i - 1)
				prevNode.next = l.getAt(i).next
			} else {
				l.head = ptr.next
			}
			l.len--
			return
		}

		ptr = ptr.next
	}
}

// GetAll returns all the jobs in the storage system.
func (l *linkedList) GetAll() (jobs []*Node) {
	ptr := l.head
	for i := 0; i < l.len; i++ {
		jobs = append(jobs, ptr)
		ptr = ptr.next
	}

	return
}

// Clean removes the node (field) and re-calculates appropriate nodes.
func (l *linkedList) Clean(now time.Time, nodes []*Node) {
	for _, node := range nodes {
		switch node.Schedule.(type) {
		case CyclicSchedule:
			// remove the ran node node.
			l.RemoveNode(node.Id)

			// then re-add the cyclic node.
			l.AddNode(now, node)

		case Schedule:
			// remove nodes with constand schedule.
			l.RemoveNode(node.Id)
		}
	}
}

// addFirst changes the head of the linked list.
func (l *linkedList) addFirst(node *Node) {
	ptrTemp := l.head
	l.head = node
	node.next = ptrTemp
}

// insertAfter inserts the node (field) between the instance it is called upon and
// the node which is linked to the instance it is called upon.
func (n *Node) insertAfter(node *Node) {
	prtTemp := n.next
	n.next = node
	node.next = prtTemp
}

func (l *linkedList) getAt(pos int) *Node {
	if pos < 0 {
		return l.head
	}

	if pos > (l.len - 1) {
		return nil
	}

	ptr := l.head
	for i := 0; i < pos; i++ {
		ptr = ptr.next
	}
	return ptr
}
