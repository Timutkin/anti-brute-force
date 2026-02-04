package service

import (
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
	"github.com/timutkin/anti-brute-force/internal/config"
)

type ListService interface {
	IsIPContains(ip string) (bool, error)
}

type InMemoryBruteForceService struct {
	blackListService   ListService
	whiteListService   ListService
	loginMap           map[string]*Bucket
	passwordMap        map[string]*Bucket
	ipMap              map[string]*Bucket
	loginMaxAttempt    int
	passwordMaxAttempt int
	ipMaxAttempt       int
	scheduler          gocron.Scheduler
	mu                 *sync.RWMutex
}

type Bucket struct {
	firstTime time.Time
	lastTime  time.Time
	attempt   int
	interval  time.Duration
	mu        *sync.Mutex
}

type BucketsView struct {
	Name    string       `json:"name"`
	Buckets []BucketView `json:"buckets"`
}

type BucketView struct {
	Key       string    `json:"key"`
	Attempts  int       `json:"attempts"`
	FirstTime time.Time `json:"firstTime"`
	LastTime  time.Time `json:"lastTime"`
}

func createBucket() *Bucket {
	now := time.Now()
	return &Bucket{
		firstTime: now,
		lastTime:  now,
		attempt:   1,
		mu:        &sync.Mutex{},
		interval:  time.Minute,
	}
}

func (i *InMemoryBruteForceService) GetLoginBuckets() []BucketsView {
	i.mu.RLock()
	defer i.mu.RUnlock()
	views := make([]BucketsView, 0, 1)
	loginBuckets := make([]BucketView, 0, len(i.loginMap))
	for k, v := range i.loginMap {
		loginBuckets = append(loginBuckets, BucketView{
			Key:       k,
			Attempts:  v.attempt,
			FirstTime: v.firstTime,
			LastTime:  v.lastTime,
		})
	}
	views = append(views, BucketsView{
		Name:    "login",
		Buckets: loginBuckets,
	})
	return views
}

func (i *InMemoryBruteForceService) DeleteBuckets(login, ip string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if login != "" {
		delete(i.loginMap, login)
	}
	if ip != "" {
		delete(i.ipMap, ip)
	}
}

func (i *InMemoryBruteForceService) IsCredentialAllowed(login, password, ip string) (bool, error) {
	isBlackList, err := i.blackListService.IsIPContains(ip)
	if err != nil {
		return false, err
	}
	if isBlackList {
		return false, nil
	}
	isWhiteList, err := i.whiteListService.IsIPContains(ip)
	if err != nil {
		return false, err
	}
	if isWhiteList {
		return true, nil
	}
	return i.checkAndIncreaseAttemptsForCredentials(login, password, ip), nil
}

func (i *InMemoryBruteForceService) checkAndIncreaseAttemptsForCredentials(login, password, ip string) bool {
	return i.checkAttemptsForCredential(login, i.loginMap, i.loginMaxAttempt) &&
		i.checkAttemptsForCredential(password, i.passwordMap, i.passwordMaxAttempt) &&
		i.checkAttemptsForCredential(ip, i.ipMap, i.ipMaxAttempt)
}

func (i *InMemoryBruteForceService) checkAttemptsForCredential(
	cred string,
	credMap map[string]*Bucket,
	maxAttempt int,
) bool {
	i.mu.RLock()
	v, ok := credMap[cred]
	i.mu.RUnlock()
	if !ok {
		i.mu.Lock()
		if v, ok := credMap[cred]; ok {
			i.mu.Unlock()
			return i.checkBucketAndIncreaseAttempt(v, cred, credMap, maxAttempt)
		}
		credMap[cred] = createBucket()
		i.createJobForDeleteBucket(cred, credMap)
		i.mu.Unlock()
		return true
	}
	return i.checkBucketAndIncreaseAttempt(v, cred, credMap, maxAttempt)
}

func (i *InMemoryBruteForceService) createJobForDeleteBucket(cred string, credMap map[string]*Bucket) {
	_, _ = i.scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(time.Now().Add(time.Minute)),
		),
		gocron.NewTask(
			func() {
				i.mu.Lock()
				bucket := credMap[cred]
				if time.Since(bucket.firstTime) >= time.Minute {
					delete(credMap, cred)
				}
				defer i.mu.Unlock()
			},
		),
	)
}

func (i *InMemoryBruteForceService) checkBucketAndIncreaseAttempt(
	v *Bucket,
	cred string,
	credMap map[string]*Bucket,
	maxAttempts int,
) bool {
	now := time.Now()
	mu := v.mu
	mu.Lock()
	defer mu.Unlock()
	if now.Sub(v.firstTime) <= v.interval {
		if maxAttempts <= v.attempt {
			log.Debug()
			return false
		}
		v.attempt++
		v.lastTime = now
		return true
	}
	credMap[cred] = createBucket()
	i.createJobForDeleteBucket(cred, credMap)
	return true
}

func NewInMemoryBruteForceService(
	blackService ListService,
	whiteService ListService,
	config config.AttemptsConfig,
) *InMemoryBruteForceService {
	s, _ := gocron.NewScheduler()
	s.Start()
	return &InMemoryBruteForceService{
		scheduler:          s,
		blackListService:   blackService,
		whiteListService:   whiteService,
		loginMaxAttempt:    config.LoginMaxAttempts,
		passwordMaxAttempt: config.PasswordMaxAttempts,
		ipMaxAttempt:       config.IPMaxAttempts,
		mu:                 &sync.RWMutex{},
		loginMap:           make(map[string]*Bucket, 1000),
		passwordMap:        make(map[string]*Bucket, 1000),
		ipMap:              make(map[string]*Bucket, 1000),
	}
}
