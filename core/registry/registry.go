package registry

import (
	"sync"
)

// Registry manages both permanent (shared) and request-isolated key-value storage.
type Registry struct {
	global   map[string]interface{}
	globalMu sync.RWMutex
}

// NewRegistry creates a new Registry instance.
func NewRegistry() *Registry {
	return &Registry{
		global: make(map[string]interface{}),
	}
}

// SetGlobal sets a key-value pair in the permanent (shared) storage.
func (r *Registry) SetGlobal(key string, value interface{}) {
	r.globalMu.Lock()
	defer r.globalMu.Unlock()
	r.global[key] = value
}

// GetGlobal retrieves a value from the permanent (shared) storage.
func (r *Registry) GetGlobal(key string) (interface{}, bool) {
	r.globalMu.RLock()
	defer r.globalMu.RUnlock()
	val, ok := r.global[key]
	return val, ok
}

// DeleteGlobal removes a key from the permanent (shared) storage.
func (r *Registry) DeleteGlobal(key string) {
	r.globalMu.Lock()
	defer r.globalMu.Unlock()
	delete(r.global, key)
}

// RequestRegistry is a per-request registry (not shared).
type RequestRegistry struct {
	local map[string]interface{}
}

// NewRequestRegistry creates a new request-isolated registry.
func NewRequestRegistry() *RequestRegistry {
	return &RequestRegistry{
		local: make(map[string]interface{}),
	}
}

// Set sets a key-value pair in the request-isolated storage.
func (rr *RequestRegistry) Set(key string, value interface{}) {
	rr.local[key] = value
}

// Get retrieves a value from the request-isolated storage.
func (rr *RequestRegistry) Get(key string) (interface{}, bool) {
	val, ok := rr.local[key]
	return val, ok
}

// Delete removes a key from the request-isolated storage.
func (rr *RequestRegistry) Delete(key string) {
	delete(rr.local, key)
}

/*
USAGE EXAMPLES:

// --- Global (shared) registry ---
var GlobalRegistry = NewRegistry()

// Set a global value
GlobalRegistry.SetGlobal("site_name", "MySite")

// Get a global value
site, ok := GlobalRegistry.GetGlobal("site_name")

// Delete a global value
GlobalRegistry.DeleteGlobal("site_name")

// --- Request-isolated registry ---
reqReg := NewRequestRegistry()

// Set a request value
reqReg.Set("user_id", 123)

// Get a request value
userID, ok := reqReg.Get("user_id")

// Delete a request value
reqReg.Delete("user_id")
*/
