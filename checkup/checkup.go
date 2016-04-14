package checkup

import (
	"errors"
	"net"
	"sync"
	"time"
)

type UpdownCallback func(host string, isAlive bool)

const (
	DefaultTimeout  = 2 * time.Second
	DefaultInterval = 5 * time.Second
)

var (
	defaultCallback = func(host string, isAlive bool) {}

	AlreadyRunningError = errors.New("Checkup is already running, you cannot start it again")
	AlreadyStoppedError = errors.New("Checkup is already stopped, you cannot stop it again")
)

func NewTCPCheckup(hosts []string) Checkup {
	return NewCheckup(hosts, "tcp")
}

func NewUDPCheckup(hosts []string) Checkup {
	return NewCheckup(hosts, "udp")
}

// NewCheckup creates a new Checkup that will check the liveness of the given
// hosts on the given protocol.
// netProto is one of the protocols the net package recognizes, including:
// "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only), "udp", "udp4" (IPv4-only),
// "udp6" (IPv6-only), "ip", "ip4" (IPv4-only), "ip6" (IPv6-only), "unix",
// "unixgram" and "unixpacket".
func NewCheckup(hosts []string, netProto string) Checkup {
	var checkup checkupInternal
	checkup.timeout = DefaultTimeout
	checkup.interval = DefaultInterval
	checkup.deadAlive = make(map[string]bool)
	checkup.netProtocol = netProto
	checkup.stateChangeCallback = defaultCallback

	for _, host := range hosts {
		checkup.deadAlive[host] = false
	}

	return &checkup
}

type Checkup interface {
	GetAliveHosts() []string
	GetDeadHosts() []string
	GetTimeout() time.Duration
	GetPingInterval() time.Duration
	SetTimeout(duration time.Duration)
	SetPingInterval(duration time.Duration)
	Start() error
	Stop() error
	SetStateChangeHandler(callback UpdownCallback)
}

type checkupInternal struct {
	timeout             time.Duration
	interval            time.Duration
	deadAlive           map[string]bool
	stateChangeCallback UpdownCallback
	mutex               sync.RWMutex
	netProtocol         string
	closeChannel        chan bool
}

func (c *checkupInternal) GetAliveHosts() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	var alive []string

	for host, isAlive := range c.deadAlive {
		if isAlive {
			alive = append(alive, host)
		}
	}

	return alive
}

func (c *checkupInternal) GetDeadHosts() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	var dead []string

	for host, isAlive := range c.deadAlive {
		if !isAlive {
			dead = append(dead, host)
		}
	}

	return dead
}

func (c *checkupInternal) GetTimeout() time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.timeout
}

func (c *checkupInternal) GetPingInterval() time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.interval
}

func (c *checkupInternal) GetCheckInterval() time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.interval
}

func (c *checkupInternal) SetTimeout(timeout time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.timeout = timeout
}

func (c *checkupInternal) SetPingInterval(interval time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.interval = interval
}

func (c *checkupInternal) Start() error {

	if c.closeChannel != nil {
		return AlreadyRunningError
	}

	c.closeChannel = make(chan bool)

	for host, _ := range c.deadAlive {
		go c.hostCheck(host)
	}

	return nil
}

func (c *checkupInternal) hostCheck(host string) {

	for {

		currentTimeout := c.GetTimeout()
		currentInterval := c.GetPingInterval()

		select {
		case <-time.After(currentInterval):
			conn, err := net.DialTimeout(c.netProtocol, host, currentTimeout)
			c.mutex.RLock()
			current := c.deadAlive[host]
			c.mutex.RUnlock()

			newstate := err == nil
			if err == nil {
				conn.Close()
			}

			if current != newstate {
				c.mutex.Lock()
				c.deadAlive[host] = newstate
				c.mutex.Unlock()

				c.mutex.RLock()
				c.stateChangeCallback(host, newstate)
				c.mutex.RUnlock()
			}

		case <-c.closeChannel:
			return
		}
	}
}

func (c *checkupInternal) Stop() error {
	if c.closeChannel == nil {
		return AlreadyStoppedError
	}

	close(c.closeChannel)
	c.closeChannel = nil

	return nil
}

func (c *checkupInternal) SetStateChangeHandler(callback UpdownCallback) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.stateChangeCallback = callback
}
