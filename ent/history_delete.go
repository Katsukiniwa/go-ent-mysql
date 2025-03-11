// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/katsukiniwa/kubernetes-sandbox/product/ent/history"
	"github.com/katsukiniwa/kubernetes-sandbox/product/ent/predicate"
)

// HistoryDelete is the builder for deleting a History entity.
type HistoryDelete struct {
	config
	hooks    []Hook
	mutation *HistoryMutation
}

// Where appends a list predicates to the HistoryDelete builder.
func (hd *HistoryDelete) Where(ps ...predicate.History) *HistoryDelete {
	hd.mutation.Where(ps...)
	return hd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (hd *HistoryDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, hd.sqlExec, hd.mutation, hd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (hd *HistoryDelete) ExecX(ctx context.Context) int {
	n, err := hd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (hd *HistoryDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(history.Table, sqlgraph.NewFieldSpec(history.FieldID, field.TypeInt))
	if ps := hd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, hd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	hd.mutation.done = true
	return affected, err
}

// HistoryDeleteOne is the builder for deleting a single History entity.
type HistoryDeleteOne struct {
	hd *HistoryDelete
}

// Where appends a list predicates to the HistoryDelete builder.
func (hdo *HistoryDeleteOne) Where(ps ...predicate.History) *HistoryDeleteOne {
	hdo.hd.mutation.Where(ps...)
	return hdo
}

// Exec executes the deletion query.
func (hdo *HistoryDeleteOne) Exec(ctx context.Context) error {
	n, err := hdo.hd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{history.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (hdo *HistoryDeleteOne) ExecX(ctx context.Context) {
	if err := hdo.Exec(ctx); err != nil {
		panic(err)
	}
}
