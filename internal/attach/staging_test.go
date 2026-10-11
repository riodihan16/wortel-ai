package attach

import (
	"errors"
	"fmt"
	"testing"
)

func stagingWithIDs(ids ...string) *Staging {
	s := NewStaging()
	i := 0
	s.newID = func() string {
		id := ids[i]
		i++
		return id
	}
	return s
}

func TestStagingAddAndGet(t *testing.T) {
	s := stagingWithIDs("a")
	m, err := s.Add(Meta{Name: "x.png", Mime: "image/png", Kind: KindImage}, []byte{1, 2, 3})
	if err != nil || m.ID != "a" || m.Size != 3 {
		t.Fatalf("%+v %v", m, err)
	}
	it, ok := s.Get("a")
	if !ok || it.Meta.Name != "x.png" || len(it.Data) != 3 || s.Len() != 1 {
		t.Fatalf("get: %+v %v", it, ok)
	}
}

func TestStagingRejectsFifthItem(t *testing.T) {
	s := stagingWithIDs("a", "b", "c", "d", "e")
	for i := 0; i < MaxPerMessage; i++ {
		if _, err := s.Add(Meta{Name: fmt.Sprintf("%d.png", i)}, []byte{1}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Add(Meta{Name: "5.png"}, []byte{1}); !errors.Is(err, ErrTooMany) {
		t.Fatalf("got %v", err)
	}
}

func TestStagingRejectsOverTotal(t *testing.T) {
	s := stagingWithIDs("a", "b", "c")
	seven := make([]byte, 7<<20)
	for i := 0; i < 2; i++ {
		if _, err := s.Add(Meta{Name: "x.pdf"}, seven); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Add(Meta{Name: "y.pdf"}, seven); !errors.Is(err, ErrTooBig) {
		t.Fatalf("got %v", err)
	}
}

func TestStagingSelectKeepsOrderAndRejectsUnknownOrDuplicate(t *testing.T) {
	s := stagingWithIDs("a", "b")
	_, _ = s.Add(Meta{Name: "a.png"}, []byte{1})
	_, _ = s.Add(Meta{Name: "b.png"}, []byte{2})
	items, err := s.Select([]string{"b", "a"})
	if err != nil || len(items) != 2 || items[0].Meta.Name != "b.png" || items[1].Meta.Name != "a.png" {
		t.Fatalf("%+v %v", items, err)
	}
	if _, err := s.Select([]string{"a", "zz"}); !errors.Is(err, ErrUnknown) {
		t.Fatalf("unknown: %v", err)
	}
	if _, err := s.Select([]string{"a", "a"}); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate: %v", err)
	}
	if s.Len() != 2 {
		t.Fatal("select must not remove")
	}
}

func TestStagingRemoveIdempotentAndFreesBudget(t *testing.T) {
	s := stagingWithIDs("a", "b")
	big := make([]byte, MaxMessageBytes)
	if _, err := s.Add(Meta{Name: "a.pdf"}, big); err != nil {
		t.Fatal(err)
	}
	s.Remove("a")
	s.Remove("a", "nope")
	if s.Len() != 0 {
		t.Fatal("not removed")
	}
	if _, err := s.Add(Meta{Name: "b.pdf"}, big); err != nil {
		t.Fatalf("budget not freed: %v", err)
	}
}

func TestStagingClear(t *testing.T) {
	s := stagingWithIDs("a")
	_, _ = s.Add(Meta{Name: "a.png"}, []byte{1})
	s.Clear()
	if _, ok := s.Get("a"); ok || s.Len() != 0 {
		t.Fatal("clear failed")
	}
}
