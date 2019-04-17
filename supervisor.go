package supervisor

import (
	"math/rand"
	"sync"
	"time"

	"github.com/brimstone/logger"
)

var log = logger.New()

type Supervisor struct {
	services []service
}

type Service interface {
	Init() error
	Run() error
}

type status struct {
	name string
	err  error
}

type service struct {
	name    string
	service Service
	backoff time.Duration
	err     error
	retries int
}

func New() *Supervisor {
	rand.Seed(time.Now().UnixNano())
	return &Supervisor{}
}

func (me *Supervisor) Add(name string, s Service) error {
	log.Printf("Adding a %#v\n", s)
	err := s.Init()
	if err != nil {
		return err
	}
	me.services = append(me.services, service{
		name:    name,
		service: s,
		backoff: time.Second,
		retries: 3,
	})
	return nil
}

func (me *Supervisor) Run() error {
	log.Info("Starting services")
	done := make(chan status)
	var wg sync.WaitGroup
	for _, s := range me.services {
		wg.Add(1)
		go me.runOne(&s, done, wg)
	}

	go func() {
		wg.Wait()
		done <- status{}
	}()

	d := <-done
	switch d.name {
	case "":
		log.Info("All services finished successfully")
	default:
		log.Error("Service failed",
			log.Field("service", d.name),
			log.Field("error", d.err),
		)
		return d.err
	}
	return nil
}

func (me *Supervisor) runOne(s *service, failed chan status, wg sync.WaitGroup) {
	for {
		log.Debug("service backoff",
			log.Field("backoff", s.backoff),
			log.Field("service", s.name),
		)
		s.err = s.service.Run()
		// If everything's good, ok to exit
		if s.err == nil {
			wg.Done()
			return
		}
		// If there are no more retires, exit with error
		if s.retries == 0 {
			failed <- status{
				name: s.name,
				err:  s.err,
			}
			wg.Done()
			return
		}
		// Decrement the retry counter and try again
		s.retries--
		jitter := time.Duration(rand.Int63n(int64(s.backoff)))
		s.backoff = s.backoff + jitter
		log.Info("Waiting to retry",
			log.Field("service", s.name),
			log.Field("delay", s.backoff),
			log.Field("retries", s.retries),
		)
		time.Sleep(s.backoff)
		s.backoff = s.backoff * 2
	}
}
