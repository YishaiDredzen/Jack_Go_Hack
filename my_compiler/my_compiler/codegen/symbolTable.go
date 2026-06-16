package codegen

// Kind represents the variable kind in the symbol table
type Kind string

const (
	KindStatic   Kind = "static"
	KindField    Kind = "field"
	KindArg      Kind = "argument"
	KindVar      Kind = "local"
	KindNone     Kind = "none"
)

type Symbol struct {
	Name  string
	Type  string
	Kind  Kind
	Index int
}

type SymbolTable struct {
	classScope      map[string]*Symbol
	subroutineScope map[string]*Symbol
	counts          map[Kind]int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		classScope:      make(map[string]*Symbol),
		subroutineScope: make(map[string]*Symbol),
		counts:          map[Kind]int{KindStatic: 0, KindField: 0, KindArg: 0, KindVar: 0},
	}
}

// StartSubroutine resets subroutine-level scope and counters
func (st *SymbolTable) StartSubroutine() {
	st.subroutineScope = make(map[string]*Symbol)
	st.counts[KindArg] = 0
	st.counts[KindVar] = 0
}

// Define adds a new symbol to the appropriate scope
func (st *SymbolTable) Define(name, typ string, kind Kind) {
	sym := &Symbol{
		Name:  name,
		Type:  typ,
		Kind:  kind,
		Index: st.counts[kind],
	}
	st.counts[kind]++

	if kind == KindStatic || kind == KindField {
		st.classScope[name] = sym
	} else {
		st.subroutineScope[name] = sym
	}
}

// Lookup finds a symbol by name (subroutine scope first, then class scope)
func (st *SymbolTable) Lookup(name string) *Symbol {
	if sym, ok := st.subroutineScope[name]; ok {
		return sym
	}
	if sym, ok := st.classScope[name]; ok {
		return sym
	}
	return nil
}

// VarCount returns how many vars of a given kind are defined in current scope
func (st *SymbolTable) VarCount(kind Kind) int {
	return st.counts[kind]
}

// kindToSegment maps a symbol kind to its VM memory segment
func kindToSegment(kind Kind) string {
	switch kind {
	case KindStatic:
		return "static"
	case KindField:
		return "this"
	case KindArg:
		return "argument"
	case KindVar:
		return "local"
	}
	return ""
}
