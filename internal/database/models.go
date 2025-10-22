package database

import "time"

// AxeOSMetric represents a single metric collection from an AxeOS miner
type AxeOSMetric struct {
	Timestamp      time.Time
	InstanceID     string
	InstanceName   string
	Hashrate       float64
	Temperature    float64
	Power          float64
	FanSpeed       int
	BestDiff       string
	SharesAccepted int
	SharesRejected int
	Frequency      int
	Voltage        float64
	CoreVoltage    float64
}

// PoolMetric represents a single metric collection from a Mining Core pool
type PoolMetric struct {
	Timestamp        time.Time
	PoolID           string
	PoolName         string
	PoolHashrate     float64
	PoolWorkers      int
	NetworkHashrate  float64
	NetworkDifficulty float64
	LastBlockTime    *time.Time
	BlocksFound      int
}

// NodeMetric represents a single metric collection from a crypto node
type NodeMetric struct {
	Timestamp       time.Time
	NodeID          string
	NodeName        string
	BlockHeight     int
	Connections     int
	Difficulty      float64
	NetworkHashrate float64
}
