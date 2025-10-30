package main

// GetCachedReducerWithConfig returns a cached reducer or builds a new one if needed.
// The cache is invalidated when new events are inserted.
// This significantly improves performance for operations that need to query task state.
//
// Note: The cache uses pointer identity for config comparison. This is safe because:
// - Each command typically loads config once and reuses the same instance
// - The DB instance is scoped to a single command execution
// - Cache is invalidated on any event insertion
// If config pointer doesn't match, we rebuild the reducer (safe but may miss some cache hits).
func (d *DB) GetCachedReducerWithConfig(config *Config) (*Reducer, error) {

	if d.reducerCache != nil && d.reducerConfig == config {
		return d.reducerCache, nil
	}

	events, err := d.GetEvents()
	if err != nil {
		return nil, err
	}

	reducer, err := BuildFromEventsWithConfig(events, config)
	if err != nil {
		return nil, err
	}

	d.reducerCache = reducer
	d.reducerConfig = config

	return reducer, nil
}
