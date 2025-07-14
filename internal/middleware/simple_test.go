package middleware

import (
	"testing"
	"time"
)

func TestSimple(t *testing.T) {
	t.Log("Simple test running")

	// 测试令牌桶创建
	tb := NewTokenBucket(5, 2, time.Second)
	if tb.capacity != 5 {
		t.Errorf("Expected capacity 5, got %d", tb.capacity)
	}
	if tb.refillRate != 2 {
		t.Errorf("Expected refillRate 2, got %d", tb.refillRate)
	}

	t.Log("Simple test passed")
}
