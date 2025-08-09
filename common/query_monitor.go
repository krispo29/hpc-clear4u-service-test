package common

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
)

// QueryMonitor provides query performance monitoring and logging
type QueryMonitor struct {
	slowQueryThreshold time.Duration
	logSlowQueries     bool
	logAllQueries      bool
}

// QueryStats represents statistics for a database query
type QueryStats struct {
	Query     string        `json:"query"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	Error     error         `json:"error,omitempty"`
}

// NewQueryMonitor creates a new query monitor instance
func NewQueryMonitor(slowQueryThreshold time.Duration, logSlowQueries, logAllQueries bool) *QueryMonitor {
	return &QueryMonitor{
		slowQueryThreshold: slowQueryThreshold,
		logSlowQueries:     logSlowQueries,
		logAllQueries:      logAllQueries,
	}
}

// MonitorQuery wraps a query execution with performance monitoring
func (qm *QueryMonitor) MonitorQuery(ctx context.Context, queryName string, queryFunc func() error) error {
	start := time.Now()
	err := queryFunc()
	duration := time.Since(start)

	stats := QueryStats{
		Query:     queryName,
		Duration:  duration,
		Timestamp: start,
		Error:     err,
	}

	qm.logQuery(stats)
	return err
}

// MonitorPreparedQuery wraps a prepared statement execution with monitoring
func (qm *QueryMonitor) MonitorPreparedQuery(ctx context.Context, stmt *pg.Stmt, queryName string, queryFunc func(*pg.Stmt) error) error {
	start := time.Now()
	err := queryFunc(stmt)
	duration := time.Since(start)

	stats := QueryStats{
		Query:     queryName,
		Duration:  duration,
		Timestamp: start,
		Error:     err,
	}

	qm.logQuery(stats)
	return err
}

// MonitorTransaction wraps a transaction with monitoring
func (qm *QueryMonitor) MonitorTransaction(ctx context.Context, db *pg.DB, transactionName string, txFunc func(*pg.Tx) error) error {
	start := time.Now()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	err = txFunc(tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	duration := time.Since(start)

	stats := QueryStats{
		Query:     fmt.Sprintf("TRANSACTION: %s", transactionName),
		Duration:  duration,
		Timestamp: start,
		Error:     err,
	}

	qm.logQuery(stats)
	return err
}

// logQuery logs query statistics based on configuration
func (qm *QueryMonitor) logQuery(stats QueryStats) {
	shouldLog := qm.logAllQueries ||
		(qm.logSlowQueries && stats.Duration >= qm.slowQueryThreshold) ||
		stats.Error != nil

	if !shouldLog {
		return
	}

	logLevel := "INFO"
	if stats.Error != nil {
		logLevel = "ERROR"
	} else if stats.Duration >= qm.slowQueryThreshold {
		logLevel = "WARN"
	}

	logMessage := fmt.Sprintf("[%s] Query: %s | Duration: %v | Timestamp: %s",
		logLevel,
		qm.sanitizeQuery(stats.Query),
		stats.Duration,
		stats.Timestamp.Format(time.RFC3339),
	)

	if stats.Error != nil {
		logMessage += fmt.Sprintf(" | Error: %v", stats.Error)
	}

	log.Println(logMessage)
}

// sanitizeQuery removes sensitive information from query strings for logging
func (qm *QueryMonitor) sanitizeQuery(query string) string {
	// Remove potential sensitive data patterns
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")

	// Collapse multiple spaces
	for strings.Contains(query, "  ") {
		query = strings.ReplaceAll(query, "  ", " ")
	}

	// Truncate very long queries
	if len(query) > 200 {
		query = query[:200] + "..."
	}

	return strings.TrimSpace(query)
}

// GetSlowQueryThreshold returns the configured slow query threshold
func (qm *QueryMonitor) GetSlowQueryThreshold() time.Duration {
	return qm.slowQueryThreshold
}

// SetSlowQueryThreshold updates the slow query threshold
func (qm *QueryMonitor) SetSlowQueryThreshold(threshold time.Duration) {
	qm.slowQueryThreshold = threshold
}

// QueryOptimizer provides query optimization utilities
type QueryOptimizer struct {
	monitor *QueryMonitor
}

// NewQueryOptimizer creates a new query optimizer instance
func NewQueryOptimizer(monitor *QueryMonitor) *QueryOptimizer {
	return &QueryOptimizer{
		monitor: monitor,
	}
}

// OptimizeJoinQuery provides optimized JOIN query patterns
func (qo *QueryOptimizer) OptimizeJoinQuery(baseTable, joinTable, joinCondition string, selectFields []string, whereConditions []string) string {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s base
		INNER JOIN %s joined ON %s`,
		strings.Join(selectFields, ", "),
		baseTable,
		joinTable,
		joinCondition,
	)

	if len(whereConditions) > 0 {
		query += fmt.Sprintf("\nWHERE %s", strings.Join(whereConditions, " AND "))
	}

	return query
}

// OptimizeBatchInsert provides optimized batch insert patterns
func (qo *QueryOptimizer) OptimizeBatchInsert(table string, columns []string, batchSize int) string {
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	valueClause := fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))

	var valueClauses []string
	for i := 0; i < batchSize; i++ {
		valueClauses = append(valueClauses, valueClause)
	}

	return fmt.Sprintf(`
		INSERT INTO %s (%s) 
		VALUES %s`,
		table,
		strings.Join(columns, ", "),
		strings.Join(valueClauses, ", "),
	)
}

// OptimizeUpdateQuery provides optimized update query patterns
func (qo *QueryOptimizer) OptimizeUpdateQuery(table string, setColumns []string, whereConditions []string) string {
	setClauses := make([]string, len(setColumns))
	for i, col := range setColumns {
		setClauses[i] = fmt.Sprintf("%s = ?", col)
	}

	query := fmt.Sprintf(`
		UPDATE %s 
		SET %s`,
		table,
		strings.Join(setClauses, ", "),
	)

	if len(whereConditions) > 0 {
		query += fmt.Sprintf("\nWHERE %s", strings.Join(whereConditions, " AND "))
	}

	return query
}

// ValidateIndexUsage checks if a query is likely to use indexes effectively
func (qo *QueryOptimizer) ValidateIndexUsage(query string, indexedColumns []string) []string {
	var warnings []string
	queryLower := strings.ToLower(query)

	// Check for potential full table scans
	if strings.Contains(queryLower, "select") && !strings.Contains(queryLower, "where") {
		warnings = append(warnings, "Query may perform full table scan - consider adding WHERE clause")
	}

	// Check for LIKE patterns that can't use indexes
	if strings.Contains(queryLower, "like '%") {
		warnings = append(warnings, "Leading wildcard LIKE pattern cannot use indexes effectively")
	}

	// Check if indexed columns are being used in WHERE clauses
	hasIndexedColumn := false
	for _, col := range indexedColumns {
		if strings.Contains(queryLower, strings.ToLower(col)) {
			hasIndexedColumn = true
			break
		}
	}

	if !hasIndexedColumn && strings.Contains(queryLower, "where") {
		warnings = append(warnings, "Query WHERE clause may not utilize available indexes")
	}

	return warnings
}
