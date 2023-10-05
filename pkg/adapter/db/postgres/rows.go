package postgres

import (
	"database/sql"
	"fmt"
)

type rowsAdapter struct {
	*sql.Rows
}

func (ra rowsAdapter) Close() {
	// returned error may be checked by calling the Err() method
	_ = ra.Rows.Close()
}

func (ra rowsAdapter) Values() ([]any, error) {
	names, err := ra.Columns()
	if err != nil {
		return nil, fmt.Errorf("column-names: %w", err)
	}
	vals := make([]any, len(names))
	valPtrs := make([]any, 0, len(names))
	for i := range vals {
		ptr := &vals[i]
		valPtrs = append(valPtrs, ptr)
	}
	err = ra.Scan(valPtrs...)
	return vals, err
}
