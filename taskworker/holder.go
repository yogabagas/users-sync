package taskworker

import (
	"errors"
	"sync"
)

type holder struct {
	mx     sync.Mutex
	active uint8
	res    []Result
}

func (h *holder) Add() {
	h.mx.Lock()
	h.active++
	h.mx.Unlock()
}
func (h *holder) GetActiveWorker() uint8 {
	h.mx.Lock()
	defer h.mx.Unlock()
	return h.active
}
func (h *holder) Store(res Result) {
	h.mx.Lock()
	h.active--
	h.res = append(h.res, res)
	h.mx.Unlock()
}

func (h *holder) GetAllResult() []Result {
	h.mx.Lock()
	defer h.mx.Unlock()

	return h.res
}

type Result struct {
	Result interface{}
	Err    error
}

var ErrorInvalidObject = errors.New("InvalidObject")
