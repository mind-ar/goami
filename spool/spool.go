package spool

import (
	//"container/list"
	"context"
	"math/rand"
	//"errors"
	"time"

	log "github.com/rs/zerolog/log"
	"gitlab.com/mind-framework/asterisk/goami/ami"
	"sync"
)

func NewSpool(callLimit int) (*Spool, error) {
	return &Spool{
		//ctx:      ctx,
		mutex: new(sync.Mutex),
		queue: make(chan ami.OriginateData, callLimit-1),
	}, nil
}

type Spool struct {
	ctx   context.Context
	mutex *sync.Mutex
	queue chan ami.OriginateData
}

//NewCall agrega una nueva llamada a la cola de llamadas del discador
func (s *Spool) NewCall(data ami.OriginateData) error {
	s.queue <- data
	return nil
}

//Start inicia el servicio spool
func (s *Spool) Start() {
	go s.start()
}

func (s *Spool) start() error {
	for {
		select {
		case call := <-s.queue:
			log.Trace().Interface("call", call).Msg("analizando llamada en queye")
		}

		//Simulo una llamada
		r := rand.Intn(5000)
		time.Sleep(time.Duration(r+5000) * time.Millisecond)

	}
}

/*
//start ejecuta un bucle infinito procesando las llamadas en cola
func (s *Spool) start() error {
	sleepTime := 2000 * time.Millisecond
	for {
		call, err := s.getNextCall()
		if call == nil || err != nil {
			time.Sleep(sleepTime)
			continue
		}

		//simulo una llamada
		time.Sleep(10 * time.Second)

		log.Trace().Interface("call", call).Msg("analizando llamada en queye")
	}
}

//getNextCall devuelve la llamada en la lista
func (s *Spool) getNextCall() (*ami.OriginateData, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.queue.Len() == 0 {
		return nil, nil
	}
	e := s.queue.Front()

	s.queue.Remove(e)
	call := e.Value.(ami.OriginateData)
	return &call, nil

}
*/
