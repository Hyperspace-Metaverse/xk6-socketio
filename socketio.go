// Copyright (c) 2025 Hyperspace Metaverse
// https://hyperspace.mv
// SPDX-License-Identifier: GPL-3.0-or-later

package socketio

import (
	"sync"
	"time"

	"github.com/zishang520/socket.io-client-go/socket"
	"go.k6.io/k6/js/modules"
)

// socketClient abstracts the socket.io client for tests.
type socketClient interface {
	Emit(event string, data ...interface{}) error
	Timeout(d time.Duration) socketClient
	EmitWithAck(event string, data interface{}) func(func([]interface{}, error))
	Close() *socket.Socket
}

// socketClientAdapter adapts a real *socket.Socket to the socketClient interface.
type socketClientAdapter struct {
	s *socket.Socket
}

func (a *socketClientAdapter) Emit(event string, data ...interface{}) error {
	return a.s.Emit(event, data...)
}

func (a *socketClientAdapter) Timeout(d time.Duration) socketClient {
	return &socketClientAdapter{s: a.s.Timeout(d)}
}

func (a *socketClientAdapter) EmitWithAck(event string, data interface{}) func(func([]interface{}, error)) {
	return func(cb func([]interface{}, error)) {
		a.s.EmitWithAck(event, data)(func(res []interface{}, err error) {
			cb(res, err)
		})
	}
}

func (a *socketClientAdapter) Close() *socket.Socket {
	return a.s.Close()
}

// connectFuncType allows dependency injection for socket connection (for testing).
type connectFuncType func(url string, opts socket.OptionsInterface) (socketClient, error)

var connectFunc connectFuncType = func(url string, opts socket.OptionsInterface) (socketClient, error) {
	c, err := socket.Connect(url, opts)
	if err != nil {
		return nil, err
	}
	return &socketClientAdapter{s: c}, nil
}

// RootModule implements the k6 module interface.
type RootModule struct {
	moduleName string
}

// SocketIOModule is the per-VU instance for the extension.
type SocketIOModule struct {
	vu         modules.VU
	client     socketClient
	mu         sync.Mutex
	moduleName string
}

// NewModuleInstance creates a new per-VU module instance.
func (r *RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &SocketIOModule{vu: vu, moduleName: r.moduleName}
}

// Connect establishes a new socket.io connection.
func (m *SocketIOModule) Connect(url string) {
	client, err := connectFunc(url, nil)
	if err != nil {
		panic(m.vu.Runtime().NewGoError(err))
	}
	m.mu.Lock()
	m.client = client
	m.mu.Unlock()
}

// Emit sends an event to the server.
func (m *SocketIOModule) Emit(event string, data interface{}) {
	m.vu.State().Logger.Debugf("[%s] Emit: event=%s, data=%#v", m.moduleName, event, data)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client == nil {
		panic("Socket.IO client not connected")
	}
	m.client.Emit(event, data)
}

// EmitWithAck emits an event and waits for an acknowledgement or timeout.
func (m *SocketIOModule) EmitWithAck(event string, data interface{}, args ...interface{}) interface{} {
	const defaultTimeout = 2000
	var timeout int
	if len(args) > 0 {
		if t, ok := args[0].(int); ok {
			timeout = t
		} else {
			timeout = defaultTimeout
		}
	} else {
		timeout = defaultTimeout
	}
	m.vu.State().Logger.Debugf("[%s] EmitWithAck: event=%s, data=%#v, timeout=%d", m.moduleName, event, data, timeout)
	m.mu.Lock()
	if m.client == nil {
		m.mu.Unlock()
		panic("Socket.IO client not connected")
	}
	client := m.client.Timeout(time.Duration(timeout) * time.Millisecond)
	m.mu.Unlock()
	ch := make(chan []interface{}, 1)
	client.EmitWithAck(event, data)(func(res []interface{}, err error) {
		if err != nil {
			m.vu.State().Logger.Debugf("[%s] Ack callback for %s: error: %v", m.moduleName, event, err)
			ch <- []interface{}{map[string]interface{}{"success": false, "error": err.Error()}}
			return
		}
		m.vu.State().Logger.Debugf("[%s] Ack callback for %s: %#v", m.moduleName, event, res)
		ch <- res
	})
	select {
	case res := <-ch:
		if len(res) == 1 {
			return res[0]
		}
		return res
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return map[string]interface{}{ "success": false, "error": "ack timeout" }
	}
}

// Disconnect closes the socket connection.
func (m *SocketIOModule) Disconnect() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client != nil {
		m.client.Close()
		m.client = nil
	}
}

// Exports returns the JS-exposed API for the module.
func (m *SocketIOModule) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"connect":     m.Connect,
			"emit":        m.Emit,
			"disconnect":  m.Disconnect,
			"emitWithAck": m.EmitWithAck,
		},
	}
}
