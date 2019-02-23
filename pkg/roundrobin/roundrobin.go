package roundrobin

import (
	"fmt"
	"sync"
)

var defaultWeight = 1

type WeightedRoundRobin struct {
	mutex         *sync.Mutex
	servers       []*server
	index         int
	currentWeight int
}

type server struct {
	url    string
	weight int
}

type ServerOption func(*server) error

func Weight(w int) ServerOption {
	return func(s *server) error {
		if w < 0 {
			return fmt.Errorf("Weight should be >= 0")
		}
		s.weight = w
		return nil
	}
}

func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{
		mutex:   &sync.Mutex{},
		servers: []*server{},
		index:   -1, // For proper WRR the index starts from -1
	}
}

func (wrr *WeightedRoundRobin) AppendServer(url string, options ...ServerOption) error {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	if url == "" {
		return fmt.Errorf("server URL can't be empty")
	}

	srv := &server{url: url}
	for _, o := range options {
		if err := o(srv); err != nil {
			return err
		}
	}

	if srv.weight == 0 {
		srv.weight = defaultWeight
	}

	wrr.servers = append(wrr.servers, srv)

	// Reset current state
	wrr.reset()

	return nil
}

func (wrr *WeightedRoundRobin) reset() {
	wrr.index = -1
	wrr.currentWeight = 0
}

func (wrr *WeightedRoundRobin) nextServer() (*server, error) {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	if len(wrr.servers) == 0 {
		return nil, fmt.Errorf("no servers in the pool")
	}

	// GCD across all servers
	gcd := wrr.weightGcd()
	// Maximum weight across all servers
	max := wrr.maxWeight()

	// WRR algo
	for {
		wrr.index = (wrr.index + 1) % len(wrr.servers)
		if wrr.index == 0 {
			wrr.currentWeight = wrr.currentWeight - gcd
			if wrr.currentWeight <= 0 {
				wrr.currentWeight = max
				if wrr.currentWeight == 0 {
					return nil, fmt.Errorf("all servers have 0 weight")
				}
			}
		}
		srv := wrr.servers[wrr.index]
		if srv.weight >= wrr.currentWeight {
			return srv, nil
		}
	}
}

func (wrr *WeightedRoundRobin) weightGcd() int {
	divisor := -1
	for _, s := range wrr.servers {
		if divisor == -1 {
			divisor = s.weight
		} else {
			divisor = gcd(divisor, s.weight)
		}
	}
	return divisor
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func (wrr *WeightedRoundRobin) maxWeight() int {
	max := -1
	for _, s := range wrr.servers {
		if s.weight > max {
			max = s.weight
		}
	}
	return max
}

func (wrr *WeightedRoundRobin) Next() (string, error) {
	srv, err := wrr.nextServer()
	if err != nil {
		return "", err
	}
	return srv.url, err
}
