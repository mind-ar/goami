package spool

import (
	//"container/list"
	"context"
	//"math/rand"
	//"time"
	//"errors"

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
func (s *Spool) Start(ctx context.Context, pool *ami.Pool) {
	go s.start(ctx, pool)
}

func (s *Spool) start(ctx context.Context, pool *ami.Pool) error {
	for {
		select {
		case call := <-s.queue:
			socket, err := pool.GetSocket()
			if err != nil {
				log.Error().Err(err).Msg("Error al Obtener llamada del pool")
			}

			uuid, _ := ami.GetUUID()
			call.ChannelID = uuid
			log.Trace().Interface("call", call).Msg("analizando llamada en queue")
			response, err := ami.Originate(ctx, socket, uuid, call)
			if err != nil {
				log.Error().Err(err).Msg("Error al enviar llamada")
			}
			actionID := response["ActionID"][0]
			log.Debug().Interface("response", response).Interface("ActionID", actionID).Msg("")
			waitForHangUp(ctx, socket, uuid)
			log.Debug().Msg("Llamada Finalizada")

			pool.Close(socket, false)
		}

		//Simulo una llamada
		//r := rand.Intn(5000)
		//time.Sleep(time.Duration(r+15000) * time.Millisecond)

	}
}

func waitForHangUp(ctx context.Context, socket ami.Client, uuid string) {
	log.Debug().Str("uuid", uuid).Msg("Esperando finalizacion de llamada")
	for {
		events, _ := ami.Events(ctx, socket)
		if len(events["Uniqueid"]) == 0 || events["Uniqueid"][0] != uuid {
			continue
		}
		event := events["Event"][0]
		log.Debug().Interface("events", events).Str("event", event).Msg("")
		if event == "Hangup" {
			return
		}

	}
}
