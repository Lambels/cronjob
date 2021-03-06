package cronjob

import (
	"reflect"
	"testing"
	"time"
)

func TestNextCycle(t *testing.T) {
	t.Run("Test Positive Duration", func(t *testing.T) {
		now := time.Now()

		l := newWithNode(now, &Node{
			Schedule: In(now, 5*time.Second),
		})

		if got, want := l.NextCycle(now), 5*time.Second; got != want {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})

	t.Run("Test Negative Duration", func(t *testing.T) {
		now := time.Now()

		l := newWithNode(now.Add(6*time.Second), &Node{
			Schedule: In(now, 5*time.Second),
		})

		if got, want := l.NextCycle(now.Add(6*time.Second)), time.Duration(0); got != want {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})
}

func TestRunNow(t *testing.T) {
	l := &linkedList{}

	now := time.Now()

	n1 := &Node{
		Schedule: In(now, 5*time.Second),
	}
	n2 := &Node{
		Schedule: In(now, 3*time.Second),
	}

	l.AddNode(now.Add(5*time.Second), n1)
	l.AddNode(now.Add(5*time.Second), n2)

	if got, want := l.RunNow(now.Add(5*time.Second)), []*Node{n2, n1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v want: %v", got, want)
	}
}

func TestAddNode(t *testing.T) {
	t.Run("Test Cyclic Node", func(t *testing.T) {
		now := time.Now()

		n1 := &Node{
			Schedule: In(now, 5*time.Second),
		}
		n2 := &Node{
			Schedule: Every(2 * time.Second),
		}

		l := newWithNode(now, n1)
		l.AddNode(now, n2)

		if got, want := l.GetAll(), []*Node{n2, n1}; !reflect.DeepEqual(got, want) {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})

	t.Run("Test Constant Node", func(t *testing.T) {
		now := time.Now()

		n1 := &Node{
			Schedule: In(now, 5*time.Second),
		}
		n2 := &Node{
			Schedule: In(now, 2*time.Second),
		}

		l := newWithNode(now, n1)
		l.AddNode(now, n2)

		if got, want := l.GetAll(), []*Node{n2, n1}; !reflect.DeepEqual(got, want) {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})

	t.Run("Test No Head Node", func(t *testing.T) {
		now := time.Now()

		n1 := &Node{
			Schedule: In(now, 5*time.Second),
		}

		l := newWithNode(now, n1)

		if got, want := l.head, n1; got != want {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})

	t.Run("Test Change Head Node", func(t *testing.T) {
		now := time.Now()

		l := newWithNode(now, &Node{
			Schedule: In(now, 5*time.Second),
		})

		n1 := &Node{
			Schedule: In(now, 2*time.Second),
		}

		l.AddNode(now, n1)

		if got, want := l.head, n1; got != want {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})
}

func TestRemoveNode(t *testing.T) {
	c := New()
	id := c.AddFunc(func() error { return nil }, In(time.Now(), 5*time.Second))
	c.scheduler.RemoveNode(id)

	if got, want := c.scheduler.GetAll(), []*Node{}; reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v want: %v", got, want)
	}
}

func TestGetAll(t *testing.T) {
	t.Run("Test Empty", func(t *testing.T) {
		l := &linkedList{}

		if got, want := len(l.GetAll()), 0; got != want {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})

	t.Run("Test Content", func(t *testing.T) {
		n1 := &Node{
			Schedule: In(time.Now(), 2*time.Second),
		}

		l := newWithNode(time.Now(), n1)

		if got, want := l.GetAll(), []*Node{n1}; !reflect.DeepEqual(got, want) {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})
}

func TestClean(t *testing.T) {
	t.Run("Test Cyclic Node", func(t *testing.T) {
		now := time.Now()

		l := newWithNode(now, &Node{Id: 1, Schedule: Every(5 * time.Second)})
		l.Clean(now.Add(6*time.Second), l.GetAll())

		// verify that it gets re added.
		if got, want := len(l.GetAll()), 1; got != want {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})

	t.Run("Test Constant Node", func(t *testing.T) {
		now := time.Now()

		l := newWithNode(now, &Node{Schedule: In(now, 5*time.Second)})
		l.Clean(now.Add(6*time.Second), l.GetAll())

		// verify that it gets removed.
		if got, want := len(l.GetAll()), 0; got != want {
			t.Fatalf("got: %v want: %v", got, want)
		}
	})
}

func newWithNode(now time.Time, node *Node) *linkedList {
	l := &linkedList{}

	l.AddNode(now, node)
	return l
}
