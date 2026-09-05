package cache

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"time"
)

type Memory struct {
	mu sync.Mutex

	items map[string]*list.Element
	order *list.List

	maxEntries int

	now func() time.Time

	hits        uint64
	misses      uint64
	evictions   uint64
	sets        uint64
	deletes     uint64
	expirations uint64
}

type entry struct {
	key       string
	value     []byte
	expiresAt time.Time
}

type Stats struct {
	Entries     int
	MaxItems    int
	Hits        uint64
	Misses      uint64
	Evictions   uint64
	Sets        uint64
	Deletes     uint64
	Expirations uint64
}

func NewMemory(
	maxEntries int,
) (*Memory, error) {
	if maxEntries <= 0 {
		return nil, errors.New(
			"cache: max entries must be greater than zero",
		)
	}

	return &Memory{
		items:      make(map[string]*list.Element),
		order:      list.New(),
		maxEntries: maxEntries,
		now:        time.Now,
	}, nil
}

func (m *Memory) Get(
	ctx context.Context,
	key string,
) ([]byte, bool, error) {
	if err := validateContext(ctx); err != nil {
		return nil, false, err
	}

	if key == "" {
		return nil, false, errors.New(
			"cache: key is required",
		)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	element, ok := m.items[key]
	if !ok {
		m.misses++
		return nil, false, nil
	}

	item := element.Value.(entry)

	if !item.expiresAt.After(m.now()) {
		delete(m.items, key)
		m.order.Remove(element)

		m.misses++
		m.expirations++

		return nil, false, nil
	}

	m.order.MoveToFront(element)
	m.hits++

	return append(
		[]byte(nil),
		item.value...,
	), true, nil
}

func (m *Memory) Set(
	ctx context.Context,
	key string,
	value []byte,
	ttl time.Duration,
) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if key == "" {
		return errors.New(
			"cache: key is required",
		)
	}

	if ttl <= 0 {
		return errors.New(
			"cache: ttl must be greater than zero",
		)
	}

	copied := append(
		[]byte(nil),
		value...,
	)

	m.mu.Lock()
	defer m.mu.Unlock()

	expiresAt := m.now().Add(ttl)

	if existing, ok := m.items[key]; ok {
		existing.Value = entry{
			key:       key,
			value:     copied,
			expiresAt: expiresAt,
		}

		m.order.MoveToFront(existing)
		m.sets++

		return nil
	}

	element := m.order.PushFront(
		entry{
			key:       key,
			value:     copied,
			expiresAt: expiresAt,
		},
	)

	m.items[key] = element
	m.sets++

	if len(m.items) > m.maxEntries {
		m.evictOldest()
	}

	return nil
}

func (m *Memory) Delete(
	ctx context.Context,
	key string,
) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if key == "" {
		return errors.New(
			"cache: key is required",
		)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if element, ok := m.items[key]; ok {
		delete(m.items, key)
		m.order.Remove(element)
		m.deletes++
	}

	return nil
}

func (m *Memory) Clear(
	ctx context.Context,
) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	clear(m.items)
	m.order.Init()

	return nil
}

func (m *Memory) Stats() Stats {
	m.mu.Lock()
	defer m.mu.Unlock()

	return Stats{
		Entries:     len(m.items),
		MaxItems:    m.maxEntries,
		Hits:        m.hits,
		Misses:      m.misses,
		Evictions:   m.evictions,
		Sets:        m.sets,
		Deletes:     m.deletes,
		Expirations: m.expirations,
	}
}

func (m *Memory) evictOldest() {
	element := m.order.Back()

	if element == nil {
		return
	}

	item := element.Value.(entry)

	delete(
		m.items,
		item.key,
	)

	m.order.Remove(element)

	m.evictions++
}

func validateContext(
	ctx context.Context,
) error {
	if ctx == nil {
		return errors.New(
			"cache: context is required",
		)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}
