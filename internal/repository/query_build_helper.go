package repository

import "strings"

type QueryParts struct {
	Where     []string // save sql conditions
	WhereArgs []any    // save info for conditions

	Set     []string // save row for SET
	SetArgs []any    // save info for SET
}

func NewQueryParts() *QueryParts {
	return &QueryParts{
		Where:     make([]string, 0),
		WhereArgs: make([]any, 0),
		Set:       make([]string, 0),
		SetArgs:   make([]any, 0),
	}
}

func (q *QueryParts) AddWhere(condition string, args ...any) {
	q.Where = append(q.Where, condition)
	q.WhereArgs = append(q.WhereArgs, args...)
}

func (q *QueryParts) AddSet(condition string, arg any) {
	q.Set = append(q.Set, condition)
	q.SetArgs = append(q.SetArgs, arg)
}

func (q *QueryParts) BuildWhere() string {
	if len(q.Where) == 0 {
		return ""
	}

	return " WHERE " + strings.Join(q.Where, " AND ")
}

func (q *QueryParts) BuildSet() string {
	if len(q.Set) == 0 {
		return ""
	}

	return " SET " + strings.Join(q.Set, ", ")
}
