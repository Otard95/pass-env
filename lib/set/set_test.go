package set

import (
	"testing"
)

func TestContains(t *testing.T) {
	s := Set[string]{}
	s.Add("foo")

	if !s.Contains("foo") {
		t.Error("Expected set to contain 'foo'")
	}

	if s.Contains("bar") {
		t.Error("Expected set to not contain 'bar'")
	}
}

func TestAdd(t *testing.T) {
	s := Set[string]{}

	if !s.Add("foo") {
		t.Error("Expected Add to return true for new element")
	}

	if s.Add("foo") {
		t.Error("Expected Add to return false for duplicate element")
	}

	if len(s) != 1 {
		t.Errorf("Expected set length 1, got %d", len(s))
	}
}

func TestRemove(t *testing.T) {
	s := Set[string]{}
	s.Add("foo")
	s.Add("bar")

	s.Remove("foo")

	if s.Contains("foo") {
		t.Error("Expected 'foo' to be removed")
	}

	if !s.Contains("bar") {
		t.Error("Expected 'bar' to still exist")
	}

	if len(s) != 1 {
		t.Errorf("Expected set length 1, got %d", len(s))
	}
}

func TestItems(t *testing.T) {
	s := Set[string]{}
	s.Add("foo")
	s.Add("bar")
	s.Add("baz")

	items := s.Items()

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Check all items are present (order doesn't matter)
	found := make(map[string]bool)
	for _, item := range items {
		found[item] = true
	}

	if !found["foo"] || !found["bar"] || !found["baz"] {
		t.Error("Expected items to contain foo, bar, and baz")
	}
}

func TestItemsEmpty(t *testing.T) {
	s := Set[string]{}
	items := s.Items()

	if len(items) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(items))
	}
}

func TestMerge(t *testing.T) {
	s1 := Set[string]{}
	s1.Add("foo")
	s1.Add("bar")

	s2 := Set[string]{}
	s2.Add("bar")
	s2.Add("baz")
	s2.Add("qux")

	s1.Merge(s2)

	if len(s1) != 4 {
		t.Errorf("Expected merged set to have 4 elements, got %d", len(s1))
	}

	expected := []string{"foo", "bar", "baz", "qux"}
	for _, item := range expected {
		if !s1.Contains(item) {
			t.Errorf("Expected merged set to contain '%s'", item)
		}
	}
}

func TestMergeEmpty(t *testing.T) {
	s1 := Set[string]{}
	s1.Add("foo")

	s2 := Set[string]{}

	s1.Merge(s2)

	if len(s1) != 1 {
		t.Errorf("Expected set to still have 1 element, got %d", len(s1))
	}

	if !s1.Contains("foo") {
		t.Error("Expected original element to remain")
	}
}

func TestMergeIntoEmpty(t *testing.T) {
	s1 := Set[string]{}

	s2 := Set[string]{}
	s2.Add("foo")
	s2.Add("bar")

	s1.Merge(s2)

	if len(s1) != 2 {
		t.Errorf("Expected set to have 2 elements, got %d", len(s1))
	}

	if !s1.Contains("foo") || !s1.Contains("bar") {
		t.Error("Expected merged elements to be present")
	}
}
