package service

import (
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/timutkin/anti-brute-force/internal/config"
)

var (
	cfg3 = config.AttemptsConfig{
		IPMaxAttempts:       3,
		PasswordMaxAttempts: 3,
		LoginMaxAttempts:    3,
	}
	cfg100 = config.AttemptsConfig{
		IPMaxAttempts:       100,
		PasswordMaxAttempts: 100,
		LoginMaxAttempts:    100,
	}
	ip       = "95.85.231.45"
	login    = "login"
	password = "password"
)

func TestIsCredentialAllowed_SingleGoroutine(t *testing.T) {
	const attempts = 3
	mockBlackListService := NewMockListService(t)
	mockWhiteListService := NewMockListService(t)
	mockBlackListService.EXPECT().IsIPContains(ip).Return(false, nil)
	mockWhiteListService.EXPECT().IsIPContains(ip).Return(false, nil)
	t.Run("login attempts more than 3 then return false", func(t *testing.T) {
		bruteForceService := NewInMemoryBruteForceService(mockBlackListService, mockWhiteListService, cfg3)
		for i := 0; i < attempts; i++ {
			ok, _ := bruteForceService.IsCredentialAllowed(login, password, ip)
			assert.True(t, ok)
		}
		ok, _ := bruteForceService.IsCredentialAllowed(login, password, ip)
		assert.False(t, ok)
	})

	t.Run("password attempts more than 3 then return false", func(t *testing.T) {
		bruteForceService := NewInMemoryBruteForceService(mockBlackListService, mockWhiteListService, cfg3)
		for i := 0; i < attempts; i++ {
			ok, _ := bruteForceService.IsCredentialAllowed(uuid.NewString(), password, ip)
			assert.True(t, ok)
		}
		ok, _ := bruteForceService.IsCredentialAllowed(uuid.NewString(), password, ip)
		assert.False(t, ok)
	})

	t.Run("ip attempts more than 3 then return false", func(t *testing.T) {
		bruteForceService := NewInMemoryBruteForceService(mockBlackListService, mockWhiteListService, cfg3)
		for i := 0; i < attempts; i++ {
			ok, _ := bruteForceService.IsCredentialAllowed(uuid.NewString(), uuid.NewString(), ip)
			assert.True(t, ok)
		}
		ok, _ := bruteForceService.IsCredentialAllowed(uuid.NewString(), uuid.NewString(), ip)
		assert.False(t, ok)
	})

	t.Run("ip in black list then return false", func(t *testing.T) {
		mockBlackListService := NewMockListService(t)
		bruteForceService := NewInMemoryBruteForceService(mockBlackListService, mockBlackListService, cfg3)
		mockBlackListService.EXPECT().IsIPContains(ip).Return(true, nil)
		ok, _ := bruteForceService.IsCredentialAllowed(login, password, ip)
		assert.False(t, ok)
	})

	t.Run("ip in white list then return true", func(t *testing.T) {
		mockBlackListService := NewMockListService(t)
		mockWhiteListService := NewMockListService(t)
		bruteForceService := NewInMemoryBruteForceService(mockBlackListService, mockWhiteListService, cfg3)
		mockBlackListService.EXPECT().IsIPContains(ip).Return(false, nil)
		mockWhiteListService.EXPECT().IsIPContains(ip).Return(true, nil)
		ok, _ := bruteForceService.IsCredentialAllowed(login, password, ip)
		assert.True(t, ok)
	})
}

func TestIsCredentialAllowed_SeveralGoroutine(t *testing.T) {
	const attempts = 100
	mockBlackListService := NewMockListService(t)
	mockWhiteListService := NewMockListService(t)
	mockBlackListService.EXPECT().IsIPContains(ip).Return(false, nil)
	mockWhiteListService.EXPECT().IsIPContains(ip).Return(false, nil)
	t.Run("login attempts more than 100 then return false", func(t *testing.T) {
		bruteForceService := NewInMemoryBruteForceService(mockBlackListService, mockWhiteListService, cfg100)
		wg := sync.WaitGroup{}
		for i := 0; i < attempts; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ok, _ := bruteForceService.IsCredentialAllowed(login, password, ip)
				assert.True(t, ok)
			}()
		}
		wg.Wait()
		ok, _ := bruteForceService.IsCredentialAllowed(login, password, ip)
		assert.False(t, ok)
	})

	t.Run("password attempts more than 100 then return false", func(t *testing.T) {
		bruteForceService := NewInMemoryBruteForceService(mockBlackListService, mockWhiteListService, cfg100)
		wg := sync.WaitGroup{}
		for i := 0; i < attempts; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ok, _ := bruteForceService.IsCredentialAllowed(uuid.NewString(), password, ip)
				assert.True(t, ok)
			}()

		}
		wg.Wait()
		ok, _ := bruteForceService.IsCredentialAllowed(uuid.NewString(), password, ip)
		assert.False(t, ok)
	})

	t.Run("ip attempts more than 100 then return false", func(t *testing.T) {
		bruteForceService := NewInMemoryBruteForceService(mockBlackListService, mockWhiteListService, cfg100)
		wg := sync.WaitGroup{}
		for i := 0; i < attempts; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ok, _ := bruteForceService.IsCredentialAllowed(uuid.NewString(), uuid.NewString(), ip)
				assert.True(t, ok)
			}()
		}
		wg.Wait()
		ok, _ := bruteForceService.IsCredentialAllowed(uuid.NewString(), uuid.NewString(), ip)
		assert.False(t, ok)
	})
}

func TestGetLoginBuckets(t *testing.T) {
	mockBlackListService := NewMockListService(t)
	mockWhiteListService := NewMockListService(t)
	mockBlackListService.EXPECT().IsIPContains(ip).Return(false, nil)
	mockWhiteListService.EXPECT().IsIPContains(ip).Return(false, nil)

	bruteForceService := NewInMemoryBruteForceService(mockBlackListService, mockWhiteListService, cfg3)

	buckets := bruteForceService.GetLoginBuckets()
	assert.Len(t, buckets, 1)

	for i := 0; i < 2; i++ {
		ok, _ := bruteForceService.IsCredentialAllowed(login, password, ip)
		assert.True(t, ok)
	}

	buckets = bruteForceService.GetLoginBuckets()
	assert.Len(t, buckets, 1)
	assert.Equal(t, "login", buckets[0].Name)
	assert.Len(t, buckets[0].Buckets, 1)
	assert.Equal(t, login, buckets[0].Buckets[0].Key)
	assert.Equal(t, 2, buckets[0].Buckets[0].Attempts)
	assert.NotZero(t, buckets[0].Buckets[0].FirstTime)
	assert.NotZero(t, buckets[0].Buckets[0].LastTime)
}

func TestDeleteBuckets(t *testing.T) {
	t.Run("delete login bucket", func(t *testing.T) {
		mockBlackListService := NewMockListService(t)
		mockWhiteListService := NewMockListService(t)
		mockBlackListService.EXPECT().IsIPContains(ip).Return(false, nil)
		mockWhiteListService.EXPECT().IsIPContains(ip).Return(false, nil)

		bruteForceService := NewInMemoryBruteForceService(mockBlackListService, mockWhiteListService, cfg3)
		ok, _ := bruteForceService.IsCredentialAllowed(login, password, ip)
		assert.True(t, ok)
		buckets := bruteForceService.GetLoginBuckets()
		assert.Len(t, buckets[0].Buckets, 1)

		bruteForceService.DeleteBuckets(login, "")
		buckets = bruteForceService.GetLoginBuckets()
		assert.Len(t, buckets[0].Buckets, 0)
	})

	t.Run("delete ip bucket", func(t *testing.T) {
		mockBlackListService := NewMockListService(t)
		mockWhiteListService := NewMockListService(t)
		mockBlackListService.EXPECT().IsIPContains(ip).Return(false, nil)
		mockWhiteListService.EXPECT().IsIPContains(ip).Return(false, nil)

		bruteForceService := NewInMemoryBruteForceService(mockBlackListService, mockWhiteListService, cfg3)
		for i := 0; i < 3; i++ {
			ok, _ := bruteForceService.IsCredentialAllowed(fmt.Sprintf("login%d", i), fmt.Sprintf("pass%d", i), ip)
			assert.True(t, ok)
		}
		ok, _ := bruteForceService.IsCredentialAllowed("login4", "pass4", ip)
		assert.False(t, ok)

		bruteForceService.DeleteBuckets("", ip)

		ok, _ = bruteForceService.IsCredentialAllowed("login5", "pass5", ip)
		assert.True(t, ok)
	})
}
