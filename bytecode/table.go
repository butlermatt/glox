package bytecode

import "fmt"

const tableMaxLoad = 0.75

type Table struct {
	count   int
	cap     int
	Entries []*Entry
}

func NewTable() *Table {
	return &Table{}
}

func (t *Table) Free() {
	t.count = 0
	t.cap = 0
	t.Entries = []*Entry{}
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

func (t *Table) dumpTable(where string) {
	fmt.Println(where)
	fmt.Printf("Table count: %d, Cap: %d Entries:\n", t.count, t.cap)
	for i := 0; i < len(t.Entries); i++ {
		e := t.Entries[i]
		var key string
		if e.Key == nil {
			key = "<nil>"
		} else {
			key = e.Key.Value
		}

		fmt.Printf("[ %v : ", key)
		PrintValue(e.Value)
		fmt.Printf(" ] ")

	}
	fmt.Println()
}

func (t *Table) Set(key *StringObj, value Value) bool {
	if float64(t.count+1) > float64(t.cap)*tableMaxLoad {
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
	if isNew && entry.Value == Nil {
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

	index := hash % uint32(t.cap)
	for {
		e := t.Entries[index]

		if e.Key == nil && e.Value == Nil {
			// Stop if we find an empty non-tombstone entry.
			return nil
		}
		if hash == e.Key.Hash && value == e.Key.Value {
			return e.Key
		}

		index = (index + 1) % uint32(t.cap)
	}
}

func (t *Table) adjustCapacity(cap int) {
	entries := make([]*Entry, cap)

	t.count = 0
	// For loop to set to nil.
	for i := 0; i < len(entries); i++ {
		entries[i] = &Entry{Value: Nil}
	}

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

func findEntry(entries []*Entry, key *StringObj) *Entry {
	ind := key.Hash % uint32(len(entries))
	var tombstone *Entry

	for {
		e := entries[ind]

		if e.Key == nil {
			if e.Value == Nil { // Empty Entry
				if tombstone != nil {
					return tombstone
				}
				return e
			} else { // Empty Key, not Value. It's a tombstone
				if tombstone == nil {
					tombstone = e
				}
			}
		}

		if e.Key == key {
			return e
		}

		ind = (ind + 1) % uint32(len(entries))
	}
}

type Entry struct {
	Key   *StringObj
	Value Value
}
