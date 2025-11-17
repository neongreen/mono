// Package v6_to_v7 implements the migration from database version 6 to 7.
//
// This migration adds support for custom item kinds (task, decision, resource, etc.)
// by introducing the item_kinds table and adding an item_kind column to the tasks table.
//
// All existing tasks are automatically assigned the 'task' kind during migration.
package v6_to_v7
