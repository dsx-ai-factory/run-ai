# dbmodels

`dbmodels` provides composable GORM base models that carry the default fields a domain model might require. Services embed the one that matches their delete semantic.

## `BaseModel` (soft deletes)

Embeds a `gorm.DeletedAt` field, so GORM soft-deletes rows (sets `deleted_at` instead of physically removing them).

```go
type Application struct {
	dbmodels.BaseModel
	Name string
}
```

## `DeleteBaseModel` (hard deletes)

Has no `DeletedAt` field, so GORM issues a real SQL `DELETE`. Suitable for association/join tables.

```go
type ProjectNodePoolResource struct {
	dbmodels.DeleteBaseModel
	ProjectID  uuid.UUID
	NodePoolID uuid.UUID
}
```

Both models implement a `BeforeCreate` hook that assigns a new UUID to `ID` when it is unset.

## `GormDeletedAtToNullableTime`

Converts a `gorm.DeletedAt` to a plain `*time.Time` for API responses — returns `nil` when the row isn't deleted, or a pointer to the deletion timestamp otherwise.
