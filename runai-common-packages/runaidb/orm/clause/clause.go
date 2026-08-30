package clause

import gormclause "gorm.io/gorm/clause"

// Core interfaces.
type (
	Expression                = gormclause.Expression
	NegationExpressionBuilder = gormclause.NegationExpressionBuilder
	Interface                 = gormclause.Interface
	Builder                   = gormclause.Builder
	Writer                    = gormclause.Writer
	Assigner                  = gormclause.Assigner
	AssociationAssigner       = gormclause.AssociationAssigner
)

// Clause building blocks.
type (
	Clause        = gormclause.Clause
	ClauseBuilder = gormclause.ClauseBuilder
	Column        = gormclause.Column
	Table         = gormclause.Table
	Expr          = gormclause.Expr
	NamedExpr     = gormclause.NamedExpr
)

// Conditions.
type (
	Where         = gormclause.Where
	AndConditions = gormclause.AndConditions
	OrConditions  = gormclause.OrConditions
	NotConditions = gormclause.NotConditions
)

// Comparisons.
type (
	Eq   = gormclause.Eq
	Neq  = gormclause.Neq
	Gt   = gormclause.Gt
	Gte  = gormclause.Gte
	Lt   = gormclause.Lt
	Lte  = gormclause.Lte
	IN   = gormclause.IN
	Like = gormclause.Like
)

// Query clauses.
type (
	OrderBy         = gormclause.OrderBy
	OrderByColumn   = gormclause.OrderByColumn
	GroupBy         = gormclause.GroupBy
	Limit           = gormclause.Limit
	Select          = gormclause.Select
	From            = gormclause.From
	CommaExpression = gormclause.CommaExpression
)

// Join clauses.
type (
	Join       = gormclause.Join
	JoinType   = gormclause.JoinType
	JoinTarget = gormclause.JoinTarget
)

// DML clauses.
type (
	Insert     = gormclause.Insert
	Values     = gormclause.Values
	Update     = gormclause.Update
	Delete     = gormclause.Delete
	Set        = gormclause.Set
	Assignment = gormclause.Assignment
	OnConflict = gormclause.OnConflict
	Returning  = gormclause.Returning
	Locking    = gormclause.Locking
)

// Association operation types.
type (
	Association       = gormclause.Association
	AssociationOpType = gormclause.AssociationOpType
)

type With = gormclause.With

// Condition functions.
var (
	And = gormclause.And
	Or  = gormclause.Or
	Not = gormclause.Not
)

// Set/Assignment helpers.
var (
	Assignments       = gormclause.Assignments
	AssignmentColumns = gormclause.AssignmentColumns
)

// Join helpers.
var (
	Has       = gormclause.Has
	JoinTable = gormclause.JoinTable
)

// Well-known columns.
var PrimaryColumn = gormclause.PrimaryColumn

// String constants.
const (
	PrimaryKey   = gormclause.PrimaryKey
	CurrentTable = gormclause.CurrentTable
	Associations = gormclause.Associations
	AndWithSpace = gormclause.AndWithSpace
	OrWithSpace  = gormclause.OrWithSpace
)

// Join type constants.
const (
	CrossJoin JoinType = gormclause.CrossJoin
	InnerJoin JoinType = gormclause.InnerJoin
	LeftJoin  JoinType = gormclause.LeftJoin
	RightJoin JoinType = gormclause.RightJoin
)

// Locking constants.
const (
	LockingStrengthUpdate    = gormclause.LockingStrengthUpdate
	LockingStrengthShare     = gormclause.LockingStrengthShare
	LockingOptionsSkipLocked = gormclause.LockingOptionsSkipLocked
	LockingOptionsNoWait     = gormclause.LockingOptionsNoWait
)

// Association operation type constants.
const (
	OpUnlink AssociationOpType = gormclause.OpUnlink
	OpDelete AssociationOpType = gormclause.OpDelete
	OpUpdate AssociationOpType = gormclause.OpUpdate
	OpCreate AssociationOpType = gormclause.OpCreate
)
