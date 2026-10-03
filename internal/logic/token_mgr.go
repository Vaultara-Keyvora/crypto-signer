package logic

import (
	"context"
	"fmt"
	"lbe_crypto_signer/pkg/tools/log"
	"sync"
)

const (
	MaxMissCount = 5
)

type TokenHolder struct {
	tokens    sync.Map
	lock      sync.Mutex
	missCount int
}

func NewTokenHolder() *TokenHolder {
	return &TokenHolder{}
}

func (t *TokenHolder) Load(ctx context.Context, token string) (any, bool) {
	value, ok := t.tokens.Load(token)
	if ok {
		return value, true
	}

	t.updateMissCount(ctx)

	return nil, false
}

func (t *TokenHolder) Store(token string, value any) {
	t.tokens.Store(token, value)
}

func (t *TokenHolder) Delete(token string) {
	t.tokens.Delete(token)
}

func (t *TokenHolder) updateMissCount(ctx context.Context) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.missCount++
	log.ZInfo(ctx, "configuration token cache miss", "stage", "token_cache", "result", "miss", "miss_count", t.missCount)
	if t.missCount >= MaxMissCount {
		t.clearTokens(ctx)
		t.missCount = 0
	}
}

func (t *TokenHolder) clearTokens(ctx context.Context) {
	deletedCount := 0
	t.tokens.Range(func(key, value interface{}) bool {
		t.tokens.Delete(key)
		deletedCount++
		return true
	})
	log.ZInfo(ctx, "configuration token cache cleared", "stage", "token_cache", "result", "success", "deleted_count", deletedCount, "miss_count", t.missCount)
}

const (
	TokenLength = 64
)

func MaskToken(token string) string {
	if len(token) < TokenLength {
		return "****"
	}
	return fmt.Sprintf("%v****%v", token[:4], token[len(token)-4:])
}
