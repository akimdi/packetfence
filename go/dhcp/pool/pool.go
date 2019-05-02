package pool

import (
	"database/sql"
	"fmt"
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
	FreeIPIndex(index uint64) error
	Capacity() uint64
	GetDHCPPool() DHCPPool
	GetIssues(macs []string) ([]string, map[uint64]string)
}

type PoolCreater func(uint64, string, *sql.DB) (PoolBackend, error)

var poolLookup = map[string]PoolCreater{
	"memory": NewMemoryPool,
	"mysql":  NewMysqlPool,
}

func CreatePool(poolType string, capacity uint64, name string, sql *sql.DB) (PoolBackend, error) {
	if creater, found := poolLookup[poolType]; found {
		return creater(capacity, name, sql)
	}

	return nil, fmt.Errorf("Pool of %s not found", poolType)
}

type DHCPPool struct {
	lock     *sync.Mutex
	free     map[uint64]bool
	mac      map[uint64]string
	capacity uint64
}
