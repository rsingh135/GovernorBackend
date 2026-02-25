package models

import (
	"time"

	"github.com/google/uuid"
)

// TransactionFilters holds filter parameters for transaction queries.
type TransactionFilters struct {
	AgentID  *uuid.UUID
	Status   *string
	FromDate *time.Time
	ToDate   *time.Time
}

// AgentFilters holds filter parameters for agent queries.
type AgentFilters struct {
	UserID *uuid.UUID
	Status *string
}
