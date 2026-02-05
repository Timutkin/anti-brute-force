package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImMemoryListService_AddCIDR(t *testing.T) {
	service := NewInMemoryListService()

	t.Run("successful add CIDR", func(t *testing.T) {
		err := service.AddCIDR("192.168.1.0/24")
		assert.NoError(t, err)

		contains, err := service.IsIPContains("192.168.1.1")
		assert.NoError(t, err)
		assert.True(t, contains)
	})

	t.Run("invalid CIDR", func(t *testing.T) {
		err := service.AddCIDR("invalid-cidr")
		assert.Error(t, err)
		assert.Equal(t, ErrParseCIDR, err)
	})
}

func TestImMemoryListService_DeleteCIDR(t *testing.T) {
	service := NewInMemoryListService()

	err := service.AddCIDR("192.168.1.0/24")
	assert.NoError(t, err)

	t.Run("delete existing CIDR", func(t *testing.T) {
		err := service.DeleteCIDR("192.168.1.0/24")
		assert.NoError(t, err)

		contains, err := service.IsIPContains("192.168.1.1")
		assert.NoError(t, err)
		assert.False(t, contains)
	})

	t.Run("delete non-existing CIDR", func(t *testing.T) {
		err := service.DeleteCIDR("10.0.0.0/8")
		assert.NoError(t, err) // No error for deleting non-existing
	})
}

func TestImMemoryListService_GetCIDRs(t *testing.T) {
	service := NewInMemoryListService()

	cidrs := service.GetCIDRs()
	assert.Empty(t, cidrs)

	err := service.AddCIDR("192.168.1.0/24")
	assert.NoError(t, err)
	err = service.AddCIDR("10.0.0.0/8")
	assert.NoError(t, err)

	cidrs = service.GetCIDRs()
	assert.Len(t, cidrs, 2)
	assert.Contains(t, cidrs, "192.168.1.0/24")
	assert.Contains(t, cidrs, "10.0.0.0/8")
}

func TestImMemoryListService_IsIPContains(t *testing.T) {
	service := NewInMemoryListService()

	err := service.AddCIDR("192.168.1.0/24")
	assert.NoError(t, err)

	t.Run("IP in CIDR", func(t *testing.T) {
		contains, err := service.IsIPContains("192.168.1.1")
		assert.NoError(t, err)
		assert.True(t, contains)
	})

	t.Run("IP not in CIDR", func(t *testing.T) {
		contains, err := service.IsIPContains("192.168.2.1")
		assert.NoError(t, err)
		assert.False(t, contains)
	})

	t.Run("invalid IP", func(t *testing.T) {
		contains, err := service.IsIPContains("invalid-ip")
		assert.Error(t, err)
		assert.Equal(t, ErrParseIP, err)
		assert.False(t, contains)
	})
}

func TestImMemoryListService_Concurrency(t *testing.T) {
	service := NewInMemoryListService()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			cidr := fmt.Sprintf("192.168.%d.0/24", id)
			err := service.AddCIDR(cidr)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	cidrs := service.GetCIDRs()
	assert.Len(t, cidrs, 10)
}
