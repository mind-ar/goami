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
		queue: make(chan spoolCall, callLimit-1),
		// ready: make(chan bool, 1),
		// maxChannels: callLimit,
		activeChannels: 0,
		GuardTime:      1000,
	}, nil
}

//Spool contiene los datos y funciones para manejar pool de llamadas salientes
type Spool struct {
	ctx   context.Context
	mutex *sync.Mutex
	queue chan spoolCall
	// ready chan bool
	// maxChannels int
	activeChannels int
	OnOriginate    func(ami.Response)
	OnHangup       func(ami.Response)
	GuardTime      time.Duration
}

type spoolCall struct {
	data     ami.OriginateData
	callback func(ami.Response)
}

//Originate agrega una nueva llamada a la cola de llamadas del discador
func (s *Spool) Originate(data ami.OriginateData, cb func(ami.Response)) error {

	call := spoolCall{
		data:     data,
		callback: cb,
	}

	s.queue <- call
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
			data := call.data
			cb := call.callback
			response := func() ami.Response {
				//originate forzado a asincrónico
				data.Async = "yes"

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
					return nil
				}
				defer pool.Close(socket, false)

				uuid, _ := ami.GetUUID()

				data.ChannelID = uuid
				// log.Trace().Interface("data", data).Msg("analizando llamada en queue")
				response, err := ami.Originate(ctx, socket, uuid, data)
				if err != nil {
					log.Error().Err(err).Msg("Error al enviar llamada")
					return response
				}
				// log.Trace().Interface("response", response).Msg("Originate completado")
				resultado := response.Get("Response")
				if resultado != "Success" {
					return response
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
				resultado = response.Get("Response")
				if resultado != "Success" {
					return response
				}

				response, _ = waitForEvent(ctx, socket, uuid, "Cdr")
				// if s.OnHangup != nil {
				// 	s.OnHangup(response)
				// }
				return response

			}()
			if cb != nil {
				cb(response)
			}
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
		// log.Debug().Interface("response", response).Interface("Uniqueid", response["Uniqueid"]).Msg("")

		responseUUID := response.Get("Uniqueid")
		if responseUUID == "" {
			responseUUID = response.Get("UniqueID")
		}

		if responseUUID != uuid {
			continue
		}

		evt := response.Get("Event")
		// log.Debug().Interface("response", response).Str("event", event).Str("evt", evt).Msg("")
		if event == evt {
			return response, nil
		}
	}
}
