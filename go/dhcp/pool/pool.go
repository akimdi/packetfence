package pool

import (
	"errors"
	"math/rand"
	"strconv"
	"sync"
)

const FreeMac = "00:00:00:00:00:00"
const FakeMac = "ff:ff:ff:ff:ff:ff"

type PoolBackend interface {
	NewDHCPPool(capacity uint64)
	ReserveIPIndex(index uint64, mac string) (error, string)
	IsFreeIPAtIndex(index uint64) bool
	GetMACIndex(index uint64) (uint64, string, error)
	GetFreeIPIndex(mac string) (uint64, string, error)
	IndexInPool(index uint64) bool
	FreeIPsRemaining() uint64
	Capacity() uint64
}

type Pool struct {
	DHCPPool *DHCPPool
	Backend  map[string]BackendType
}

type BackendType struct {
	// Option string
	Manage PoolBackend
}

var Tlv = Pool{
	Backend: map[string]BackendType{
		// "memory": Memory,
		"mysql": BackendType{Mysql},
	},
}

var Mysql Mysqlt

type Mysqlt struct {
	DHCPPool_local *DHCPPool
}

func (pool Mysqlt) NewDHCPPool(capacity uint64) {
	d := &DHCPPool{
		lock:     &sync.Mutex{},
		free:     make(map[uint64]bool),
		mac:      make(map[uint64]string),
		capacity: capacity,
	}
	for i := uint64(0); i < d.capacity; i++ {
		d.free[i] = true
	}
	pool.DHCPPool_local = d
	// return d
}

// Reserves an IP in the pool, returns an error if the IP has already been reserved
func (pool Mysqlt) ReserveIPIndex(index uint64, mac string) (error, string) {
	pool.DHCPPool_local.lock.Lock()
	defer pool.DHCPPool_local.lock.Unlock()

	if index >= pool.DHCPPool_local.capacity {
		return errors.New("Trying to reserve an IP that is outside the capacity of this pool"), FreeMac
	}

	if _, free := pool.DHCPPool_local.free[index]; free {
		delete(pool.DHCPPool_local.free, index)
		pool.DHCPPool_local.mac[index] = mac
		return nil, mac
	} else {
		return errors.New("IP is already reserved"), FreeMac
	}
}

// Frees an IP in the pool, returns an error if the IP is already free
func (dp Mysqlt) FreeIPIndex(index uint64) error {
	dp.DHCPPool_local.lock.Lock()
	defer dp.DHCPPool_local.lock.Unlock()

	if !dp.IndexInPool(index) {
		return errors.New("Trying to free an IP that is outside the capacity of this pool")
	}

	if _, free := dp.DHCPPool_local.free[index]; free {
		return errors.New("IP is already free")
	} else {
		dp.DHCPPool_local.free[index] = true
		delete(dp.DHCPPool_local.mac, index)
		return nil
	}
}

// Check if the IP is free at the index
func (dp Mysqlt) IsFreeIPAtIndex(index uint64) bool {
	dp.DHCPPool_local.lock.Lock()
	defer dp.DHCPPool_local.lock.Unlock()

	if !dp.IndexInPool(index) {
		return false
	}

	if _, free := dp.DHCPPool_local.free[index]; free {
		return true
	} else {
		return false
	}
}

// Check if the IP is free at the index
func (dp Mysqlt) GetMACIndex(index uint64) (uint64, string, error) {
	dp.DHCPPool_local.lock.Lock()
	defer dp.DHCPPool_local.lock.Unlock()

	if !dp.DHCPPool_local.IndexInPool(index) {
		return index, FreeMac, errors.New("The index is not part of the pool")
	}

	if _, free := dp.DHCPPool_local.free[index]; free {
		return index, FreeMac, nil
	} else {
		return index, dp.DHCPPool_local.mac[index], nil
	}
}

// Returns a random free IP address, an error if the pool is full
func (dp Mysqlt) GetFreeIPIndex(mac string) (uint64, string, error) {
	dp.DHCPPool_local.lock.Lock()
	defer dp.DHCPPool_local.lock.Unlock()

	if len(dp.DHCPPool_local.free) == 0 {
		return 0, FreeMac, errors.New("DHCP pool is full")
	}
	index := rand.Intn(len(dp.DHCPPool_local.free))

	var available uint64
	for available = range dp.DHCPPool_local.free {
		if index == 0 {
			break
		}
		index--
	}

	delete(dp.DHCPPool_local.free, available)
	dp.DHCPPool_local.mac[available] = mac

	return available, mac, nil
}

// Returns whether or not a specific index is in the capacity of the pool
func (dp Mysqlt) IndexInPool(index uint64) bool {
	return index < dp.DHCPPool_local.capacity
}

// Returns the amount of free IPs in the pool
func (dp Mysqlt) FreeIPsRemaining() uint64 {
	dp.DHCPPool_local.lock.Lock()
	defer dp.DHCPPool_local.lock.Unlock()
	return uint64(len(dp.DHCPPool_local.free))
}

// Returns the capacity of the pool
func (dp Mysqlt) Capacity() uint64 {
	return dp.DHCPPool_local.capacity
}

type DHCPPool struct {
	lock     *sync.Mutex
	free     map[uint64]bool
	mac      map[uint64]string
	capacity uint64
}

var Memory Memoryt

type Memoryt struct {
	DHCPPool_local *DHCPPool
}

func NewDHCPPool(capacity uint64) *DHCPPool {
	d := &DHCPPool{
		lock:     &sync.Mutex{},
		free:     make(map[uint64]bool),
		mac:      make(map[uint64]string),
		capacity: capacity,
	}
	for i := uint64(0); i < d.capacity; i++ {
		d.free[i] = true
	}
	return d
}

func (dp Memoryt) NewDHCPPool(capacity uint64) {
	d := &DHCPPool{
		lock:     &sync.Mutex{},
		free:     make(map[uint64]bool),
		mac:      make(map[uint64]string),
		capacity: capacity,
	}
	for i := uint64(0); i < d.capacity; i++ {
		d.free[i] = true
	}
	dp.DHCPPool_local = d
	// return d
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
func (dp Memoryt) ReserveIPIndex(index uint64, mac string) (error, string) {
	dp.DHCPPool_local.lock.Lock()
	defer dp.DHCPPool_local.lock.Unlock()

	if index >= dp.DHCPPool_local.capacity {
		return errors.New("Trying to reserve an IP that is outside the capacity of this pool"), FreeMac
	}

	if _, free := dp.DHCPPool_local.free[index]; free {
		delete(dp.DHCPPool_local.free, index)
		dp.DHCPPool_local.mac[index] = mac
		return nil, mac
	} else {
		return errors.New("IP is already reserved"), FreeMac
	}
}

// Frees an IP in the pool, returns an error if the IP is already free
func (dp Memoryt) FreeIPIndex(index uint64) error {
	dp.DHCPPool_local.lock.Lock()
	defer dp.DHCPPool_local.lock.Unlock()

	if !dp.DHCPPool_local.IndexInPool(index) {
		return errors.New("Trying to free an IP that is outside the capacity of this pool")
	}

	if _, free := dp.DHCPPool_local.free[index]; free {
		return errors.New("IP is already free")
	} else {
		dp.DHCPPool_local.free[index] = true
		delete(dp.DHCPPool_local.mac, index)
		return nil
	}
}

// Check if the IP is free at the index
func (dp Memoryt) IsFreeIPAtIndex(index uint64) bool {
	dp.DHCPPool_local.lock.Lock()
	defer dp.DHCPPool_local.lock.Unlock()

	if !dp.DHCPPool_local.IndexInPool(index) {
		return false
	}

	if _, free := dp.DHCPPool_local.free[index]; free {
		return true
	} else {
		return false
	}
}

// Check if the IP is free at the index
func (dp Memoryt) GetMACIndex(index uint64) (uint64, string, error) {
	dp.DHCPPool_local.lock.Lock()
	defer dp.DHCPPool_local.lock.Unlock()

	if !dp.DHCPPool_local.IndexInPool(index) {
		return index, FreeMac, errors.New("The index is not part of the pool")
	}

	if _, free := dp.DHCPPool_local.free[index]; free {
		return index, FreeMac, nil
	} else {
		return index, dp.DHCPPool_local.mac[index], nil
	}
}

// Returns a random free IP address, an error if the pool is full
func (dp Memoryt) GetFreeIPIndex(mac string) (uint64, string, error) {
	dp.DHCPPool_local.lock.Lock()
	defer dp.DHCPPool_local.lock.Unlock()

	if len(dp.DHCPPool_local.free) == 0 {
		return 0, FreeMac, errors.New("DHCP pool is full")
	}
	index := rand.Intn(len(dp.DHCPPool_local.free))

	var available uint64
	for available = range dp.DHCPPool_local.free {
		if index == 0 {
			break
		}
		index--
	}

	delete(dp.DHCPPool_local.free, available)
	dp.DHCPPool_local.mac[available] = mac

	return available, mac, nil
}

// Returns whether or not a specific index is in the capacity of the pool
func (dp Memoryt) IndexInPool(index uint64) bool {
	return index < dp.DHCPPool_local.capacity
}

// Returns the amount of free IPs in the pool
func (dp Memoryt) FreeIPsRemaining() uint64 {
	dp.DHCPPool_local.lock.Lock()
	defer dp.DHCPPool_local.lock.Unlock()
	return uint64(len(dp.DHCPPool_local.free))
}

// Returns the capacity of the pool
func (dp Memoryt) Capacity() uint64 {
	return dp.DHCPPool_local.capacity
}
