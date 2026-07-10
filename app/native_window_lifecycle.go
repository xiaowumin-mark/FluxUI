package app

// nativeWindowTarget identifies one lifetime of a platform window handle.
// Native handles can be recycled by the OS, so the numeric handle alone is not
// sufficient to authorize a delayed operation.
type nativeWindowTarget struct {
	entry      *windowEntry
	handle     uintptr
	generation uint64
}

func (entry *windowEntry) withNativeWindowOperation(run func() bool) bool {
	if entry == nil || run == nil {
		return false
	}
	if entry.nativeCallbackDepth.Load() != 0 {
		return false
	}
	entry.nativeOperationMu.Lock()
	defer entry.nativeOperationMu.Unlock()
	if entry.nativeCallbackDepth.Load() != 0 {
		return false
	}
	return run()
}

func (entry *windowEntry) nativeWindowTargetSnapshot() nativeWindowTarget {
	if entry == nil {
		return nativeWindowTarget{}
	}
	entry.mu.RLock()
	if entry.nativeHandleReadyGeneration != entry.nativeHandleGeneration {
		entry.mu.RUnlock()
		return nativeWindowTarget{}
	}
	target := nativeWindowTarget{
		entry:      entry,
		handle:     entry.nativeHandle,
		generation: entry.nativeHandleGeneration,
	}
	entry.mu.RUnlock()
	return target
}

func (entry *windowEntry) nativeRawWindowTargetSnapshot() nativeWindowTarget {
	if entry == nil {
		return nativeWindowTarget{}
	}
	entry.mu.RLock()
	target := nativeWindowTarget{
		entry:      entry,
		handle:     entry.nativeHandle,
		generation: entry.nativeHandleGeneration,
	}
	entry.mu.RUnlock()
	return target
}

func (entry *windowEntry) setNativeWindowHandle(handle uintptr) nativeWindowTarget {
	target := entry.stageNativeWindowHandle(handle)
	entry.markNativeWindowTargetReady(target)
	return target
}

func (entry *windowEntry) stageNativeWindowHandle(handle uintptr) nativeWindowTarget {
	if entry == nil {
		return nativeWindowTarget{}
	}
	entry.mu.Lock()
	if entry.nativeHandle != handle {
		entry.nativeHandle = handle
		entry.nativeHandleGeneration++
		if entry.nativeHandleGeneration == 0 {
			entry.nativeHandleGeneration++
		}
		entry.nativeHandleReadyGeneration = 0
		entry.nativeWindowThreadID.Store(0)
		entry.nativeMaximizePointerDown = false
		entry.nativeMaximizeMouseDown = false
	}
	target := nativeWindowTarget{
		entry:      entry,
		handle:     entry.nativeHandle,
		generation: entry.nativeHandleGeneration,
	}
	entry.mu.Unlock()
	return target
}

func (entry *windowEntry) markNativeWindowTargetReady(target nativeWindowTarget) bool {
	if entry == nil || target.entry != entry || target.handle == 0 {
		return false
	}
	entry.mu.Lock()
	ready := target.matchesEntryLocked()
	if ready {
		entry.nativeHandleReadyGeneration = target.generation
	}
	entry.mu.Unlock()
	return ready
}

func (entry *windowEntry) invalidateNativeWindowHandle() {
	entry.setNativeWindowHandle(0)
}

func (entry *windowEntry) invalidateNativeWindowTarget(target nativeWindowTarget) bool {
	if entry == nil || target.entry != entry || target.handle == 0 {
		return false
	}
	entry.mu.Lock()
	if !target.matchesEntryLocked() {
		entry.mu.Unlock()
		return false
	}
	entry.nativeHandle = 0
	entry.nativeHandleGeneration++
	if entry.nativeHandleGeneration == 0 {
		entry.nativeHandleGeneration++
	}
	entry.nativeHandleReadyGeneration = 0
	entry.nativeWindowThreadID.Store(0)
	entry.nativeMaximizePointerDown = false
	entry.nativeMaximizeMouseDown = false
	entry.mu.Unlock()
	return true
}

func (target nativeWindowTarget) matchesEntryLocked() bool {
	return target.entry != nil &&
		target.handle != 0 &&
		target.entry.nativeHandle == target.handle &&
		target.entry.nativeHandleGeneration == target.generation
}

func (target nativeWindowTarget) valid() bool {
	entry := target.entry
	if entry == nil || target.handle == 0 || !entry.alive.Load() {
		return false
	}

	windowRegistryMu.RLock()
	current := windowRegistry[entry.id]
	if current != entry {
		windowRegistryMu.RUnlock()
		return false
	}
	entry.mu.RLock()
	valid := entry.alive.Load() && target.matchesEntryLocked()
	entry.mu.RUnlock()
	windowRegistryMu.RUnlock()
	return valid
}
