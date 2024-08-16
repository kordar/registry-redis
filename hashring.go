package registry_redis

import (
	"encoding/json"
	"github.com/g4zhuj/hashring"
	"github.com/spf13/cast"
)

type HashringRegistry struct {
	count map[string]int
	hash  *hashring.HashRing
	node  string
}

func NewHashringRegistry(spots int, node string) HashringRegistry {
	return HashringRegistry{
		count: map[string]int{},
		hash:  hashring.NewHashRing(spots),
		node:  node,
	}
}

func (h *HashringRegistry) Load(values []string) {
	weights := map[string]int{}
	for _, value := range values {
		var tempMap map[string]interface{}
		err := json.Unmarshal([]byte(value), &tempMap)
		if err != nil {
			continue
		}
		node := cast.ToString(tempMap["node"])
		// 1:remove, 2:add, 3:none
		if h.count[node] == 0 {
			h.count[node] = 2
		} else if h.count[node] == 1 {
			h.count[node] = 3
		}

		weights[node] = cast.ToInt(tempMap["weight"])
	}

	tmp := map[string]int{}
	for n, v := range h.count {
		if v == 1 {
			h.hash.RemoveNode(n)
			continue
		}
		if v == 2 {
			w := 1
			if weights[n] != 0 {
				w = weights[n]
			}
			h.hash.AddNode(n, w)
		}
		tmp[n] = 1
	}

	h.count = tmp
}

func (h *HashringRegistry) GetNode(key string) string {
	return h.hash.GetNode(key)
}

func (h *HashringRegistry) Can(key string) bool {
	return h.hash.GetNode(key) == h.node
}
