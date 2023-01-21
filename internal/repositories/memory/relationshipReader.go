package memory

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/go-memdb"

	"github.com/Permify/permify/internal/repositories"
	"github.com/Permify/permify/internal/repositories/memory/snapshot"
	"github.com/Permify/permify/internal/repositories/memory/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	"github.com/Permify/permify/pkg/helper"
	"github.com/Permify/permify/pkg/logger"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// RelationshipReader - Structure for Relationship Reader
type RelationshipReader struct {
	database *db.Memory
	// logger
	logger logger.Interface
}

// NewRelationshipReader - Creates a new RelationshipReader
func NewRelationshipReader(database *db.Memory, logger logger.Interface) *RelationshipReader {
	return &RelationshipReader{
		database: database,
		logger:   logger,
	}
}

// QueryRelationships - Reads relation tuples from the repository.
func (r *RelationshipReader) QueryRelationships(ctx context.Context, tenantID string, filter *base.TupleFilter, _ string) (collection database.ITupleCollection, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	collection = database.NewTupleCollection()

	index, args := utils.GetIndexNameAndArgsByFilters(tenantID, filter)
	var it memdb.ResultIterator

	it, err = txn.Get(RelationTuplesTable, index, args...)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	fit := memdb.NewFilterIterator(it, utils.FilterQuery(filter))
	for obj := fit.Next(); obj != nil; obj = fit.Next() {
		t, ok := obj.(repositories.RelationTuple)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		collection.Add(t.ToTuple())
	}

	return collection, nil
}

// GetUniqueEntityIDsByEntityType - Gets all entity IDs for a given entity type (unique)
func (r *RelationshipReader) GetUniqueEntityIDsByEntityType(ctx context.Context, tenantID string, typ, _ string) (array []string, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var it memdb.ResultIterator
	it, err = txn.Get(RelationTuplesTable, "entity-type-index", tenantID, typ)
	if err != nil {
		return nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	var result []string
	for obj := it.Next(); obj != nil; obj = it.Next() {
		t, ok := obj.(repositories.RelationTuple)
		if !ok {
			return nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		result = append(result, t.EntityID)
	}

	return helper.RemoveDuplicate(result), nil
}

// HeadSnapshot - Reads the latest version of the snapshot from the repository.
func (r *RelationshipReader) HeadSnapshot(ctx context.Context, _ string) (token.SnapToken, error) {
	return snapshot.NewToken(time.Now()), nil
}
