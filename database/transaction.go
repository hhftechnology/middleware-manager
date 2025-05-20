package database

import (
  "context"
  "database/sql"
  "fmt"
  "log"
  "time"
)

// TxFn represents a function that uses a transaction
type TxFn func(*sql.Tx) error

// WithTransaction wraps a function with a transaction
func (db *DB) WithTransaction(fn TxFn) error {
  tx, err := db.Begin()
  if err != nil {
    return fmt.Errorf("failed to begin transaction: %w", err)
  }
  
  defer func() {
    if p := recover(); p != nil {
      // Ensure rollback on panic
      log.Printf("Recovered from panic in transaction: %v", p)
      tx.Rollback()
      panic(p) // Re-throw panic after rollback
    }
  }()
  
  if err := fn(tx); err != nil {
    if rbErr := tx.Rollback(); rbErr != nil {
      log.Printf("Warning: Rollback failed: %v (original error: %v)", rbErr, err)
      return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
    }
    log.Printf("Transaction rolled back due to error: %v", err)
    return err
  }
  
  if err := tx.Commit(); err != nil {
    log.Printf("Error committing transaction: %v", err)
    return fmt.Errorf("commit failed: %w", err)
  }
  
  return nil
}

// WithTimeoutTransaction wraps a function with a transaction that has a timeout
func (db *DB) WithTimeoutTransaction(ctx context.Context, timeout time.Duration, fn TxFn) error {
  // Create a context with timeout
  ctx, cancel := context.WithTimeout(ctx, timeout)
  defer cancel()
  
  // Create a done channel to signal completion
  done := make(chan error, 1)
  
  // Run the transaction in a goroutine
  go func() {
    done <- db.WithTransaction(fn)
  }()
  
  // Wait for either context timeout or transaction completion
  select {
  case <-ctx.Done():
    // Context timed out
    return fmt.Errorf("transaction timed out after %v: %w", timeout, ctx.Err())
  case err := <-done:
    // Transaction completed
    return err
  }
}

// BatchTransaction executes multiple operations in a single transaction
// All operations must succeed or the transaction is rolled back
func (db *DB) BatchTransaction(operations []TxFn) error {
  return db.WithTransaction(func(tx *sql.Tx) error {
    for i, op := range operations {
      if err := op(tx); err != nil {
        return fmt.Errorf("operation %d failed: %w", i, err)
      }
    }
    return nil
  })
}

// UpdateInTransaction updates a record in a transaction
func (db *DB) UpdateInTransaction(table string, id string, updates map[string]interface{}) error {
  return db.WithTransaction(func(tx *sql.Tx) error {
    // Build the update statement
    query := fmt.Sprintf("UPDATE %s SET ", table)
    var params []interface{}
    
    i := 0
    for field, value := range updates {
      if i > 0 {
        query += ", "
      }
      query += field + " = ?"
      params = append(params, value)
      i++
    }
    
    // Add the WHERE clause and updated_at
    query += ", updated_at = ? WHERE id = ?"
    params = append(params, time.Now(), id)
    
    // Execute the update
    result, err := tx.Exec(query, params...)
    if err != nil {
      return fmt.Errorf("update failed: %w", err)
    }
    
    // Check if any rows were affected
    rowsAffected, err := result.RowsAffected()
    if err != nil {
      return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
      return fmt.Errorf("no rows affected, record with ID %s not found", id)
    }
    
    return nil
  })
}