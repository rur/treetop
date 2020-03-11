package treetop

import (
	"strconv"
	"testing"
)

func Test_viewQueue(t *testing.T) {
	// sanity check, queue works as expected
	queue := &viewQueue{}
	for i := 0; i < 10; i++ {
		queue.add(NewView(strconv.Itoa(i+1)+".html", Noop))
	}
	expect := []string{
		"1.html",
		"2.html",
		"3.html",
		"4.html",
		"5.html",
		"6.html",
		"7.html",
		"8.html",
		"9.html",
		"10.html",
	}
	var got []string
	for !queue.empty() {
		v, err := queue.next()
		if err != nil {
			t.Errorf("next returned an unexpected error %s", err)
			return
		}
		got = append(got, v.Template)
	}
	for len(got) < len(expect) {
		// pad got for diff purposes
		got = append(got, "")
	}
	for i := range got {
		if got[i] == "" {
			t.Errorf("Expecting template %s, got nothing", expect[i])
		} else if len(expect) <= i {
			t.Errorf("Unexpected template %s", got[i])
		} else if expect[i] != got[i] {
			t.Errorf("Expecting template %s, got %s", expect[i], got[i])
		}
	}
	if _, err := queue.next(); err != errEmptyViewQueue {
		t.Errorf("Expecting error '%s', got '%s'", errEmptyViewQueue, err)
	}
}
