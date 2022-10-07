package async_test

import (
	"testing"

	"github.com/abiiranathan/gowrap/async"
)

func TestMapInit(t *testing.T) {
	t.Parallel()
	m := async.NewAsyncMap[uint, int]()
	if m == nil {
		t.Error("AsyncMapInit failed")
	}
}

func TestMapGet(t *testing.T) {
	t.Parallel()
	m := async.NewAsyncMap[uint, int]()
	m.Set(1, 1)
	v, ok := m.Get(1)
	if !ok {
		t.Error("AsyncMapGet failed")
	}

	if v != 1 {
		t.Error("AsyncMapGet failed")
	}
}

func TestMapDelete(t *testing.T) {
	t.Parallel()
	m := async.NewAsyncMap[uint, int]()
	m.Set(1, 1)
	m.Delete(1)
	v, ok := m.Get(1)
	if ok {
		t.Error("AsyncMapDelete failed")
	}
	if v != 0 {
		t.Error("AsyncMapDelete failed")
	}
}

func TestMapClear(t *testing.T) {
	t.Parallel()
	m := async.NewAsyncMap[uint, int]()
	m.Set(1, 1)
	m.Clear()
	v, ok := m.Get(1)
	if ok {
		t.Error("AsyncMapClear failed")
	}
	if v != 0 {
		t.Error("AsyncMapClear failed")
	}
}

func TestMapSet(t *testing.T) {
	t.Parallel()
	m := async.NewAsyncMap[uint, int]()
	m.Set(1, 1)
	v, ok := m.Get(1)
	if !ok {
		t.Error("AsyncMapSet failed")
	}
	if v != 1 {
		t.Error("AsyncMapSet failed")
	}
}

func TestMapContains(t *testing.T) {
	t.Parallel()
	m := async.NewAsyncMap[uint, int]()
	m.Set(1, 1)
	if !m.Contains(1) {
		t.Error("AsyncMapContains failed")
	}
}

func TestMapLen(t *testing.T) {
	t.Parallel()
	m := async.NewAsyncMap[uint, int]()
	m.Set(1, 1)
	if m.Len() != 1 {
		t.Error("AsyncMapLen failed")
	}
}

func TestMapKeys(t *testing.T) {
	t.Parallel()
	m := async.NewAsyncMap[uint, int]()
	m.Set(1, 1)
	m.Set(2, 2)
	m.Set(3, 3)
	keys := m.Keys()
	if len(keys) != 3 {
		t.Error("AsyncMapKeys failed")
	}
}

func TestMapValues(t *testing.T) {
	t.Parallel()
	m := async.NewAsyncMap[uint, int]()
	m.Set(1, 1)
	m.Set(2, 2)
	m.Set(3, 3)
	values := m.Values()
	if len(values) != 3 {
		t.Error("AsyncMapValues failed")
	}
}

func TestMapIsEmpty(t *testing.T) {

	t.Parallel()
	m := async.NewAsyncMap[uint, int]()
	if !m.IsEmpty() {
		t.Error("AsyncMapIsEmpty failed")
	}

	m.Set(1, 1)
	if m.IsEmpty() {
		t.Error("AsyncMapIsEmpty failed")
	}
}
