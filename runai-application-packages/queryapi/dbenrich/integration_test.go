package dbenrich_test

import (
	"testing"

	"github.com/run-ai/runai/runai-application-packages/queryapi"
	"github.com/run-ai/runai/runai-application-packages/queryapi/dbenrich"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

func TestParserAndEnrichers(t *testing.T) {
	mapping := queryapi.APIFieldsMapping{
		"name":      {Column: "name"},
		"createdAt": {Column: "created_at", Validators: []queryapi.Validator{queryapi.DateValidator{}}},
	}

	parser := queryapi.NewParser(mapping)
	filters, err := parser.Filters([]string{"name==alice||bob", "createdAt>=2023-01-01"})
	require.NoError(t, err)

	sortColumn, err := parser.GetMapping("createdAt")
	require.NoError(t, err)

	db, mocker := dbtesting.NewMocker(t)
	mocker.ExpectQuery(
		"SELECT * FROM items WHERE (name = ? OR name = ?) AND created_at >= ? ORDER BY created_at ASC NULLS FIRST",
		"alice", "bob", "2023-01-01",
	).Returns([]item{})

	var out []item
	err = dbenrich.SortBy(sortColumn, "asc")(
		dbenrich.Filters(filters)(db.Table("items")),
	).Find(&out)
	require.NoError(t, err)
}
