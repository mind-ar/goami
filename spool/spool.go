package spool

import (
	//"container/list"
	"context"
	//"math/rand"
	"time"
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
		// ready: make(chan bool, 1),
		// maxChannels: callLimit,
		activeChannels: 0,
		GuardTime: 1000,
	}, nil
}

//Spool contiene los datos y funciones para manejar pool de llamadas salientes
type Spool struct {
	ctx   context.Context
	mutex *sync.Mutex
	queue chan ami.OriginateData
	// ready chan bool
	// maxChannels int
	activeChannels int
	OnOriginate func(ami.Response)
	OnHangup func(ami.Response)
	GuardTime time.Duration
}

//Originate agrega una nueva llamada a la cola de llamadas del discador
func (s *Spool) Originate(data ami.OriginateData) error {
	// select {
	// case   <- s.ready:
		s.queue <- data
	// }
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
			func(){
				//originate forzado a asincrónico
				call.Async = "yes"

				/*
				//contador de llamadas activas (uso futuro?)
				s.mutex.Lock()
				s.activeChannels ++
				s.mutex.Unlock()
				defer func() {
					s.mutex.Lock()
					s.activeChannels --
					s.mutex.Unlock()
				}()
				*/
				
				socket, err := pool.GetSocket()
				if err != nil {
					log.Error().Err(err).Msg("Error al Obtener llamada del pool")
					s.OnOriginate(nil)
					return
				}
				defer pool.Close(socket, false)

				uuid, _ := ami.GetUUID()

				call.ChannelID = uuid
				// log.Trace().Interface("call", call).Msg("analizando llamada en queue")
				response, err := ami.Originate(ctx, socket, uuid, call)
				if err != nil {
					log.Error().Err(err).Msg("Error al enviar llamada")
					if s.OnOriginate != nil {
						s.OnOriginate(nil)
					}
					return
				}
				log.Trace().Interface("response", response).Msg("Originate completado")
				resultado := response["Response"][0]
				if resultado != "Success" {
					if s.OnOriginate != nil {
						s.OnOriginate(response)
					}
 					return
				}
				// actionID := response["ActionID"][0]
				// log.Debug().Interface("response", response).Interface("ActionID", actionID).Msg("")

				/*
				Posibles resultados en evento OriginateResponse
				0 = no such extension or number. Also bad dial tech ie. name of a sip trunk that doesn’t exist
				3 = no anwer
				4 = answered
				5 = busy
				8 = congested or not available (Disconnected Number)
				*/
				response, _ = waitForEvent(ctx, socket, uuid, "OriginateResponse")
				if s.OnOriginate != nil {
					s.OnOriginate(response)
				}
				// log.Trace().Interface("response", response).Msg("Resultado Originate")
				resultado = response["Response"][0]
				if resultado != "Success" {
					// log.Error().Interface("Message",response["Message"]).Msg("Error al enviar llamada")
					return
				}
				response, _ = waitForEvent(ctx, socket, uuid, "Hangup")
				if s.OnHangup != nil {
					s.OnHangup(response)
				}

			}()
			time.Sleep(s.GuardTime * time.Millisecond)
		}

		//Simulo una llamada
		//r := rand.Intn(5000)
		//time.Sleep(time.Duration(r+15000) * time.Millisecond)

	}
}



func waitForEvent(ctx context.Context, socket ami.Client, uuid string, event string) (ami.Response, error) {
	// log.Debug().Str("uuid", uuid).Str("event", event).Msg("Esperando Evento")
	for {
		response, _ := ami.Events(ctx, socket)
		if len(response["Uniqueid"]) == 0 || response["Uniqueid"][0] != uuid {
			continue
		}
		evt := response["Event"][0]
		// log.Debug().Interface("response", response).Str("event", event).Msg("")
		if event ==  evt{
			return response, nil
		}
	}
}


// func waitOriginateResponse(ctx context.Context, socket ami.Client, uuid string) (ami.Response, error) {
// 	log.Debug().Str("uuid", uuid).Msg("Esperando Atención de llamada")
// 	for {
// 		events, _ := ami.Events(ctx, socket)
// 		if len(events["Uniqueid"]) == 0 || events["Uniqueid"][0] != uuid {
// 			continue
// 		}
// 		event := events["Event"][0]
// 		log.Debug().Interface("events", events).Str("event", event).Msg("")
// 		if event == "OriginateResponse" {
// 			return events, nil
// 		}
// 	}
// }


// func waitForHangUp(ctx context.Context, socket ami.Client, uuid string) {
// 	log.Debug().Str("uuid", uuid).Msg("Esperando finalizacion de llamada")
// 	for {
// 		events, _ := ami.Events(ctx, socket)
// 		if len(events["Uniqueid"]) == 0 || events["Uniqueid"][0] != uuid {
// 			continue
// 		}
// 		event := events["Event"][0]
// 		log.Debug().Interface("events", events).Str("event", event).Msg("")
// 		// if event == "Hangup" {
// 		// 	return
// 		// }

// 	}
// }
