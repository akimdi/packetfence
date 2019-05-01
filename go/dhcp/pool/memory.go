package pool

import (
	"errors"
	"math/rand"
	"strconv"
	"sync"

)

var Memory Memoryt

type Memoryt struct {
	DHCPPool *DHCPPool
}

func NewMemoryPool(capacity uint64) (PoolBackend, error) {
	Pool := &Memoryt{}
	Pool.NewDHCPPool(capacity)
	return Pool, nil
}

func (dp *Memoryt) NewDHCPPool(capacity uint64) {
	d := &DHCPPool{
		lock:     &sync.Mutex{},
		free:     make(map[uint64]bool),
		mac:      make(map[uint64]string),
		capacity: capacity,
	}
	for i := uint64(0); i < d.capacity; i++ {
		d.free[i] = true
	}
	dp.DHCPPool = d
}

func (dp *Memoryt) GetDHCPPool() DHCPPool {
	return *dp.DHCPPool
}

// Compare what we have in the cache with what we have in the pool
func (dp *DHCPPool) GetIssues(macs []string) ([]string, map[uint64]string) {
	dp.lock.Lock()
	defer dp.lock.Unlock()
	var found bool
	found = false
	var inPoolNotInCache []string
	var duplicateInPool map[uint64]string
	duplicateInPool = make(map[uint64]string)

	var count int
	var saveindex uint64
	for i := uint64(0); i < dp.capacity; i++ {
		if dp.free[i] {
			continue
		}
		for _, mac := range macs {
			if dp.mac[i] == mac {
				found = true
			}
		}
		if !found {
			inPoolNotInCache = append(inPoolNotInCache, dp.mac[i]+", "+strconv.Itoa(int(i)))
		}
	}
	for _, mac := range macs {
		count = 0
		saveindex = 0

		for i := uint64(0); i < dp.capacity; i++ {
			if dp.free[i] {
				continue
			}
			if dp.mac[i] == mac {
				if count == 0 {
					saveindex = i
				}
				if count == 1 {
					duplicateInPool[saveindex] = mac
					duplicateInPool[i] = mac
				} else if count > 1 {
					duplicateInPool[i] = mac
				}
				count++
			}
		}
	}
	return inPoolNotInCache, duplicateInPool
}

// Reserves an IP in the pool, returns an error if the IP has already been reserved
func (dp *Memoryt) ReserveIPIndex(index uint64, mac string) (error, string) {
	dp.DHCPPool.lock.Lock()
	defer dp.DHCPPool.lock.Unlock()

	if index >= dp.DHCPPool.capacity {
		return errors.New("Trying to reserve an IP that is outside the capacity of this pool"), FreeMac
	}

	if _, free := dp.DHCPPool.free[index]; free {
		delete(dp.DHCPPool.free, index)
		dp.DHCPPool.mac[index] = mac
		return nil, mac
	} else {
		return errors.New("IP is already reserved"), FreeMac
	}
}

// Frees an IP in the pool, returns an error if the IP is already free
func (dp *Memoryt) FreeIPIndex(index uint64) error {
	dp.DHCPPool.lock.Lock()
	defer dp.DHCPPool.lock.Unlock()

	if !dp.IndexInPool(index) {
		return errors.New("Trying to free an IP that is outside the capacity of this pool")
	}

	if _, free := dp.DHCPPool.free[index]; free {
		return errors.New("IP is already free")
	} else {
		dp.DHCPPool.free[index] = true
		delete(dp.DHCPPool.mac, index)
		return nil
	}
}

// Check if the IP is free at the index
func (dp *Memoryt) IsFreeIPAtIndex(index uint64) bool {
	dp.DHCPPool.lock.Lock()
	defer dp.DHCPPool.lock.Unlock()

	if !dp.IndexInPool(index) {
		return false
	}

	if _, free := dp.DHCPPool.free[index]; free {
		return true
	} else {
		return false
	}
}

// Check if the IP is free at the index
func (dp *Memoryt) GetMACIndex(index uint64) (uint64, string, error) {
	dp.DHCPPool.lock.Lock()
	defer dp.DHCPPool.lock.Unlock()

	if !dp.IndexInPool(index) {
		return index, FreeMac, errors.New("The index is not part of the pool")
	}

	if _, free := dp.DHCPPool.free[index]; free {
		return index, FreeMac, nil
	} else {
		return index, dp.DHCPPool.mac[index], nil
	}
}

// Returns a random free IP address, an error if the pool is full
func (dp *Memoryt) GetFreeIPIndex(mac string) (uint64, string, error) {
	dp.DHCPPool.lock.Lock()
	defer dp.DHCPPool.lock.Unlock()

	if len(dp.DHCPPool.free) == 0 {
		return 0, FreeMac, errors.New("DHCP pool is full")
	}
	index := rand.Intn(len(dp.DHCPPool.free))

	var available uint64
	for available = range dp.DHCPPool.free {
		if index == 0 {
			break
		}
		index--
	}

	delete(dp.DHCPPool.free, available)
	dp.DHCPPool.mac[available] = mac

	return available, mac, nil
}

// Returns whether or not a specific index is in the capacity of the pool
func (dp *Memoryt) IndexInPool(index uint64) bool {
	return index < dp.DHCPPool.capacity
}

// Returns the amount of free IPs in the pool
func (dp *Memoryt) FreeIPsRemaining() uint64 {
	dp.DHCPPool.lock.Lock()
	defer dp.DHCPPool.lock.Unlock()
	return uint64(len(dp.DHCPPool.free))
}

// Returns the capacity of the pool
func (dp *Memoryt) Capacity() uint64 {
	return dp.DHCPPool.capacity
}
