package TTL

import (
	"sync"
	"time"
)

type item struct {
	id        int
	queue     string
	ttl       int64
	timestamp int64
}

type TTLMap struct {
	m map[string]*item
	l sync.Mutex
}

func New(ln int, maxTTL int) (m *TTLMap) {
	m = &TTLMap{m: make(map[string]*item, ln)}
	go func() {
		for now := range time.Tick(time.Second) {
			m.l.Lock()
			for k, v := range m.m {
				if now.Unix() > v.timestamp+int64(v.ttl) {
					delete(m.m, k)
				}
			}
			m.l.Unlock()
		}
	}()
	return
}

func (m *TTLMap) Len() int {
	return len(m.m)
}

func (m *TTLMap) Put(k string, v item) {
	m.l.Lock()
	it, ok := m.m[k]
	if !ok {
		it = &v
		m.m[k] = it
	}
	it.timestamp = time.Now().Unix()
	m.l.Unlock()
}

func (m *TTLMap) Get(k string) (v item) {
	m.l.Lock()
	if it, ok := m.m[k]; ok {
		v = *it
		it.timestamp = time.Now().Unix()
	}
	m.l.Unlock()
	return

}
