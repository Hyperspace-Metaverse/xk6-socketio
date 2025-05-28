// Copyright (c) 2025 Hyperspace Metaverse
// https://hyperspace.mv
// SPDX-License-Identifier: GPL-3.0-or-later

package socketio

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grafana/sobek"
	"github.com/sirupsen/logrus"
	"github.com/zishang520/socket.io-client-go/socket"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/lib"
)

// =====================
// Mocks and Helpers
// =====================

type mockSocket struct {
	emitted         []struct{ event string; data interface{} }
	closed          bool
	timeoutDuration time.Duration
	ackHandler      func(event string, data interface{}, cb func([]interface{}, error))
}

func (m *mockSocket) Emit(event string, data ...interface{}) error {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	m.emitted = append(m.emitted, struct{ event string; data interface{} }{event, d})
	return nil
}

func (m *mockSocket) Timeout(d time.Duration) socketClient {
	m.timeoutDuration = d
	return m
}

func (m *mockSocket) EmitWithAck(event string, data interface{}) func(func([]interface{}, error)) {
	return func(cb func([]interface{}, error)) {
		if m.ackHandler != nil {
			m.ackHandler(event, data, cb)
		} else {
			cb([]interface{}{map[string]interface{}{"success": true}}, nil)
		}
	}
}

func (m *mockSocket) Close() *socket.Socket {
	m.closed = true
	return nil
}

// Patch point for connectFunc
var origConnectFunc = connectFunc

func patchConnectFunc(mock socketClient, err error) func() {
	connectFunc = func(url string, opts socket.OptionsInterface) (socketClient, error) {
		return mock, err
	}
	return func() { connectFunc = origConnectFunc }
}

// Minimal mockVU implementing modules.VU
type mockVU struct {
	logger *logrus.Logger
	state  *lib.State
}

func (vu *mockVU) Context() context.Context { return context.Background() }
func (vu *mockVU) Runtime() *sobek.Runtime { return nil }
func (vu *mockVU) State() *lib.State      { return vu.state }
func (vu *mockVU) Events() common.Events  { return common.Events{} }
func (vu *mockVU) InitEnv() *common.InitEnvironment { return nil }
func (vu *mockVU) RegisterCallback() func(func() error) { return func(fn func() error) {} }

func newMockState(logger *logrus.Logger) *lib.State {
	return &lib.State{Logger: logger}
}

// Adapter for testing: implements socketClient, not *socket.Socket
type fakeSocketClient struct {
	emitted   []struct{ event string; data interface{} }
	timeout   time.Duration
	closed    bool
	ackCb     func(cb func([]interface{}, error))
}
func (f *fakeSocketClient) Emit(event string, data ...interface{}) error {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	f.emitted = append(f.emitted, struct{ event string; data interface{} }{event, d})
	return nil
}
func (f *fakeSocketClient) Timeout(d time.Duration) socketClient { f.timeout = d; return f }
func (f *fakeSocketClient) EmitWithAck(event string, data interface{}) func(func([]interface{}, error)) {
	return func(cb func([]interface{}, error)) {
		if f.ackCb != nil {
			f.ackCb(cb)
		} else {
			cb([]interface{}{map[string]interface{}{"ack": true}}, nil)
		}
	}
}
func (f *fakeSocketClient) Close() *socket.Socket { f.closed = true; return nil }

// =====================
// Connect
// =====================

func TestSocketIOModule_Connect(t *testing.T) {
	mock := &mockSocket{}
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, moduleName: "testmod"}
	unpatch := patchConnectFunc(mock, nil)
	defer unpatch()

	m.Connect("ws://fake")
	if m.client == nil {
		t.Fatal("client should be set after Connect")
	}
}

func TestSocketIOModule_Connect_Error(t *testing.T) {
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, moduleName: "testmod"}
	unpatch := patchConnectFunc(nil, errors.New("fail"))
	defer unpatch()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on connect error")
		}
	}()
	m.Connect("ws://fail")
}

// =====================
// Emit
// =====================

func TestSocketIOModule_Emit(t *testing.T) {
	mock := &mockSocket{}
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, client: mock, moduleName: "testmod"}
	m.Emit("foo", 123)
	if len(mock.emitted) != 1 || mock.emitted[0].event != "foo" || mock.emitted[0].data != 123 {
		t.Errorf("Emit did not call mock correctly: %+v", mock.emitted)
	}
}

func TestSocketIOModule_Emit_NotConnected(t *testing.T) {
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, moduleName: "testmod"}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when emitting without connection")
		}
	}()
	m.Emit("foo", 123)
}

// =====================
// EmitWithAck
// =====================

func TestSocketIOModule_EmitWithAck(t *testing.T) {
	mock := &mockSocket{}
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, client: mock, moduleName: "testmod"}
	mock.ackHandler = func(event string, data interface{}, cb func([]interface{}, error)) {
		cb([]interface{}{map[string]interface{}{"ack": true}}, nil)
	}
	res := m.EmitWithAck("ack", map[string]interface{}{"foo": "bar"})
	if m, ok := res.(map[string]interface{}); !ok || m["ack"] != true {
		t.Errorf("EmitWithAck did not return expected ack: %#v", res)
	}
}

func TestSocketIOModule_EmitWithAck_Error(t *testing.T) {
	mock := &mockSocket{}
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, client: mock, moduleName: "testmod"}
	mock.ackHandler = func(event string, data interface{}, cb func([]interface{}, error)) {
		cb(nil, errors.New("ack error"))
	}
	res := m.EmitWithAck("ack", map[string]interface{}{"foo": "bar"})
	if mm, ok := res.(map[string]interface{}); !ok || mm["success"] != false || mm["error"] == nil {
		t.Errorf("EmitWithAck error branch did not return expected error map: %#v", res)
	}
}

func TestSocketIOModule_EmitWithAck_Timeout(t *testing.T) {
	mock := &mockSocket{
		ackHandler: func(event string, data interface{}, cb func([]interface{}, error)) {
			// do not call cb, simulate timeout
		},
	}
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, client: mock, moduleName: "testmod"}
	res := m.EmitWithAck("ack", map[string]interface{}{"foo": "bar"}, 10) // 10ms timeout
	if m, ok := res.(map[string]interface{}); !ok || m["success"] != false {
		t.Errorf("EmitWithAck did not return timeout error: %#v", res)
	}
}

func TestSocketIOModule_EmitWithAck_NotConnected(t *testing.T) {
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, moduleName: "testmod"}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when EmitWithAck without connection")
		}
	}()
	m.EmitWithAck("ack", nil)
}

func TestSocketIOModule_EmitWithAck_SingleElementSlice(t *testing.T) {
	mock := &mockSocket{}
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, client: mock, moduleName: "testmod"}
	mock.ackHandler = func(event string, data interface{}, cb func([]interface{}, error)) {
		cb([]interface{}{42}, nil)
	}
	res := m.EmitWithAck("ack", map[string]interface{}{"foo": "bar"})
	if res != 42 {
		t.Errorf("EmitWithAck single-element slice should return the element directly, got: %#v", res)
	}
}

func TestSocketIOModule_EmitWithAck_MultiElementSlice(t *testing.T) {
	mock := &mockSocket{}
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, client: mock, moduleName: "testmod"}
	mock.ackHandler = func(event string, data interface{}, cb func([]interface{}, error)) {
		cb([]interface{}{1, 2, 3}, nil)
	}
	res := m.EmitWithAck("ack", map[string]interface{}{"foo": "bar"})
	if arr, ok := res.([]interface{}); !ok || len(arr) != 3 || arr[0] != 1 || arr[1] != 2 || arr[2] != 3 {
		t.Errorf("EmitWithAck multi-element slice should return the slice, got: %#v", res)
	}
}

// =====================
// Disconnect
// =====================

func TestSocketIOModule_Disconnect(t *testing.T) {
	mock := &mockSocket{}
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, client: mock, moduleName: "testmod"}
	m.Disconnect()
	if !mock.closed {
		t.Error("Disconnect did not close the socket")
	}
	if m.client != nil {
		t.Error("Disconnect did not nil the client")
	}
}

// =====================
// Exports
// =====================

func TestSocketIOModule_Exports(t *testing.T) {
	vu := &mockVU{logger: nil, state: newMockState(logrus.New())}
	m := &SocketIOModule{vu: vu, moduleName: "testmod"}
	exp := m.Exports()
	if exp.Named["connect"] == nil || exp.Named["emit"] == nil || exp.Named["disconnect"] == nil || exp.Named["emitWithAck"] == nil {
		t.Error("Exports did not return all expected functions")
	}
}

// =====================
// socketClient Adapter
// =====================

func TestSocketClient_AllMethods(t *testing.T) {
	fake := &fakeSocketClient{}
	fake.Emit("foo", 123)
	if len(fake.emitted) != 1 || fake.emitted[0].event != "foo" || fake.emitted[0].data != 123 {
		t.Errorf("Emit did not call fake correctly: %+v", fake.emitted)
	}
	fake.Timeout(42 * time.Millisecond)
	if fake.timeout != 42*time.Millisecond {
		t.Errorf("Timeout not set correctly: %v", fake.timeout)
	}
	var ackCalled bool
	fake.ackCb = func(cb func([]interface{}, error)) {
		ackCalled = true
		cb([]interface{}{map[string]interface{}{"ack": true}}, nil)
	}
	cb := fake.EmitWithAck("ack", map[string]interface{}{"foo": "bar"})
	var got interface{}
	cb(func(res []interface{}, err error) {
		got = res
	})
	if !ackCalled || got == nil {
		t.Errorf("EmitWithAck did not call ack: %v %v", ackCalled, got)
	}
	fake.Close()
	if !fake.closed {
		t.Error("Close did not set closed flag")
	}
}
