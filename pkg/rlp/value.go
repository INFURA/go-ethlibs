package rlp

// Value represents a decoded RLP value, which is either a String or List.
type Value struct {
	// Only one of String or List is valid.  If String is "" then List is assumed valid.
	String string
	List   []Value
}
