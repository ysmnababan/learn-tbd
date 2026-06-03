package abstraction

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestEntityBeforeCreate(t *testing.T) {
	tests := []struct {
		name string
		entity *Entity
		verify func(t *testing.T, entity *Entity)
	}{
		{
			name: "creates timestamp",
			entity: &Entity{
				CreatedBy: "user1",
			},
			verify: func(t *testing.T, entity *Entity) {
				assert.Greater(t, entity.CreatedAt, int64(0))
				now := time.Now().UnixMilli()
				assert.LessOrEqual(t, entity.CreatedAt, now)
				assert.Equal(t, "user1", entity.CreatedBy)
			},
		},
		{
			name: "preserves existing fields",
			entity: &Entity{
				CreatedBy:  "test_user",
				ModifiedBy: "another_user",
			},
			verify: func(t *testing.T, entity *Entity) {
				assert.Greater(t, entity.CreatedAt, int64(0))
				assert.Equal(t, "test_user", entity.CreatedBy)
				assert.Equal(t, "another_user", entity.ModifiedBy)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entity.BeforeCreate(&gorm.DB{})
			assert.NoError(t, err)
			tt.verify(t, tt.entity)
		})
	}
}

func TestEntityBeforeUpdate(t *testing.T) {
	tests := []struct {
		name string
		entity *Entity
		verify func(t *testing.T, entity *Entity)
	}{
		{
			name: "updates modified timestamp",
			entity: &Entity{
				CreatedAt: 1000000,
				CreatedBy: "user1",
			},
			verify: func(t *testing.T, entity *Entity) {
				assert.NotNil(t, entity.ModifiedAt)
				assert.Greater(t, *entity.ModifiedAt, int64(0))
				now := time.Now().UnixMilli()
				assert.LessOrEqual(t, *entity.ModifiedAt, now)
				assert.Equal(t, int64(1000000), entity.CreatedAt)
			},
		},
		{
			name: "preserves other fields",
			entity: &Entity{
				CreatedAt:  1000000,
				CreatedBy:  "user1",
				ModifiedBy: "user2",
				DeletedBy:  "user3",
			},
			verify: func(t *testing.T, entity *Entity) {
				assert.NotNil(t, entity.ModifiedAt)
				assert.Equal(t, int64(1000000), entity.CreatedAt)
				assert.Equal(t, "user1", entity.CreatedBy)
				assert.Equal(t, "user2", entity.ModifiedBy)
				assert.Equal(t, "user3", entity.DeletedBy)
			},
		},
		{
			name: "handles nil entity",
			entity: &Entity{},
			verify: func(t *testing.T, entity *Entity) {
				assert.NotNil(t, entity.ModifiedAt)
				assert.Greater(t, *entity.ModifiedAt, int64(0))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entity.BeforeUpdate(&gorm.DB{})
			assert.NoError(t, err)
			tt.verify(t, tt.entity)
		})
	}
}

func TestEntityStructFields(t *testing.T) {
	entity := &Entity{
		CreatedAt:  1000000,
		CreatedBy:  "user1",
		ModifiedAt: ptrInt64(2000000),
		ModifiedBy: "user2",
		DeletedAt:  ptrInt64(3000000),
		DeletedBy:  "user3",
	}

	assert.Equal(t, int64(1000000), entity.CreatedAt)
	assert.Equal(t, "user1", entity.CreatedBy)
	assert.Equal(t, int64(2000000), *entity.ModifiedAt)
	assert.Equal(t, "user2", entity.ModifiedBy)
	assert.Equal(t, int64(3000000), *entity.DeletedAt)
	assert.Equal(t, "user3", entity.DeletedBy)
}

func TestFilterStructFields(t *testing.T) {
	createdAtVal := int64(1000000)
	createdByVal := 123
	modifiedAtVal := int64(2000000)
	modifiedByVal := 456

	filter := &Filter{
		CreatedAt:  &createdAtVal,
		CreatedBy:  &createdByVal,
		ModifiedAt: &modifiedAtVal,
		ModifiedBy: &modifiedByVal,
	}

	assert.NotNil(t, filter.CreatedAt)
	assert.Equal(t, int64(1000000), *filter.CreatedAt)
	assert.NotNil(t, filter.CreatedBy)
	assert.Equal(t, 123, *filter.CreatedBy)
	assert.NotNil(t, filter.ModifiedAt)
	assert.Equal(t, int64(2000000), *filter.ModifiedAt)
	assert.NotNil(t, filter.ModifiedBy)
	assert.Equal(t, 456, *filter.ModifiedBy)
}

func TestFilterNilFields(t *testing.T) {
	filter := &Filter{}

	assert.Nil(t, filter.CreatedAt)
	assert.Nil(t, filter.CreatedBy)
	assert.Nil(t, filter.ModifiedAt)
	assert.Nil(t, filter.ModifiedBy)
}

func ptrInt64(val int64) *int64 {
	return &val
}
