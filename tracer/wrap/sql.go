package wrap

import (
	"context"
	"database/sql/driver"
	"strings"

	"github.com/olaola-chat/rbp-library/tracer"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

type wrappedDriver struct {
	parent driver.Driver
}

type wrappedConn struct {
	parent driver.Conn
}

type wrappedTx struct {
	ctx    context.Context
	parent driver.Tx
	span   opentracing.Span
}

type wrappedStmt struct {
	span   opentracing.Span
	ctx    context.Context
	query  string
	parent driver.Stmt
}

type wrappedResult struct {
	ctx    context.Context
	parent driver.Result
}

type wrappedRows struct {
	ctx    context.Context
	parent driver.Rows
}

// Driver will wrap the passed SQL driver
func Driver(driver driver.Driver) driver.Driver {
	return wrappedDriver{parent: driver}
}

func (d wrappedDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.parent.Open(name)
	if err != nil {
		return nil, err
	}

	return wrappedConn{parent: conn}, nil
}

func (c wrappedConn) Prepare(query string) (driver.Stmt, error) {
	parent, err := c.parent.Prepare(query)
	if err != nil {
		return nil, err
	}

	return wrappedStmt{query: query, parent: parent}, nil
}

func (c wrappedConn) Close() error {
	return c.parent.Close()
}

func (c wrappedConn) Begin() (driver.Tx, error) {
	return nil, driver.ErrSkip
}

func (c wrappedConn) BeginTx(ctx context.Context, opts driver.TxOptions) (tx driver.Tx, err error) {
	span, _ := tracer.StartOpentracingSpan(ctx, "sql-tx")
	if connBeginTx, ok := c.parent.(driver.ConnBeginTx); ok {
		tx, err = connBeginTx.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return wrappedTx{ctx: ctx, parent: tx, span: span}, nil
	}
	return nil, driver.ErrSkip
}

func (c wrappedConn) PrepareContext(ctx context.Context, query string) (stmt driver.Stmt, err error) {
	if connPrepareCtx, ok := c.parent.(driver.ConnPrepareContext); ok {
		stmt, err := connPrepareCtx.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		span, _ := tracer.StartOpentracingSpan(ctx, "sql-query")
		if span != nil {
			span.SetTag("db.statement", query)
		}
		return wrappedStmt{
			ctx:    ctx,
			parent: stmt,
			span:   span,
		}, nil
	}

	return c.Prepare(query)
}

func (c wrappedConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return nil, driver.ErrSkip
}

func (c wrappedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Result, err error) {
	return nil, driver.ErrSkip
	/*
		fmt.Println("wrappedConn ExecContext start")
		if execContext, ok := c.parent.(driver.ExecerContext); ok {
			res, err := execContext.ExecContext(ctx, query, args)
			if err != nil {
				fmt.Println("ExecContext", err)
				return nil, err
			}

			return wrappedResult{Tracer: c.Tracer, ctx: ctx, parent: res}, nil
		}

		// Fallback implementation
		dargs, err := namedValueToValue(args)
		if err != nil {
			return nil, err
		}

		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		return c.Exec(query, dargs)
	*/
}

func (c wrappedConn) Ping(ctx context.Context) (err error) {
	if pinger, ok := c.parent.(driver.Pinger); ok {
		return pinger.Ping(ctx)
	}

	return nil
}

func (c wrappedConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return nil, driver.ErrSkip
}

func (c wrappedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (rows driver.Rows, err error) {
	return nil, driver.ErrSkip
	/*
		fmt.Println("wrappedConn QueryContext start")
		if queryerContext, ok := c.parent.(driver.QueryerContext); ok {
			rows, err := queryerContext.QueryContext(ctx, query, args)
			if err != nil {
				fmt.Println("QueryContext error", err)
				return nil, err
			}

			return wrappedRows{Tracer: c.Tracer, ctx: ctx, parent: rows}, nil
		}

		dargs, err := namedValueToValue(args)
		if err != nil {
			return nil, err
		}

		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		return c.Query(query, dargs)
	*/
}

func (t wrappedTx) Commit() (err error) {
	defer func() {
		if t.span != nil {
			if err != nil {
				t.span.LogFields(
					log.String("event", "error"),
					log.String("stack", trimError(err)),
				)
			}
			t.span.SetTag("result", "commit")
			t.span.Finish()
		}
	}()
	return t.parent.Commit()
}

func (t wrappedTx) Rollback() (err error) {
	defer func() {
		if t.span != nil {
			if err != nil {
				t.span.LogFields(
					log.String("event", "error"),
					log.String("stack", trimError(err)),
				)
			}
			t.span.SetTag("result", "rollback")
			t.span.Finish()
		}
	}()
	return t.parent.Rollback()
}

func (s wrappedStmt) Close() (err error) {
	defer func() {
		if s.span != nil {
			s.span.Finish()
		}
	}()
	return s.parent.Close()
}

func (s wrappedStmt) NumInput() int {
	return s.parent.NumInput()
}

func (s wrappedStmt) Exec(args []driver.Value) (res driver.Result, err error) {
	return nil, driver.ErrSkip
}

func (s wrappedStmt) Query(args []driver.Value) (rows driver.Rows, err error) {
	return nil, driver.ErrSkip
}

func (s wrappedStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (res driver.Result, err error) {
	if stmtExecContext, ok := s.parent.(driver.StmtExecContext); ok {
		res, err := stmtExecContext.ExecContext(ctx, args)
		if err != nil {
			if s.span != nil {
				s.span.LogFields(
					log.String("event", "error"),
					log.String("stack", trimError(err)),
				)
			}
			return nil, err
		}

		return wrappedResult{ctx: ctx, parent: res}, nil
	}

	// Fallback implementation
	dargs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return s.Exec(dargs)
}

func (s wrappedStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	if stmtQueryContext, ok := s.parent.(driver.StmtQueryContext); ok {
		rows, err := stmtQueryContext.QueryContext(ctx, args)
		if err != nil {
			return nil, err
		}

		return wrappedRows{ctx: ctx, parent: rows}, nil
	}

	dargs, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return s.Query(dargs)
}

func (r wrappedResult) LastInsertId() (id int64, err error) {
	return r.parent.LastInsertId()
}

func (r wrappedResult) RowsAffected() (num int64, err error) {
	return r.parent.RowsAffected()
}

func (r wrappedRows) Columns() []string {
	return r.parent.Columns()
}

func (r wrappedRows) Close() error {
	return r.parent.Close()
}

func (r wrappedRows) Next(dest []driver.Value) (err error) {
	return r.parent.Next(dest)
}

// namedValueToValue is a helper function copied from the database/sql package
func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("sql: driver does not support the use of Named Parameters")
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}

func trimError(err error) string {
	return strings.TrimSpace(strings.TrimPrefix(err.Error(), "Error"))
}
