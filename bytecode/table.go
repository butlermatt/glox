package bytecode

import "fmt"

const tableMaxLoad = 0.75

type Table struct {
	count   int
	cap     int
	Entries []Entry
}
func NewTable() *Table {
	return &Table{}
}

func (t *Table) Free() {
	t.count = 0
	t.cap = 0
	t.Entries = []Entry{}
}

func (t *Table) Get(key *StringObj) (Value, bool) {
	if t.count == 0 {
		return nil, false
	}

	e := findEntry(t.Entries, key)
	if e.Key == nil {
		return nil, false
	}

	return e.Value, true
}

func (t *Table) Set(key *StringObj, value Value) bool {
	if float64(t.count + 1) > float64(t.cap) * tableMaxLoad {
		var c int
		if t.cap < 8 {
			c = 8
		} else {
			c = t.cap * 2
		}
		t.adjustCapacity(c)
	}

	entry := findEntry(t.Entries, key)

	isNew := entry.Key == nil
	if isNew && entry.Value == nil {
		t.count++
	}

	entry.Key = key
	entry.Value = value
	return isNew
}

func (t *Table) Delete(key *StringObj) bool {
	if t.count == 0 {
		return false
	}

	e := findEntry(t.Entries, key)
	if e.Key == nil {
		return false
	}

	// place a tombstone
	e.Key = nil
	e.Value = BoolAsValue(true)

	return true
}

func (t *Table) AddAll(from *Table) {
	for i := 0; i < from.cap; i++ {
		e := from.Entries[i]
		if e.Key != nil {
			t.Set(e.Key, e.Value)
		}
	}
}

func (t *Table) FindString(value string, hash uint32) *StringObj {
	if t.count == 0 {
		return nil
	}

	fmt.Printf("Looking for: %q\n", value)

	index := hash % uint32(t.cap)
	for {
		e := t.Entries[index]

		if e.Key == nil && e.Value == nil {
			// Stop if we find an empty non-tombstone entry.
			return nil
		}
		if hash == e.Key.Hash && value == e.Key.Value {
			return e.Key
		}
	}
}

func (t *Table) adjustCapacity(cap int) {
	entries := make([]Entry, cap)

	t.count = 0
	// For loop to set to nil. Not needed in go
	for i := 0; i < len(t.Entries); i++ {
		e := t.Entries[i]
		if e.Key == nil {
			continue
		}

		dest := findEntry(entries, e.Key)
		dest.Key = e.Key
		dest.Value = e.Value
		t.count++
	}

	t.Entries = entries
	t.cap = cap
}

func findEntry(entries []Entry, key *StringObj) Entry {
	ind := key.Hash % uint32(len(entries))
	var tombstone Entry

	for {
		e := entries[ind]
		// Found match
		if e.Key == key { return e }

		if e.Key == nil {
			// Empty Entry
			if e.Value == nil {
				if tombstone.Value == True { // Have a tombstone. Return that
					return tombstone
				}
				// No tombstone from previous iteration. Just return empty.
				return e
			} else if tombstone.Value == nil {
				tombstone = e // assign if there isn't an existing tombstone
			}
		}

		// non-empty entry, not a match. Collision.
		ind = (ind + 1) % uint32(len(entries))
	}
}


type Entry struct {
	Key   *StringObj
	Value Value
}
