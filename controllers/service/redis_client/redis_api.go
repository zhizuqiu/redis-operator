package redis_client

import (
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/zhizuqiu/redis-operator/controllers/service/exec"
	"github.com/zhizuqiu/redis-operator/controllers/util"
	"strings"
)

type RedisApi interface {
	info(namespace, podName, containerName, password, section string) (string, error)
	makeMaster(namespace, podName, containerName, password string) (string, error)
	slaveOf(namespace, podName, containerName, password, masterIP, masterPort string) (string, error)
	sentinelMonitor(namespace, podName, containerName string) (string, error)
	sentinelRemoveMaster(namespace, podName, containerName string) (string, error)
	sentinelRemoveERRCanIgnore(output string) bool
	sentinelMonitorRedis(namespace, podName, containerName, monitor, port, quorum string) (string, error)
	sentinelSetPassword(namespace, podName, containerName, password string) (string, error)
	sentinelInfo(namespace, podName, containerName, section string) (string, error)
	sentinelReset(namespace, podName, containerName string) (string, error)
	applyRedisConfig(namespace, podName, containerName, password, parameter, value string) (string, error)
	applySentinelConfig(namespace, podName, containerName, parameter, value string) (string, error)
	rewriteRedisConfig(namespace, podName, containerName, password string) (string, error)
	getRedisClientPassword(namespace, podName, containerName string) (string, error)
	setRedisMasterauthPassword(namespace, podName, containerName, oldPassword, newPassword string) (string, error)
	setRedisRequirepassPassword(namespace, podName, containerName, oldPassword, newPassword string) (string, error)
}

type RedisExecApi struct {
	Log            logr.Logger
	Execer         exec.IExec
	RedisExport    string
	SentinelExport string
}

// New returns a redis client
func NewRedisExecApi(log logr.Logger, execer exec.IExec) RedisApi {
	log = log.WithValues("redisClient", "RedisExecApi")
	return &RedisExecApi{
		Log:            log,
		Execer:         execer,
		RedisExport:    "export REDIS_PORT=$(cat /data/conf/redis.conf | grep port | awk '{print $2}') && ",
		SentinelExport: "export REDIS_PORT=$(cat /data/conf/sentinel.conf | grep port | awk '{print $2}') && ",
	}
}

func (r *RedisExecApi) info(namespace, podName, containerName, password, section string) (string, error) {
	password = EscapeRedisPassword(password)

	var command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" info " + section
	if password != "" {
		command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" --no-auth-warning -a " + password + " info " + section
	}

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		return output, nil
	}
}

func (r *RedisExecApi) makeMaster(namespace, podName, containerName, password string) (string, error) {
	password = EscapeRedisPassword(password)

	var command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" SLAVEOF NO ONE"
	if password != "" {
		command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" --no-auth-warning -a " + password + " SLAVEOF NO ONE"
	}

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !isOk(output) {
			return output, errors.New("SLAVEOF NO ONE err: " + output)
		}
		_, err := r.rewriteRedisConfig(namespace, podName, containerName, password)
		if err != nil {
			Error2(err.Error(), namespace, podName)
		}
		return output, nil
	}
}

func (r *RedisExecApi) slaveOf(namespace, podName, containerName, password, masterIP, masterPort string) (string, error) {
	password = EscapeRedisPassword(password)

	var command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" SLAVEOF " + masterIP + " " + masterPort
	if password != "" {
		command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" --no-auth-warning -a " + password + " SLAVEOF " + masterIP + " " + masterPort
	}

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !isSlaveOfOk(output) {
			return output, errors.New("SLAVEOF " + masterIP + " " + masterPort + " err: " + output)
		}
		_, err := r.rewriteRedisConfig(namespace, podName, containerName, password)
		if err != nil {
			Error2(err.Error(), namespace, podName)
		}
		return output, nil
	}
}

func isSlaveOfOk(output string) bool {
	if output == "" {
		return false
	}
	outputs := strings.Split(output, "\n")
	if len(outputs) > 0 {
		if strings.Contains(outputs[0], "OK") && strings.Index(outputs[0], "OK") == 0 {
			return true
		}
	} else {
		return false
	}
	return false
}

func (r *RedisExecApi) sentinelMonitor(namespace, podName, containerName string) (string, error) {
	var command = r.SentinelExport + "redis-cli -p \"${REDIS_PORT}\" SENTINEL master " + masterName

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !isSentinelMasterSuccess(output) {
			return output, errors.New("SENTINEL master " + masterName + " err: " + output)
		}
		return output, nil
	}
}

func isSentinelMasterSuccess(output string) bool {
	if output == "" {
		return false
	}
	outputs := strings.Split(output, "\n")
	if len(outputs) > 0 {
		if outputs[0] == "name" {
			return true
		}
	} else {
		return false
	}
	return false
}

func (r *RedisExecApi) sentinelRemoveMaster(namespace, podName, containerName string) (string, error) {
	var command = r.SentinelExport + "redis-cli -p \"${REDIS_PORT}\" SENTINEL REMOVE " + masterName

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !r.sentinelRemoveERRCanIgnore(output) {
			return output, errors.New("SENTINEL REMOVE err: " + output)
		}
		return output, nil
	}
}

func (r *RedisExecApi) sentinelRemoveERRCanIgnore(output string) bool {
	if output == "" {
		return false
	}
	outputs := strings.Split(output, "\n")
	if len(outputs) > 0 {
		if outputs[0] == "OK" {
			return true
		}
		if outputs[0] == "ERR No such master with that name" {
			return true
		}
	} else {
		return false
	}
	return false
}

func isOk(output string) bool {
	if output == "" {
		return false
	}
	outputs := strings.Split(output, "\n")
	if len(outputs) > 0 {
		if outputs[0] == "OK" {
			return true
		}
	} else {
		return false
	}
	return false
}

func (r *RedisExecApi) sentinelMonitorRedis(namespace, podName, containerName, monitor, port, quorum string) (string, error) {
	var command = r.SentinelExport + "redis-cli -p \"${REDIS_PORT}\" SENTINEL MONITOR " + masterName + " " + monitor + " " + port + " " + quorum

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !isOk(output) {
			return output, errors.New("SENTINEL MONITOR err: " + output)
		}
		return output, nil
	}
}

func (r *RedisExecApi) sentinelSetPassword(namespace, podName, containerName, password string) (string, error) {
	password = EscapeRedisPassword(password)

	var command = r.SentinelExport + "redis-cli -p \"${REDIS_PORT}\" SENTINEL SET " + masterName + " auth-pass \"" + password + "\""

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !isOk(output) {
			return output, errors.New("SENTINEL set auth-pass err: " + output)
		}
		return output, nil
	}
}

func (r *RedisExecApi) sentinelInfo(namespace, podName, containerName, section string) (string, error) {
	var command = r.SentinelExport + "redis-cli -p \"${REDIS_PORT}\" info " + section

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		return output, nil
	}
}

func (r *RedisExecApi) sentinelReset(namespace, podName, containerName string) (string, error) {
	var command = r.SentinelExport + "redis-cli -p \"${REDIS_PORT}\" SENTINEL reset \"*\""

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !hasBeenReset(output) {
			return output, errors.New("SENTINEL reset * err: " + output)
		}
		return output, nil
	}
}

func hasBeenReset(output string) bool {

	if output == "" {
		return false
	}
	outputs := strings.Split(output, "\n")
	if len(outputs) > 0 {
		if IsDigit(outputs[0]) {
			return true
		}
	} else {
		return false
	}
	return false
}

func isSingleDigit(data string) bool {
	digit := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	for _, item := range digit {
		if data == item {
			return true
		}
	}
	return false
}

func IsDigit(data string) bool {
	for _, item := range data {
		if isSingleDigit(string(item)) {
			continue
		} else {
			return false
		}
	}
	return true
}

func (r *RedisExecApi) applyRedisConfig(namespace, podName, containerName, password, parameter, value string) (string, error) {
	password = EscapeRedisPassword(password)

	var command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" CONFIG SET " + parameter + " " + value
	if password != "" {
		command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" --no-auth-warning -a " + password + " CONFIG SET " + parameter + " " + value
	}

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !isOk(output) {
			return output, errors.New("REDIS CONFIG SET err: " + output)
		}
		if parameter == "requirepass" {
			_, err := r.rewriteRedisConfig(namespace, podName, containerName, value)
			if err != nil {
				Error2(err.Error(), namespace, podName)
			}
		} else {
			_, err := r.rewriteRedisConfig(namespace, podName, containerName, password)
			if err != nil {
				Error2(err.Error(), namespace, podName)
			}
		}

		return output, nil
	}
}

func (r *RedisExecApi) applySentinelConfig(namespace, podName, containerName, parameter, value string) (string, error) {
	var command = r.SentinelExport + "redis-cli -p \"${REDIS_PORT}\" SENTINEL SET " + masterName + " " + parameter + " " + value

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !isOk(output) {
			return output, errors.New("SENTINEL CONFIG SET err: " + output)
		}
		return output, nil
	}
}

func (r *RedisExecApi) rewriteRedisConfig(namespace, podName, containerName, password string) (string, error) {
	password = EscapeRedisPassword(password)

	var command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" CONFIG REWRITE"
	if password != "" {
		command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" --no-auth-warning -a " + password + " CONFIG REWRITE"
	}

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !isOk(output) {
			return output, errors.New("CONFIG REWRITE err: " + output)
		}
		return output, nil
	}
}

func (r *RedisExecApi) getRedisClientPassword(namespace, podName, containerName string) (string, error) {
	var command = "cat " + util.GetRedisConfigWritablePath() + " | grep requirepass | awk -F\\\" '{print $2}'"

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		return strings.Split(output, "\n")[0], nil
	}
}

func (r *RedisExecApi) setRedisMasterauthPassword(namespace, podName, containerName, oldPassword, newPassword string) (string, error) {
	var command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" CONFIG SET masterauth \"" + newPassword + "\""
	if oldPassword != "" {
		command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" --no-auth-warning -a " + oldPassword + " CONFIG SET masterauth \"" + newPassword + "\""
	}

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !isOk(output) {
			return output, errors.New("REDIS CONFIG SET err: " + output)
		}
		_, err := r.rewriteRedisConfig(namespace, podName, containerName, oldPassword)
		if err != nil {
			Error2(err.Error(), namespace, podName)
		}
		return output, nil
	}
}

func (r *RedisExecApi) setRedisRequirepassPassword(namespace, podName, containerName, oldPassword, newPassword string) (string, error) {
	var command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" CONFIG SET requirepass \"" + newPassword + "\""
	if oldPassword != "" {
		command = r.RedisExport + "redis-cli -p \"${REDIS_PORT}\" --no-auth-warning -a " + oldPassword + " CONFIG SET requirepass \"" + newPassword + "\""
	}

	output, stderr, err := r.Execer.ExecCommandInContainerWithFullOutputBySh(namespace, podName, containerName, command)

	if len(stderr) != 0 {
		fmt.Println("STDERR:", stderr, containerName, podName, namespace)
	}
	if err != nil {
		return "", err
	} else {
		if !isOk(output) {
			return output, errors.New("REDIS CONFIG SET err: " + output)
		}
		_, err := r.rewriteRedisConfig(namespace, podName, containerName, newPassword)
		if err != nil {
			Error2(err.Error(), namespace, podName)
		}
		return output, nil
	}
}
