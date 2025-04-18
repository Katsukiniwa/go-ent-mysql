// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/katsukiniwa/go-ent-mysql/product/ent/product"
)

// Product is the model entity for the Product schema.
type Product struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Stock holds the value of the "stock" field.
	Stock int `json:"stock,omitempty"`
	// Title holds the value of the "title" field.
	Title string `json:"title,omitempty"`
	// SaleStatus holds the value of the "sale_status" field.
	SaleStatus   product.SaleStatus `json:"sale_status,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Product) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case product.FieldID, product.FieldStock:
			values[i] = new(sql.NullInt64)
		case product.FieldTitle, product.FieldSaleStatus:
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Product fields.
func (pr *Product) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case product.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			pr.ID = int(value.Int64)
		case product.FieldStock:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field stock", values[i])
			} else if value.Valid {
				pr.Stock = int(value.Int64)
			}
		case product.FieldTitle:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[i])
			} else if value.Valid {
				pr.Title = value.String
			}
		case product.FieldSaleStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field sale_status", values[i])
			} else if value.Valid {
				pr.SaleStatus = product.SaleStatus(value.String)
			}
		default:
			pr.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Product.
// This includes values selected through modifiers, order, etc.
func (pr *Product) Value(name string) (ent.Value, error) {
	return pr.selectValues.Get(name)
}

// Update returns a builder for updating this Product.
// Note that you need to call Product.Unwrap() before calling this method if this Product
// was returned from a transaction, and the transaction was committed or rolled back.
func (pr *Product) Update() *ProductUpdateOne {
	return NewProductClient(pr.config).UpdateOne(pr)
}

// Unwrap unwraps the Product entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (pr *Product) Unwrap() *Product {
	_tx, ok := pr.config.driver.(*txDriver)
	if !ok {
		panic("ent: Product is not a transactional entity")
	}
	pr.config.driver = _tx.drv
	return pr
}

// String implements the fmt.Stringer.
func (pr *Product) String() string {
	var builder strings.Builder
	builder.WriteString("Product(")
	builder.WriteString(fmt.Sprintf("id=%v, ", pr.ID))
	builder.WriteString("stock=")
	builder.WriteString(fmt.Sprintf("%v", pr.Stock))
	builder.WriteString(", ")
	builder.WriteString("title=")
	builder.WriteString(pr.Title)
	builder.WriteString(", ")
	builder.WriteString("sale_status=")
	builder.WriteString(fmt.Sprintf("%v", pr.SaleStatus))
	builder.WriteByte(')')
	return builder.String()
}

// Products is a parsable slice of Product.
type Products []*Product
