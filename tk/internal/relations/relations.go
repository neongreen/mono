package relations

import ()

// RelationEdge represents an edge in the relation graph with OR-set semantics

// Source task UUID
// Relation type
// Destination task UUID
// Optional note
// Set of add tags (event_id -> exists)
// Set of removed tags (event_id -> exists)
// Last event ID that modified this edge
// Lamport timestamp when created

// RelationsGraph stores all relation edges with OR-set CRDT semantics

// Key: src:type:dst
