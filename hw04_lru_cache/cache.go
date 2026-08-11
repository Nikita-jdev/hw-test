package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type cacheItem struct {
	key   Key
	value interface{}
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	if item, exists := c.items[key]; exists {
		item.Value.(*cacheItem).value = value
		c.queue.MoveToFront(item)
		return true
	}

	newItem := &cacheItem{key: key, value: value}
	listItem := c.queue.PushFront(newItem)
	c.items[key] = listItem

	if c.queue.Len() > c.capacity {
		back := c.queue.Back()
		if back != nil {
			item := back.Value.(*cacheItem)
			delete(c.items, item.key)
			c.queue.Remove(back)
		}
	}

	return false
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	if item, exists := c.items[key]; exists {
		c.queue.MoveToFront(item)
		return item.Value.(*cacheItem).value, true
	}
	return nil, false
}

func (c *lruCache) Clear() {
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}
