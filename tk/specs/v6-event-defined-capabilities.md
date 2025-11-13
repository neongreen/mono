````markdown
# tk containers and event defined kinds — implementation spec v0

## 0 scope

this spec is for implementing primitive containers and event defined kinds in tk

it is written for claude code or similar to actually modify tk

out of scope  
- agent behaviour  
- higher level planning logic  
- fancy historical queries like “what did this look like last month”

in scope  
- primitive container behaviours in core  
- event types to define container kinds and other schema bits  
- sqlite tables to materialize current schema and container state  
- minimal cli surface

---

## 1 primitives

### 1 1 container primitives

hard coded primitive container types

```go
type ContainerPrimitive string

const (
    PrimitiveQueue ContainerPrimitive = "queue"
    PrimitiveStack ContainerPrimitive = "stack"
    PrimitiveGroup ContainerPrimitive = "group"
)
````

semantics

* queue

  * ordered fifo
  * push appends at tail
  * pop removes from head

* stack

  * ordered lifo
  * push appends at tail
  * pop removes from tail

* group

  * unordered set
  * add inserts if not present
  * remove deletes if present

these behaviours live in go code, not in schema

### 1 2 item ids and container ids

existing tk item ids like `tk-123` stay as they are

containers get ids in the same namespace style, but with prefixes

* queues `q-123`
* stacks `s-123`
* groups `g-123`

exact encoding can follow existing id generator in tk

---

## 2 persistent model

tk is already event sourced with sqlite, assume an `events` table with a json payload

we add **materialized tables** for schema and container state

### 2 1 schema tables

#### 2 1 1 container kinds

```sql
create table container_kinds (
    name           text primary key,        -- e g "sprint", "focus", "today"
    primitive      text not null,          -- "queue" | "stack" | "group"
    hint           text,                   -- optional human / llm hint
    deprecated     integer not null default 0, -- 0 = active, 1 = deprecated
    created_at     integer not null,       -- event index or unix time
    created_by     text not null           -- actor or agent id
);
```

notes

* kinds are defined by events (see section 3)
* no hard delete, only deprecate

#### 2 1 2 status, relation, metadata defs

if tk already has some tables for this, adapt instead of duplicating

otherwise create minimal ones

```sql
create table status_defs (
    name       text primary key,
    category   text not null,      -- e g "open", "in_progress", "closed"
    hint       text,
    deprecated integer not null default 0
);

create table relation_type_defs (
    name       text primary key,
    inverse    text not null,
    hint       text,
    deprecated integer not null default 0
);

create table metadata_field_defs (
    name       text primary key,
    type       text not null,      -- e g "string", "url", "sha"
    hint       text,
    deprecated integer not null default 0
);
```

all three are populated by events

### 2 2 containers

#### 2 2 1 container instances

```sql
create table containers (
    id          text primary key,        -- e g "q-1"
    primitive   text not null,          -- queue | stack | group
    kind        text not null,          -- foreign key to container_kinds(name)
    name        text not null,
    metadata    text,                   -- json blob, optional
    removed     integer not null default 0
);
```

notes

* `removed` is soft delete flag
* `metadata` is json text field, use existing json helper code if any

#### 2 2 2 container members

```sql
create table container_members (
    container_id  text not null,        -- fk to containers(id)
    item_id       text not null,        -- fk to items table or generic id
    position      integer,              -- ordering for queue / stack, null for group
    removed       integer not null default 0,
    primary key (container_id, item_id)
);
create index idx_container_members_container_pos
    on container_members(container_id, position);
```

semantics

* for queue and stack

  * `position` is a monotonically increasing integer
  * queue pop = smallest `position`
  * stack pop = largest `position`

* for group

  * `position` is null
  * membership is a set (unordered)

---

## 3 event types

assume tk has an `events` table with columns something like

```sql
create table events (
    id        integer primary key autoincrement,
    type      text not null,
    payload   text not null,  -- json
    created_at integer not null,
    actor     text not null
);
```

we add new `type` values and corresponding go structs

### 3 1 schema events

#### 3 1 1 define container kind

type `define_container_kind`

```json
{
  "name": "sprint",
  "primitive": "queue",
  "hint": "timeboxed work period"
}
```

go struct

```go
type DefineContainerKind struct {
    Name      string             `json:"name"`
    Primitive ContainerPrimitive `json:"primitive"`
    Hint      string             `json:"hint,omitempty"`
}
```

on replay

* insert into `container_kinds` if not present
* if present, do nothing or update hint only, up to you but avoid semantic changes

#### 3 1 2 deprecate container kind

type `deprecate_container_kind`

```json
{ "name": "old_sprint" }
```

on replay

* set `deprecated = 1` in `container_kinds` for that name

#### 3 1 3 define status

type `define_status`

```json
{
  "name": "todo",
  "category": "open",
  "hint": "not started yet"
}
```

similar structure for `define_relation_type`, `define_metadata_field`

no hard deletion events for these in v0, only deprecate

### 3 2 container instance events

#### 3 2 1 create container

type `create_container`

```json
{
  "id": "q-1",
  "primitive": "queue",
  "kind": "sprint",
  "name": "nov sprint",
  "metadata": { "project": "lovable" }
}
```

go struct

```go
type CreateContainer struct {
    ID        string             `json:"id"`
    Primitive ContainerPrimitive `json:"primitive"`
    Kind      string             `json:"kind"`
    Name      string             `json:"name"`
    Metadata  map[string]any     `json:"metadata,omitempty"`
}
```

on replay

* insert into `containers`
* assume `kind` already exists in `container_kinds`

#### 3 2 2 rename container

type `rename_container`

```json
{
  "id": "q-1",
  "name": "nov sprint v2"
}
```

on replay

* update `containers.name`

#### 3 2 3 update container metadata

type `update_container_metadata`

```json
{
  "id": "q-1",
  "metadata": { "project": "lovable", "priority": "high" }
}
```

on replay

* overwrite metadata json for now
* v0 does not require field level merges

#### 3 2 4 remove container

type `remove_container`

```json
{ "id": "q-1" }
```

on replay

* set `containers.removed = 1`
* optionally also set `container_members.removed = 1` for this container

### 3 3 membership events

**important:** `item_id` in all membership events must be a resolved task uid (like `tsk_01ABC...`),
never a display id (like `tk-123`)

commands must resolve task references before writing events
display logic must render task uids back to display ids for user output

positions are assigned by replay logic, not passed in explicitly by commands

#### 3 3 1 queue push

type `queue_push`

```json
{
  "container_id": "q-1",
  "item_id": "tk-123"
}
```

on replay

* find current max `position` for this container where `removed = 0`
* next position = max + 1 or 1 if none
* insert or update row in `container_members`

#### 3 3 2 queue pop

type `queue_pop`

```json
{
  "container_id": "q-1",
  "item_id": "tk-123"
}
```

note

* `item_id` in event is for audit
* actual selection of which item to pop should be done at command time, before event write

on replay

* set `removed = 1` for that `(container_id, item_id)`

#### 3 3 3 stack push / pop

same payload as queue push / pop

replay logic identical, behaviour difference only in **how the CLI selects item to pop**

* queue pop chooses smallest position
* stack pop chooses largest position

the event just records which item was popped

#### 3 3 4 group add / remove

type `group_add`

```json
{
  "container_id": "g-1",
  "item_id": "tsk_01ABC..."
}
```

note: `item_id` must be a resolved task uid, not a display id

on replay

* insert into `container_members` with `position = null`

type `group_remove`

* set `removed = 1` for that pair

---

## 4 replay and materialization

tk already has replay logic to build sqlite views from events, extend it

### 4 1 replay rules

for each event

* decode by `type`
* apply pure function to sqlite db inside a transaction

order

* kinds and schema events can appear anywhere before they are used
* code should assume well behaved event writer

### 4 2 rebuild strategy

v0 strategy

* on init or `tk migrate`

  * drop `container_*` and `*_defs` tables if exist
  * recreate
  * replay all events from start

if performance becomes an issue later, add snapshotting, out of scope here

---

## 5 cli surface

minimal commands so tk is usable

### 5 1 schema level

#### 5 1 1 define container kind

```bash
tk container-kind add queue sprint --hint "timeboxed project work"
tk schema add-kind group today --description "intended for today"
```

implementation

* validate primitive string
* write `define_container_kind` event

#### 5 1 2 list container kinds

```bash
tk container-kind list
```

implementation

* read from `container_kinds`
* skip deprecated unless `--all`

### 5 2 containers

#### 5 2 1 create containers

```bash
tk queue create sprint "nov sprint"
tk stack create return_to "return to later"
tk group create today "today"
```

implementation

* look up kind by name and primitive
* allocate new id
* write `create_container` event

#### 5 2 2 list containers

```bash
tk queue list
tk group list
```

simple selects from `containers` where `primitive = ?` and `removed = 0`

### 5 3 membership

**important:** items can exist in multiple containers simultaneously by design
this enables overlapping collections (see tk-13)

**task reference resolution:** before pushing/adding items, resolve the task reference
to ensure it exists, using tk's existing task resolution logic

#### 5 3 1 queue operations

```bash
tk queue push q-1 tk-123
tk queue pop q-1
tk queue list q-1
```

impl sketch

* `push`

  * validate q-1 is queue
  * **resolve tk-123 to ensure it exists**
  * write `queue_push` event

* `pop`

  * look up head item in `container_members`
  * if exists, write `queue_pop` with that `item_id`
  * print the popped item

* `list`

  * select items ordered by `position` where `removed = 0`

#### 5 3 2 stack operations

analogous to queue operations

```bash
tk stack push s-1 tk-123
tk stack pop s-1
tk stack list s-1
```

impl sketch

* `push`
  * validate s-1 is stack
  * **resolve tk-123 to ensure it exists**
  * write `stack_push` event

* `pop`
  * look up tail item (largest position) in `container_members`
  * if exists, write `stack_pop` with that `item_id`
  * print the popped item

only difference from queue is which `position` to choose on pop (max vs min)

#### 5 3 3 group operations

note: renamed from "cluster" to "group" for better ux

```bash
tk group add g-1 tk-123
tk group remove g-1 tk-123
tk group list g-1
```

impl

* `add`
  * **resolve tk-123 to ensure it exists**
  * writes `group_add`
* `remove` writes `group_remove`

---

## 6 schema export for agents and tools

add a command to dump schema snapshot

```bash
tk schema export --json
```

payload roughly

```json
{
  "statuses": [ ... ],
  "relation_types": [ ... ],
  "metadata_fields": [ ... ],
  "queue_kinds": [ ... ],
  "stack_kinds": [ ... ],
  "group_kinds": [ ... ]
}
```

implementation

* read from `*_defs` tables and `container_kinds`
* omit deprecated entries by default

---

## 7 migration considerations

* existing data

  * for v0 you can assume no existing containers
  * migration just creates new tables

* seeding defaults

  * optional step in migration to insert some initial events

    * `define_container_kind` for `focus`, `sprint`, `today`
    * some default statuses etc

seeding should be done via events, not direct table inserts

---

```
```
