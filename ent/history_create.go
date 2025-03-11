// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/katsukiniwa/kubernetes-sandbox/product/ent/history"
	"github.com/katsukiniwa/kubernetes-sandbox/product/ent/user"
)

// HistoryCreate is the builder for creating a History entity.
type HistoryCreate struct {
	config
	mutation *HistoryMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetAmount sets the "amount" field.
func (hc *HistoryCreate) SetAmount(i int) *HistoryCreate {
	hc.mutation.SetAmount(i)
	return hc
}

// SetNillableAmount sets the "amount" field if the given value is not nil.
func (hc *HistoryCreate) SetNillableAmount(i *int) *HistoryCreate {
	if i != nil {
		hc.SetAmount(*i)
	}
	return hc
}

// SetUserID sets the "user_id" field.
func (hc *HistoryCreate) SetUserID(i int) *HistoryCreate {
	hc.mutation.SetUserID(i)
	return hc
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (hc *HistoryCreate) SetNillableUserID(i *int) *HistoryCreate {
	if i != nil {
		hc.SetUserID(*i)
	}
	return hc
}

// SetID sets the "id" field.
func (hc *HistoryCreate) SetID(i int) *HistoryCreate {
	hc.mutation.SetID(i)
	return hc
}

// SetUser sets the "user" edge to the User entity.
func (hc *HistoryCreate) SetUser(u *User) *HistoryCreate {
	return hc.SetUserID(u.ID)
}

// Mutation returns the HistoryMutation object of the builder.
func (hc *HistoryCreate) Mutation() *HistoryMutation {
	return hc.mutation
}

// Save creates the History in the database.
func (hc *HistoryCreate) Save(ctx context.Context) (*History, error) {
	hc.defaults()
	return withHooks(ctx, hc.sqlSave, hc.mutation, hc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (hc *HistoryCreate) SaveX(ctx context.Context) *History {
	v, err := hc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (hc *HistoryCreate) Exec(ctx context.Context) error {
	_, err := hc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (hc *HistoryCreate) ExecX(ctx context.Context) {
	if err := hc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (hc *HistoryCreate) defaults() {
	if _, ok := hc.mutation.Amount(); !ok {
		v := history.DefaultAmount
		hc.mutation.SetAmount(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (hc *HistoryCreate) check() error {
	if _, ok := hc.mutation.Amount(); !ok {
		return &ValidationError{Name: "amount", err: errors.New(`ent: missing required field "History.amount"`)}
	}
	if v, ok := hc.mutation.ID(); ok {
		if err := history.IDValidator(v); err != nil {
			return &ValidationError{Name: "id", err: fmt.Errorf(`ent: validator failed for field "History.id": %w`, err)}
		}
	}
	return nil
}

func (hc *HistoryCreate) sqlSave(ctx context.Context) (*History, error) {
	if err := hc.check(); err != nil {
		return nil, err
	}
	_node, _spec := hc.createSpec()
	if err := sqlgraph.CreateNode(ctx, hc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != _node.ID {
		id := _spec.ID.Value.(int64)
		_node.ID = int(id)
	}
	hc.mutation.id = &_node.ID
	hc.mutation.done = true
	return _node, nil
}

func (hc *HistoryCreate) createSpec() (*History, *sqlgraph.CreateSpec) {
	var (
		_node = &History{config: hc.config}
		_spec = sqlgraph.NewCreateSpec(history.Table, sqlgraph.NewFieldSpec(history.FieldID, field.TypeInt))
	)
	_spec.OnConflict = hc.conflict
	if id, ok := hc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := hc.mutation.Amount(); ok {
		_spec.SetField(history.FieldAmount, field.TypeInt, value)
		_node.Amount = value
	}
	if nodes := hc.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   history.UserTable,
			Columns: []string{history.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.UserID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.History.Create().
//		SetAmount(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.HistoryUpsert) {
//			SetAmount(v+v).
//		}).
//		Exec(ctx)
func (hc *HistoryCreate) OnConflict(opts ...sql.ConflictOption) *HistoryUpsertOne {
	hc.conflict = opts
	return &HistoryUpsertOne{
		create: hc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.History.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (hc *HistoryCreate) OnConflictColumns(columns ...string) *HistoryUpsertOne {
	hc.conflict = append(hc.conflict, sql.ConflictColumns(columns...))
	return &HistoryUpsertOne{
		create: hc,
	}
}

type (
	// HistoryUpsertOne is the builder for "upsert"-ing
	//  one History node.
	HistoryUpsertOne struct {
		create *HistoryCreate
	}

	// HistoryUpsert is the "OnConflict" setter.
	HistoryUpsert struct {
		*sql.UpdateSet
	}
)

// SetAmount sets the "amount" field.
func (u *HistoryUpsert) SetAmount(v int) *HistoryUpsert {
	u.Set(history.FieldAmount, v)
	return u
}

// UpdateAmount sets the "amount" field to the value that was provided on create.
func (u *HistoryUpsert) UpdateAmount() *HistoryUpsert {
	u.SetExcluded(history.FieldAmount)
	return u
}

// AddAmount adds v to the "amount" field.
func (u *HistoryUpsert) AddAmount(v int) *HistoryUpsert {
	u.Add(history.FieldAmount, v)
	return u
}

// SetUserID sets the "user_id" field.
func (u *HistoryUpsert) SetUserID(v int) *HistoryUpsert {
	u.Set(history.FieldUserID, v)
	return u
}

// UpdateUserID sets the "user_id" field to the value that was provided on create.
func (u *HistoryUpsert) UpdateUserID() *HistoryUpsert {
	u.SetExcluded(history.FieldUserID)
	return u
}

// ClearUserID clears the value of the "user_id" field.
func (u *HistoryUpsert) ClearUserID() *HistoryUpsert {
	u.SetNull(history.FieldUserID)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.History.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(history.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *HistoryUpsertOne) UpdateNewValues() *HistoryUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(history.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.History.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *HistoryUpsertOne) Ignore() *HistoryUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *HistoryUpsertOne) DoNothing() *HistoryUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the HistoryCreate.OnConflict
// documentation for more info.
func (u *HistoryUpsertOne) Update(set func(*HistoryUpsert)) *HistoryUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&HistoryUpsert{UpdateSet: update})
	}))
	return u
}

// SetAmount sets the "amount" field.
func (u *HistoryUpsertOne) SetAmount(v int) *HistoryUpsertOne {
	return u.Update(func(s *HistoryUpsert) {
		s.SetAmount(v)
	})
}

// AddAmount adds v to the "amount" field.
func (u *HistoryUpsertOne) AddAmount(v int) *HistoryUpsertOne {
	return u.Update(func(s *HistoryUpsert) {
		s.AddAmount(v)
	})
}

// UpdateAmount sets the "amount" field to the value that was provided on create.
func (u *HistoryUpsertOne) UpdateAmount() *HistoryUpsertOne {
	return u.Update(func(s *HistoryUpsert) {
		s.UpdateAmount()
	})
}

// SetUserID sets the "user_id" field.
func (u *HistoryUpsertOne) SetUserID(v int) *HistoryUpsertOne {
	return u.Update(func(s *HistoryUpsert) {
		s.SetUserID(v)
	})
}

// UpdateUserID sets the "user_id" field to the value that was provided on create.
func (u *HistoryUpsertOne) UpdateUserID() *HistoryUpsertOne {
	return u.Update(func(s *HistoryUpsert) {
		s.UpdateUserID()
	})
}

// ClearUserID clears the value of the "user_id" field.
func (u *HistoryUpsertOne) ClearUserID() *HistoryUpsertOne {
	return u.Update(func(s *HistoryUpsert) {
		s.ClearUserID()
	})
}

// Exec executes the query.
func (u *HistoryUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for HistoryCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *HistoryUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *HistoryUpsertOne) ID(ctx context.Context) (id int, err error) {
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *HistoryUpsertOne) IDX(ctx context.Context) int {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// HistoryCreateBulk is the builder for creating many History entities in bulk.
type HistoryCreateBulk struct {
	config
	err      error
	builders []*HistoryCreate
	conflict []sql.ConflictOption
}

// Save creates the History entities in the database.
func (hcb *HistoryCreateBulk) Save(ctx context.Context) ([]*History, error) {
	if hcb.err != nil {
		return nil, hcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(hcb.builders))
	nodes := make([]*History, len(hcb.builders))
	mutators := make([]Mutator, len(hcb.builders))
	for i := range hcb.builders {
		func(i int, root context.Context) {
			builder := hcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*HistoryMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, hcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = hcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, hcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil && nodes[i].ID == 0 {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, hcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (hcb *HistoryCreateBulk) SaveX(ctx context.Context) []*History {
	v, err := hcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (hcb *HistoryCreateBulk) Exec(ctx context.Context) error {
	_, err := hcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (hcb *HistoryCreateBulk) ExecX(ctx context.Context) {
	if err := hcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.History.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.HistoryUpsert) {
//			SetAmount(v+v).
//		}).
//		Exec(ctx)
func (hcb *HistoryCreateBulk) OnConflict(opts ...sql.ConflictOption) *HistoryUpsertBulk {
	hcb.conflict = opts
	return &HistoryUpsertBulk{
		create: hcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.History.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (hcb *HistoryCreateBulk) OnConflictColumns(columns ...string) *HistoryUpsertBulk {
	hcb.conflict = append(hcb.conflict, sql.ConflictColumns(columns...))
	return &HistoryUpsertBulk{
		create: hcb,
	}
}

// HistoryUpsertBulk is the builder for "upsert"-ing
// a bulk of History nodes.
type HistoryUpsertBulk struct {
	create *HistoryCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.History.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(history.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *HistoryUpsertBulk) UpdateNewValues() *HistoryUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(history.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.History.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *HistoryUpsertBulk) Ignore() *HistoryUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *HistoryUpsertBulk) DoNothing() *HistoryUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the HistoryCreateBulk.OnConflict
// documentation for more info.
func (u *HistoryUpsertBulk) Update(set func(*HistoryUpsert)) *HistoryUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&HistoryUpsert{UpdateSet: update})
	}))
	return u
}

// SetAmount sets the "amount" field.
func (u *HistoryUpsertBulk) SetAmount(v int) *HistoryUpsertBulk {
	return u.Update(func(s *HistoryUpsert) {
		s.SetAmount(v)
	})
}

// AddAmount adds v to the "amount" field.
func (u *HistoryUpsertBulk) AddAmount(v int) *HistoryUpsertBulk {
	return u.Update(func(s *HistoryUpsert) {
		s.AddAmount(v)
	})
}

// UpdateAmount sets the "amount" field to the value that was provided on create.
func (u *HistoryUpsertBulk) UpdateAmount() *HistoryUpsertBulk {
	return u.Update(func(s *HistoryUpsert) {
		s.UpdateAmount()
	})
}

// SetUserID sets the "user_id" field.
func (u *HistoryUpsertBulk) SetUserID(v int) *HistoryUpsertBulk {
	return u.Update(func(s *HistoryUpsert) {
		s.SetUserID(v)
	})
}

// UpdateUserID sets the "user_id" field to the value that was provided on create.
func (u *HistoryUpsertBulk) UpdateUserID() *HistoryUpsertBulk {
	return u.Update(func(s *HistoryUpsert) {
		s.UpdateUserID()
	})
}

// ClearUserID clears the value of the "user_id" field.
func (u *HistoryUpsertBulk) ClearUserID() *HistoryUpsertBulk {
	return u.Update(func(s *HistoryUpsert) {
		s.ClearUserID()
	})
}

// Exec executes the query.
func (u *HistoryUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the HistoryCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for HistoryCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *HistoryUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
