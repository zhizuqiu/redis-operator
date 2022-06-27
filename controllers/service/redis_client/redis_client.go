package redis_client

type RedisParam struct {
	Ip            string
	NameSpace     string
	Name          string
	ContainerName string
}

// Client defines the functions neccesary to connect to redis and sentinel to get or set what we nned
type RedisClient interface {
	GetNumberSentinelsInMemory(redisParam RedisParam) (int32, error)
	GetNumberSentinelSlavesInMemory(sentinel RedisParam) (int32, error)
	ResetSentinel(sentinel RedisParam) error
	GetSlaveOf(redisParam RedisParam, password string) (string, error)
	IsMaster(redisParam RedisParam, password string) (bool, error)
	MonitorRedis(redisParam RedisParam, monitor, quorum, password string) error
	MonitorRedisWithPort(redisParam RedisParam, monitor, port, quorum, password string) error
	MakeMaster(redisParam RedisParam, password string) error
	MakeSlaveOf(redisParam RedisParam, password, masterIP string) error
	MakeSlaveOfWithPort(redisParam RedisParam, password, masterIP, masterPort string) error
	GetSentinelMonitor(redisParam RedisParam) (string, string, error)
	SetCustomSentinelConfig(redisParam RedisParam, configs []string) error
	SetCustomRedisConfig(redisParam RedisParam, configs []string, password string) error
	SetRedisPassword(redisParam RedisParam, newPassword string) error
	SetSentinelPassword(redisParam RedisParam, newPassword string) error
	GetRedisPassword(redisParam RedisParam) (string, error)
}
