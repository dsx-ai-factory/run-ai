package testing

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/stretchr/testify/require"
)

type encryptedModel struct {
	ID     int16  `gorm:"column:id;primaryKey"`
	Secret string `gorm:"column:secret;serializer:encrypted"`
}

func (encryptedModel) TableName() string { return "encrypted_models" }

func TestEncryptedField_Write(t *testing.T) {
	mocker := SetSQLMock(t)
	mocker.ExpectTransaction().Begin()
	// The encrypted secret is non-deterministic (random nonce), so its arg is AnyArg.
	mocker.ExpectQuery(`INSERT INTO encrypted_models (secret,id) VALUES (?,?) RETURNING id`, AnyArg(), int16(1)).
		Returns([]encryptedModel{{ID: 1}})
	mocker.ExpectTransaction().Commit()

	_, err := runaidb.DB().WithContext(context.Background()).Create(&encryptedModel{ID: 1, Secret: "s3cr3t"})
	require.NoError(t, err)
}

func TestEncryptedField_Read(t *testing.T) {
	mocker := SetSQLMock(t)
	mocker.ExpectQuery(`SELECT * FROM encrypted_models ORDER BY encrypted_models.id LIMIT ?`).
		Returns([]encryptedModel{{ID: 1, Secret: EncryptForTest(t, "s3cr3t")}})

	var got encryptedModel
	err := runaidb.DB().WithContext(context.Background()).First(&got)
	require.NoError(t, err)
	require.Equal(t, "s3cr3t", got.Secret)
}
