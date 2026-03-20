package in_memory

import "container/list"

type lruEntry struct {
	short  string
	origin string
}

type lruCache struct {
	capacity int

	ll *list.List

	shortToElem  map[string]*list.Element
	originToElem map[string]*list.Element
}

func newLRUCache(capacity int) *lruCache {

	return &lruCache{
		capacity:     capacity,
		ll:           list.New(),
		shortToElem:  make(map[string]*list.Element, capacity),
		originToElem: make(map[string]*list.Element, capacity),
	}
}

func (c *lruCache) GetByShort(short string) (string, bool) {

	elem, ok := c.shortToElem[short]
	if !ok {
		return "", false
	}

	c.ll.MoveToFront(elem)
	ent := elem.Value.(*lruEntry)
	return ent.origin, true
}

func (c *lruCache) GetByOrigin(origin string) (string, bool) {

	elem, ok := c.originToElem[origin]
	if !ok {
		return "", false
	}

	c.ll.MoveToFront(elem)
	ent := elem.Value.(*lruEntry)
	return ent.short, true
}

func (c *lruCache) Put(short, origin string) {

	elem := c.ll.PushFront(&lruEntry{short: short, origin: origin})
	c.shortToElem[short] = elem
	c.originToElem[origin] = elem

	if c.ll.Len() > c.capacity {
		c.removeOldest()
	}
}

func (c *lruCache) removeOldest() {
	elem := c.ll.Back()
	if elem == nil {
		return
	}

	ent := elem.Value.(*lruEntry)
	delete(c.shortToElem, ent.short)
	delete(c.originToElem, ent.origin)
	c.ll.Remove(elem)
}
