package internal

import "strconv"

type PathID uint64

const rootPathID PathID = 1

type pathSegmentKind uint8

const (
	pathSegmentChild pathSegmentKind = iota + 1
	pathSegmentScope
)

type pathLookupKey struct {
	parent PathID
	kind   pathSegmentKind
	index  int
	name   string
}

type pathDebugEntry struct {
	parent   PathID
	kind     pathSegmentKind
	index    int
	name     string
	readable string
}

// MemoryKey is the allocation-light identity used for runtime memory,
// effects and animation state. Opaque keeps compatibility with legacy string
// keys; structured keys use Path/Namespace/Slot.
type MemoryKey struct {
	Path      PathID
	Namespace string
	Slot      int
	NoSlot    bool
	Opaque    string
}

func memoryKeyString(key string) MemoryKey {
	if key == "" {
		return MemoryKey{}
	}
	return MemoryKey{Opaque: key}
}

func (k MemoryKey) valid() bool {
	return k.Opaque != "" || (k.Path != 0 && k.Namespace != "")
}

func (r *Runtime) initPathTable() {
	if r == nil {
		return
	}
	if r.pathIDs == nil {
		r.pathIDs = make(map[pathLookupKey]PathID)
	}
	if r.pathDebug == nil {
		r.pathDebug = make(map[PathID]*pathDebugEntry)
	}
	r.pathDebug[rootPathID] = &pathDebugEntry{readable: "root"}
	if r.nextPathID <= rootPathID {
		r.nextPathID = rootPathID + 1
	}
}

func (r *Runtime) childPath(parent PathID, index int) PathID {
	if r == nil {
		return 0
	}
	r.initPathTable()
	key := pathLookupKey{parent: normalizePathID(parent), kind: pathSegmentChild, index: index}
	if id, ok := r.pathIDs[key]; ok {
		return id
	}
	id := r.nextPathID
	r.nextPathID++
	r.pathIDs[key] = id
	r.pathDebug[id] = &pathDebugEntry{parent: key.parent, kind: pathSegmentChild, index: index}
	return id
}

func (r *Runtime) scopePath(parent PathID, name string) PathID {
	if r == nil {
		return 0
	}
	r.initPathTable()
	key := pathLookupKey{parent: normalizePathID(parent), kind: pathSegmentScope, name: name}
	if id, ok := r.pathIDs[key]; ok {
		return id
	}
	id := r.nextPathID
	r.nextPathID++
	r.pathIDs[key] = id
	r.pathDebug[id] = &pathDebugEntry{parent: key.parent, kind: pathSegmentScope, name: name}
	return id
}

func normalizePathID(id PathID) PathID {
	if id == 0 {
		return rootPathID
	}
	return id
}

func joinPath(parent, segment string) string {
	if parent == "" {
		return segment
	}
	if segment == "" {
		return parent
	}
	return parent + "/" + segment
}

func (r *Runtime) DebugPath(id PathID) string {
	if r == nil {
		return ""
	}
	id = normalizePathID(id)
	r.initPathTable()
	entry := r.pathDebug[id]
	if entry == nil {
		return "path#" + strconv.FormatUint(uint64(id), 10)
	}
	if entry.readable != "" {
		return entry.readable
	}
	segment := entry.name
	if entry.kind == pathSegmentChild {
		segment = strconv.Itoa(entry.index)
	}
	entry.readable = joinPath(r.DebugPath(entry.parent), segment)
	return entry.readable
}

func (r *Runtime) DebugMemoryKey(key MemoryKey) string {
	if r == nil {
		return debugMemoryKeyWithoutRuntime(key, "")
	}
	if key.Opaque != "" {
		return key.Opaque
	}
	if !key.valid() {
		return ""
	}
	path := r.DebugPath(key.Path)
	if key.NoSlot {
		return joinPath(path, key.Namespace)
	}
	return joinPath(path, key.Namespace+":"+strconv.Itoa(key.Slot))
}

func debugMemoryKeyWithoutRuntime(key MemoryKey, path string) string {
	if key.Opaque != "" {
		return key.Opaque
	}
	if !key.valid() {
		return ""
	}
	if path == "" {
		path = "root"
	}
	if key.NoSlot {
		return joinPath(path, key.Namespace)
	}
	return joinPath(path, key.Namespace+":"+strconv.Itoa(key.Slot))
}
