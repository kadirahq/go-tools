package monitor

import "testing"

func TestNew(t *testing.T) {
	if s := New("test"); s == nil {
		t.Fatal("s shouldn't be nil")
	}
}

func TestTrack(t *testing.T) {
	s := New("test")
	s.Register("foo", Counter)
	s.Track("foo", 100)

	m := s.vals["app.test:foo"]
	if m == nil {
		t.Fatal("m shouldn't be nil")
	}

	if m.Value() != 100 {
		t.Fatal("incorrect value")
	}
}
