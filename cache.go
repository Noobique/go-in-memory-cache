/*
in-memory cache with expiration
*/

package cache

import (
    "context"
    "sync"
    "time"
)

type Message struct {
    id      string
    payload string
}

type Cache struct {
    mu    sync.RWMutex
    cache map[string]*CachedItem
    ttl   time.Duration
}

type CachedItem struct {
    message    *Message
    expiration time.Time
}

func (c *Cache) Get(key string) (*Message, bool) {
    c.mu.Lock()
    ci, ok := c.cache[key]
    c.mu.Unlock()
    if !ok {
         return nil, false
    }
    if time.Now().After(ci.expiration) {
        return nil, false
    }

    return ci.message, true
}

func (c *Cache) Set(m *Message) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.cache[m.id] = &CachedItem{
        message: m,
        expiration: time.Now().Add(c.ttl),
    }
}

func (c *Cache) Clr(ctx context.Context) {
    var clr []string
    ticker := time.NewTicker(c.ttl)
    defer ticker.Stop()
    for {
        select {
        case <- ctx.Done():
            return
        case <- ticker.C:
            for key, ci := range c.cache {
                if time.Now().After(ci.expiration) {
                    clr = append(clr, key)
                }
            }
            c.mu.Lock()
            for _, key := range clr {
                delete(c.cache, key)
            }
            c.mu.Unlock()
        }
    }
}

func NewCache(ctx context.Context, ttl time.Duration) *Cache {
    cache := &Cache{
        cache: make(map[string] *CachedItem),
        ttl: ttl,
    }
    go cache.Clr(ctx)
    return cache
}