package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TxFn represents a function that uses a transaction
type TxFn func(*sql.Tx) error

// WithTransaction executes a function within a database transaction
// Automatically handles commit on success and rollback on error
func WithTransaction(db *sql.DB, fn TxFn) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
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
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// TransactionResponse represents the result of a transaction operation
type TransactionResponse struct {
	Success bool
	Data    interface{}
	Error   error
}

// WithTransactionAndResponse executes a transaction and returns a response suitable for gin
// This helper reduces boilerplate in handlers
func WithTransactionAndResponse(db *sql.DB, fn func(*sql.Tx) (interface{}, error)) TransactionResponse {
	var result interface{}
	err := WithTransaction(db, func(tx *sql.Tx) error {
		var txErr error
		result, txErr = fn(tx)
		return txErr
	})

	if err != nil {
		return TransactionResponse{
			Success: false,
			Error:   err,
		}
	}

	return TransactionResponse{
		Success: true,
		Data:    result,
	}
}

// ExecuteInTransaction is a gin middleware-style helper that executes a transaction
// and handles the HTTP response automatically
func ExecuteInTransaction(c *gin.Context, db *sql.DB, operation string, fn func(*sql.Tx) (interface{}, error)) bool {
	response := WithTransactionAndResponse(db, fn)

	if !response.Success {
		log.Printf("Error in %s: %v", operation, response.Error)
		ResponseWithError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to %s", operation))
		return false
	}

	return true
}

// ExecuteInTransactionWithResult executes a transaction and returns the result to the client
func ExecuteInTransactionWithResult(c *gin.Context, db *sql.DB, operation string, fn func(*sql.Tx) (interface{}, error)) {
	response := WithTransactionAndResponse(db, fn)

	if !response.Success {
		log.Printf("Error in %s: %v", operation, response.Error)
		ResponseWithError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to %s", operation))
		return
	}

	c.JSON(http.StatusOK, response.Data)
}

// DeleteInTransaction is a specialized helper for delete operations
func DeleteInTransaction(c *gin.Context, db *sql.DB, table string, id string, additionalDeletes ...func(*sql.Tx) error) {
	err := WithTransaction(db, func(tx *sql.Tx) error {
		// Execute any additional deletes first (e.g., related records)
		for _, deleteFn := range additionalDeletes {
			if err := deleteFn(tx); err != nil {
				return err
			}
		}

		// Delete the main record
		result, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", table), id)
		if err != nil {
			return fmt.Errorf("delete failed: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("no rows affected")
		}

		return nil
	})

	if err != nil {
		log.Printf("Error deleting from %s: %v", table, err)
		ResponseWithError(c, http.StatusInternalServerError, "Failed to delete record")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Record deleted successfully"})
}

// UpsertInTransaction performs an upsert (insert or update) operation
func UpsertInTransaction(db *sql.DB, insertQuery string, updateQuery string, insertArgs []interface{}, updateArgs []interface{}) error {
	return WithTransaction(db, func(tx *sql.Tx) error {
		// Try insert first
		result, err := tx.Exec(insertQuery, insertArgs...)
		if err == nil {
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				return nil
			}
		}

		// If insert failed or didn't affect rows, try update
		result, err = tx.Exec(updateQuery, updateArgs...)
		if err != nil {
			return fmt.Errorf("upsert failed: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("upsert didn't affect any rows")
		}

		return nil
	})
}
