package database

type PoolStats struct {
	AcquireCount            int64
	AcquireDurationMS       int64
	AcquiredConnections     int32
	ConstructingConnections int32
	EmptyAcquireCount       int64
	IdleConnections         int32
	MaxConnections          int32
	TotalConnections        int32
}

func (db *Database) Stats() PoolStats {

	stats := db.pool.Stat()

	return PoolStats{
		AcquireCount:            stats.AcquireCount(),
		AcquireDurationMS:       stats.AcquireDuration().Milliseconds(),
		AcquiredConnections:     stats.AcquiredConns(),
		ConstructingConnections: stats.ConstructingConns(),
		EmptyAcquireCount:       stats.EmptyAcquireCount(),
		IdleConnections:         stats.IdleConns(),
		MaxConnections:          stats.MaxConns(),
		TotalConnections:        stats.TotalConns(),
	}
}
