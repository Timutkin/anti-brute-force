package service

import (
	"errors"
	"net"
	"sync"
)

var ErrParseCIDR = errors.New("failed to parse CIDR")
var ErrParseIP = errors.New("failed to parse IP")

type ImMemoryListService struct {
	list map[string]*net.IPNet
	rwm  *sync.RWMutex
}

func NewInMemoryListService() *ImMemoryListService {
	return &ImMemoryListService{
		list: make(map[string]*net.IPNet),
		rwm:  &sync.RWMutex{},
	}
}

func (b *ImMemoryListService) AddCIDR(CIDR string) error {
	_, network, err := net.ParseCIDR(CIDR)
	if err != nil {
		return ErrParseCIDR
	}
	b.rwm.Lock()
	defer b.rwm.Unlock()
	b.list[CIDR] = network
	return nil
}

func (b *ImMemoryListService) DeleteCIDR(CIDR string) error {
	b.rwm.Lock()
	defer b.rwm.Unlock()
	delete(b.list, CIDR)
	return nil
}

func (b *ImMemoryListService) GetCIDRs() []string {
	b.rwm.Lock()
	defer b.rwm.Unlock()
	list := b.list
	result := make([]string, 0, len(list))
	for _, v := range b.list {
		result = append(result, v.String())
	}
	return result
}

func (b *ImMemoryListService) IsIPContains(ip string) (bool, error) {
	IP := net.ParseIP(ip)
	if IP == nil {
		return false, ErrParseIP
	}
	b.rwm.RLock()
	defer b.rwm.RUnlock()
	for _, ipNet := range b.list {
		if ipNet.Contains(IP) {
			return true, nil
		}
	}
	return false, nil
}
