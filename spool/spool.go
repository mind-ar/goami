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

//NewSpool retorna un nuevo pool de llamadas salientes
func NewSpool(callLimit int) (*Spool, error) {
	return &Spool{
		//ctx:      ctx,
		mutex: new(sync.Mutex),
		queue: make(chan ami.OriginateData, callLimit-1),
	}, nil
}

//Spool contiene los datos y funciones para manejar pool de llamadas salientes
type Spool struct {
	ctx   context.Context
	mutex *sync.Mutex
	queue chan ami.OriginateData
}

//Originate agrega una nueva llamada a la cola de llamadas del discador
func (s *Spool) Originate(data ami.OriginateData) error {
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
