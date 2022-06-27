package redis_client

import (
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"regexp"
	"strconv"
	"strings"
)

const (
	sentinelsNumberREString = "sentinels=([0-9]+)"
	slaveNumberREString     = "slaves=([0-9]+)"
	sentinelStatusREString  = "status=([a-z]+)"
	redisMasterHostREString = "master_host:([0-9.]+)"
	redisRoleMaster         = "role:master"
	redisPort               = "6379"
	sentinelPort            = "26379"
	masterName              = "mymaster"
)

var (
	sentinelNumberRE  = regexp.MustCompile(sentinelsNumberREString)
	sentinelStatusRE  = regexp.MustCompile(sentinelStatusREString)
	slaveNumberRE     = regexp.MustCompile(slaveNumberREString)
	redisMasterHostRE = regexp.MustCompile(redisMasterHostREString)
)

type RedisExecClienter struct {
	Log      logr.Logger
	RedisApi RedisApi
}

// New returns a redis client
func NewRedisExecClienter(log logr.Logger, redisApi RedisApi) RedisClient {
	log = log.WithValues("redisClient", "RedisExecClienter")
	return &RedisExecClienter{
		Log:      log,
		RedisApi: redisApi,
	}
}

func (rc *RedisExecClienter) GetNumberSentinelsInMemory(redisParam RedisParam) (int32, error) {
	info, err := rc.RedisApi.sentinelInfo(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, "sentinel")
	if err != nil {
		return 0, err
	}
	if err2 := isSentinelReady(info); err2 != nil {
		return 0, err2
	}
	match := sentinelNumberRE.FindStringSubmatch(info)
	if len(match) == 0 {
		return 0, errors.New("Seninel regex not found")
	}
	nSentinels, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}
	return int32(nSentinels), nil
}

func isSentinelReady(info string) error {
	matchStatus := sentinelStatusRE.FindStringSubmatch(info)
	if len(matchStatus) == 0 || matchStatus[1] != "ok" {
		return errors.New("Sentinels not ready")
	}
	return nil
}

func (rc *RedisExecClienter) GetNumberSentinelSlavesInMemory(redisParam RedisParam) (int32, error) {
	info, err := rc.RedisApi.sentinelInfo(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, "sentinel")
	if err != nil {
		return 0, err
	}
	if err2 := isSentinelReady(info); err2 != nil {
		return 0, err2
	}
	match := slaveNumberRE.FindStringSubmatch(info)
	if len(match) == 0 {
		return 0, errors.New("Slaves regex not found")
	}
	nSlaves, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}
	return int32(nSlaves), nil
}

func (rc *RedisExecClienter) ResetSentinel(sentinel RedisParam) error {
	_, err := rc.RedisApi.sentinelReset(sentinel.NameSpace, sentinel.Name, sentinel.ContainerName)
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisExecClienter) GetSlaveOf(redisParam RedisParam, password string) (string, error) {
	info, err := rc.RedisApi.info(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, password, "replication")
	if err != nil {
		return "", err
	}
	match := redisMasterHostRE.FindStringSubmatch(info)
	if len(match) == 0 {
		return "", nil
	}
	return match[1], nil
}

func (rc *RedisExecClienter) IsMaster(redisParam RedisParam, password string) (bool, error) {
	info, err := rc.RedisApi.info(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, password, "replication")
	if err != nil {
		return false, err
	}
	return strings.Contains(info, redisRoleMaster), nil
}

func (rc *RedisExecClienter) MonitorRedis(redisParam RedisParam, monitor, quorum, password string) error {
	return rc.MonitorRedisWithPort(redisParam, monitor, redisPort, quorum, password)
}

func (rc *RedisExecClienter) MonitorRedisWithPort(redisParam RedisParam, monitor, port, quorum, password string) error {
	_, err := rc.RedisApi.sentinelRemoveMaster(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName)
	if err != nil {
		return err
	}
	_, err = rc.RedisApi.sentinelMonitorRedis(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, monitor, port, quorum)
	if err != nil {
		return err
	}
	if password != "" {
		_, err = rc.RedisApi.sentinelSetPassword(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, password)
		if err != nil {
			return err
		}
	}
	return nil
}

func (rc *RedisExecClienter) MakeMaster(redisParam RedisParam, password string) error {
	_, err := rc.RedisApi.makeMaster(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, password)
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisExecClienter) MakeSlaveOf(redisParam RedisParam, password, masterIP string) error {
	return rc.MakeSlaveOfWithPort(redisParam, password, masterIP, redisPort)
}

func (rc *RedisExecClienter) MakeSlaveOfWithPort(redisParam RedisParam, password, masterIP, masterPort string) error {
	_, err := rc.RedisApi.slaveOf(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, password, masterIP, masterPort)
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisExecClienter) GetSentinelMonitor(redisParam RedisParam) (string, string, error) {
	output, err := rc.RedisApi.sentinelMonitor(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName)
	if err != nil {
		return "", "", err
	}
	res := strings.Split(output, "\n")
	masterIP := res[3]
	masterPort := res[5]
	return masterIP, masterPort, nil
}

func (rc *RedisExecClienter) SetCustomSentinelConfig(redisParam RedisParam, configs []string) error {
	for _, config := range configs {
		param, value, err := rc.getConfigParameters(config)
		if err != nil {
			return err
		}
		if _, err := rc.RedisApi.applySentinelConfig(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, param, value); err != nil {
			return err
		}
	}
	return nil
}

func (rc *RedisExecClienter) SetCustomRedisConfig(redisParam RedisParam, configs []string, password string) error {
	for _, config := range configs {
		param, value, err := rc.getConfigParameters(config)
		if err != nil {
			return err
		}

		if _, err = rc.RedisApi.applyRedisConfig(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, password, param, value); err != nil {
			return err
		}
	}
	return nil
}

func (rc *RedisExecClienter) getConfigParameters(config string) (parameter string, value string, err error) {
	s := strings.Split(config, " ")
	if len(s) < 2 {
		return "", "", fmt.Errorf("configuration '%s' malformed", config)
	}
	return s[0], strings.Join(s[1:], " "), nil
}

func (rc *RedisExecClienter) SetRedisPassword(redisParam RedisParam, newPassword string) error {

	oldPassword, err := rc.RedisApi.getRedisClientPassword(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName)
	if err != nil {
		return err
	}
	_, err = rc.RedisApi.setRedisMasterauthPassword(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, oldPassword, newPassword)
	if err != nil {
		return err
	}

	_, err = rc.RedisApi.setRedisRequirepassPassword(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, oldPassword, newPassword)
	if err != nil {
		return err
	}

	return nil
}

func (rc *RedisExecClienter) SetSentinelPassword(redisParam RedisParam, newPassword string) error {
	_, err := rc.RedisApi.sentinelSetPassword(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName, newPassword)
	if err != nil {
		return err
	}
	return nil
}

func (rc *RedisExecClienter) GetRedisPassword(redisParam RedisParam) (string, error) {
	password, err := rc.RedisApi.getRedisClientPassword(redisParam.NameSpace, redisParam.Name, redisParam.ContainerName)
	if err != nil {
		return "", err
	}
	return strings.Split(password, "\n")[0], err
}

func EscapeRedisPassword(pass string) string {
	passResult := ""
	for i := 0; i < len(pass); i++ {
		ch := pass[i]
		if ch == '$' {
			passResult = passResult + "\\" + string(ch)
		} else {
			passResult = passResult + string(ch)
		}
	}
	return passResult
}
