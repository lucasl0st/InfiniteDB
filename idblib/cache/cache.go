/*
 * Copyright (c) 2023 Lucas Pape
 */

package cache

import (
	idblib "github.com/lucasl0st/InfiniteDB/idblib/object"
	"sort"
	"sync"
	"time"
)

type Cache struct {
	m      sync.Map
	count  uint
	max    uint
	active bool
}

func New(max uint) *Cache {
	c := &Cache{
		count:  0,
		max:    max,
		active: true,
	}

	go func() {
		for {
			if !c.active {
				return
			}

			c.collector()

			time.Sleep(time.Second * 60)
		}
	}()

	return c
}

func (c *Cache) Kill() {
	c.active = false
}

func (c *Cache) Set(o idblib.Object) {
	c.count++

	c.m.Store(o.Id, object{
		m:        o.M,
		priority: 0,
	})
}

func (c *Cache) Remove(id int64) {
	c.m.Delete(id)
}

func (c *Cache) Get(id int64) *map[string]interface{} {
	a, have := c.m.Load(id)

	if have {
		o := a.(object)
		o.priority++

		c.m.Store(id, o)

		return &o.m
	}

	return nil
}

func (c *Cache) collector() {
	if c.count > c.max {
		var all []int64

		c.m.Range(func(key, value any) bool {
			all = append(all, key.(int64))
			return true
		})

		sort.Slice(all, func(i, j int) bool {
			first, _ := c.m.Load(all[i])
			second, _ := c.m.Load(all[j])

			return first.(object).priority > second.(object).priority
		})

		if len(all) < int(c.max) {
			c.count = uint(len(all))
			return
		}

		all = all[:c.max]

		c.m.Range(func(key, value any) bool {
			found := false

			for _, cid := range all {
				if key == cid {
					found = true
					break
				}
			}

			if !found {
				c.m.Delete(key)
			}

			return true
		})

		c.count = c.max
	}
}
