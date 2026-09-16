package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testLogger() *zap.Logger {
	return zap.NewNop()
}

func TestHybridCache_Disabled(t *testing.T) {
	hc, err := NewHybridCache(10, nil, testLogger(), false)
	require.NoError(t, err)
	defer hc.Close()

	ctx := context.Background()
	require.NoError(t, hc.Set(ctx, "k", "v", time.Minute))
	_, found := hc.Get(ctx, "k")
	require.False(t, found)
	require.False(t, hc.RedisEnabled())
}

func TestHybridCache_LocalOnly(t *testing.T) {
	hc, err := NewHybridCache(100, nil, testLogger(), true)
	require.NoError(t, err)
	defer hc.Close()

	require.False(t, hc.RedisEnabled())
	require.NotNil(t, hc.ristretto)
	require.Nil(t, hc.redis)

	ctx := context.Background()
	require.NoError(t, hc.Set(ctx, "session:1", map[string]string{"id": "1"}, time.Minute))

	got, found := hc.Get(ctx, "session:1")
	require.True(t, found)
	m, ok := got.(map[string]string)
	require.True(t, ok)
	require.Equal(t, "1", m["id"])
	require.Equal(t, int64(1), hc.GetMetrics().RistrettoHits)

	require.NoError(t, hc.Delete(ctx, "session:1"))
	_, found = hc.Get(ctx, "session:1")
	require.False(t, found)
}

func TestHybridCache_RedisOnly_NoLocalStore(t *testing.T) {
	mr := newMiniredis(t)
	client := redis.NewClient(&redis.Options{Addr: mr})
	t.Cleanup(func() { _ = client.Close() })

	hc, err := NewHybridCache(100, client, testLogger(), true)
	require.NoError(t, err)
	defer hc.Close()

	require.True(t, hc.RedisEnabled())
	require.Nil(t, hc.ristretto)

	ctx := context.Background()
	payload := map[string]any{"user": "ada"}
	require.NoError(t, hc.Set(ctx, "k1", payload, time.Minute))

	raw, err := client.Get(ctx, "k1").Result()
	require.NoError(t, err)
	require.Contains(t, raw, "ada")

	got, found := hc.Get(ctx, "k1")
	require.True(t, found)
	require.Equal(t, int64(1), hc.GetMetrics().RedisHits)
	require.Equal(t, int64(0), hc.GetMetrics().RistrettoHits)

	decoded, ok := got.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "ada", decoded["user"])

	require.NoError(t, hc.Delete(ctx, "k1"))
	_, found = hc.Get(ctx, "k1")
	require.False(t, found)
}

func TestHybridCache_RedisSetDoesNotWriteRistretto(t *testing.T) {
	mr := newMiniredis(t)
	client := redis.NewClient(&redis.Options{Addr: mr})
	t.Cleanup(func() { _ = client.Close() })

	hc, err := NewHybridCache(100, client, testLogger(), true)
	require.NoError(t, err)
	defer hc.Close()

	ctx := context.Background()
	require.NoError(t, hc.Set(ctx, "only-redis", "x", time.Minute))
	require.Nil(t, hc.ristretto)

	b, err := json.Marshal("x")
	require.NoError(t, err)
	_ = b
}
