package consistenthashing

import (
	"hash/fnv"
	"sort"
	"strconv"
)

type vnode struct {
	hash     uint32
	instance string
}

type ConsistentHash struct {
	replicas int
	ring     []vnode
}

func NewConsistentHash(replicas int, instances []string) *ConsistentHash {
	ch := &ConsistentHash{
		replicas: replicas,
		ring:     []vnode{},
	}

	for _, inst := range instances {
		ch.addInstance(inst)
	}

	sort.Slice(ch.ring, func(i, j int) bool {
		return ch.ring[i].hash < ch.ring[j].hash
	})

	return ch
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))

	return h.Sum32()
}

func (ch *ConsistentHash) addInstance(instance string) {
	for i := 0; i < ch.replicas; i++ {
		key := instance + "#" + strconv.Itoa(i)
		ch.ring = append(ch.ring, vnode{
			hash:     hash(key),
			instance: instance,
		})
	}
}

func (ch *ConsistentHash) Next(instances []string, key string) string {
	if len(ch.ring) == 0 {
		return ""
	}

	h := hash(key)

	idx := sort.Search(len(ch.ring), func(i int) bool {
		return ch.ring[i].hash >= h
	})

	if idx == len(ch.ring) {
		idx = 0
	}

	return ch.ring[idx].instance
}
