package pool

import (
	"database/sql"
	"errors"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type Mysql struct {
	PoolName string
	DHCPPool *DHCPPool
	SQL      *sql.DB
}

func NewMysqlPool(capacity uint64, name string, sql *sql.DB) (PoolBackend, error) {
	dp := Mysql{}
	dp.PoolName = name
	dp.SQL = sql
	dp.NewDHCPPool(capacity)
	return &dp, nil
}

func (dp *Mysql) NewDHCPPool(capacity uint64) {
	d := &DHCPPool{
		lock:     &sync.Mutex{},
		free:     make(map[uint64]bool),
		mac:      make(map[uint64]string),
		capacity: capacity,
	}
	for i := uint64(0); i < capacity; i++ {
		// Need to test err
		rows, _ := dp.SQL.Query("INSERT INTO dhcppool (pool_name, index) VALUES (?, ?) ON DUPLICATE KEY UPDATE id=id", dp.PoolName, i)
		defer rows.Close()
	}
	dp.DHCPPool = d
}

func (dp *Mysql) GetDHCPPool() DHCPPool {
	return *dp.DHCPPool
}

// Reserves an IP in the pool, returns an error if the IP has already been reserved
func (dp *Mysql) ReserveIPIndex(index uint64, mac string) (error, string) {

	if index >= dp.DHCPPool.capacity {
		return errors.New("Trying to reserve an IP that is outside the capacity of this pool"), FreeMac
	}
	query := "UPDATE dhcppool SET free = 0, mac = $3 WHERE index = $1 AND free = 1 AND pool_name = $2"
	res, err := dp.SQL.Exec(query, index, dp.PoolName, mac)

	if err != nil {
		return errors.New("IP is already reserved"), FreeMac
	} else {
		count, err2 := res.RowsAffected()
		if err2 != nil {
			return errors.New("IP is already reserved"), FreeMac
		} else {
			if count == 1 {
				return nil, mac
			} else {
				return errors.New("IP is already reserved"), FreeMac
			}
		}
	}
}

// Frees an IP in the pool, returns an error if the IP is already free
func (dp *Mysql) FreeIPIndex(index uint64) error {

	if !dp.IndexInPool(index) {
		return errors.New("Trying to free an IP that is outside the capacity of this pool")
	}

	query := "UPDATE dhcppool set free = 1, mac = $3, released = NOW() WHERE index = $1 AND free = 0 AND pool_name = $2"
	res, err := dp.SQL.Exec(query, index, dp.PoolName, FreeMac)

	if err != nil {
		return errors.New("IP is already free")
	} else {
		count, err2 := res.RowsAffected()
		if err2 != nil {
			return errors.New("IP is already free")
		} else {
			if count == 1 {
				return nil
			} else {
				return errors.New("IP is already free")
			}
		}
	}
}

// Check if the IP is free at the index
func (dp *Mysql) IsFreeIPAtIndex(index uint64) bool {

	if !dp.IndexInPool(index) {
		return false
	}

	query := "SELECT free FROM dhcppool WHERE free = 1 AND index = $1 AND pool_name = $2"
	res, err := dp.SQL.Exec(query, index, dp.PoolName)

	if err != nil {
		return false
	} else {
		count, err2 := res.RowsAffected()
		if err2 != nil {
			return false
		} else {
			if count == 1 {
				return true
			} else {
				return false
			}
		}
	}
}

// Check if the IP is free at the index
func (dp *Mysql) GetMACIndex(index uint64) (uint64, string, error) {

	if !dp.IndexInPool(index) {
		return index, FreeMac, errors.New("The index is not part of the pool")
	}

	rows, err := dp.SQL.Query("SELECT index, mac FROM dhcppool WHERE index = ? AND pool_name = ?", index, dp.PoolName)
	defer rows.Close()
	if err != nil {
		return index, FreeMac, nil
	}
	var (
		Index int
		mac   string
	)
	for rows.Next() {
		err := rows.Scan(&Index, &mac)
		if err != nil {
			return index, FreeMac, nil
		}
	}
	return uint64(Index), mac, nil
}

// Returns a random free IP address, an error if the pool is full
func (dp *Mysql) GetFreeIPIndex(mac string) (uint64, string, error) {

	// Pool is full ?
	rows, err := dp.SQL.Query("SELECT COUNT(*) FROM dhcppool WHERE free = 1 AND pool_name = ?", dp.PoolName)
	defer rows.Close()

	if err != nil {
		return 0, FreeMac, errors.New("DHCP pool is full")
	}
	var (
		Count int
	)
	for rows.Next() {
		err := rows.Scan(&Count)
		if err != nil {
			return 0, FreeMac, errors.New("DHCP pool is full")
		}
		if Count == 0 {
			return 0, FreeMac, errors.New("DHCP pool is full")
		}
	}

	// Search for available index

	tx, err := dp.SQL.Begin()

	if err != nil {
		// log.Fatal(err)
	}

	query := "UPDATE dhcppool SET mac = $1, free = 0 WHERE index = (SELECT index FROM dhcppool WHERE free = 1 AND pool_name = $2 ORDER BY RAND() LIMIT 1) AND @tmp_index := index"
	res, err := tx.Exec(query, mac, dp.PoolName)

	if err != nil {
		return 0, FreeMac, errors.New("DHCP pool is full")

	} else {
		count, err2 := res.RowsAffected()
		if err2 != nil {
			return 0, FreeMac, errors.New("DHCP pool is full")
		} else {
			if count == 1 {
				query = "SELECT @tmp_index"
				rows, err = tx.Query(query)
				var (
					Index int
				)
				for rows.Next() {
					err := rows.Scan(&Index)
					if err != nil {
						return uint64(Index), mac, nil
					}
				}
				return 0, FreeMac, errors.New("DHCP pool is full")
				// return true
			} else {
				return 0, FreeMac, errors.New("DHCP pool is full")
			}
		}
	}
}

// Returns whether or not a specific index is in the capacity of the pool
func (dp *Mysql) IndexInPool(index uint64) bool {
	return index < dp.DHCPPool.capacity
}

// Returns the amount of free IPs in the pool
func (dp *Mysql) FreeIPsRemaining() uint64 {
	dp.DHCPPool.lock.Lock()
	defer dp.DHCPPool.lock.Unlock()
	return uint64(len(dp.DHCPPool.free))
}

// Returns the capacity of the pool
func (dp *Mysql) Capacity() uint64 {
	return dp.DHCPPool.capacity
}

// Compare what we have in the cache with what we have in the pool
func (dp *Mysql) GetIssues(macs []string) ([]string, map[uint64]string) {
	var inPoolNotInCache []string
	var duplicateInPool map[uint64]string
	duplicateInPool = make(map[uint64]string)

	// var count int
	// var saveindex uint64
	// for i := uint64(0); i < dp.DHCPPool.capacity; i++ {
	// 	if dp.DHCPPool.free[i] {
	// 		continue
	// 	}
	// 	for _, mac := range macs {
	// 		if dp.DHCPPool.mac[i] == mac {
	// 			found = true
	// 		}
	// 	}
	// 	if !found {
	// 		inPoolNotInCache = append(inPoolNotInCache, dp.DHCPPool.mac[i]+", "+strconv.Itoa(int(i)))
	// 	}
	// }
	// for _, mac := range macs {
	// 	count = 0
	// 	saveindex = 0
	//
	// 	for i := uint64(0); i < dp.DHCPPool.capacity; i++ {
	// 		if dp.DHCPPool.free[i] {
	// 			continue
	// 		}
	// 		if dp.DHCPPool.mac[i] == mac {
	// 			if count == 0 {
	// 				saveindex = i
	// 			}
	// 			if count == 1 {
	// 				duplicateInPool[saveindex] = mac
	// 				duplicateInPool[i] = mac
	// 			} else if count > 1 {
	// 				duplicateInPool[i] = mac
	// 			}
	// 			count++
	// 		}
	// 	}
	// }
	return inPoolNotInCache, duplicateInPool
}
