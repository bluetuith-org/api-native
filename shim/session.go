//go:build !linux

package shim

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	ac "github.com/bluetuith-org/api-native/api/appfeatures"
	"github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/api/config"
	"github.com/bluetuith-org/api-native/api/errorkinds"
	sstore "github.com/bluetuith-org/api-native/api/helpers/sessionstore"
	"github.com/bluetuith-org/api-native/shim/internal/commands"
	"github.com/bluetuith-org/api-native/shim/internal/serde"
	"github.com/puzpuzpuz/xsync/v3"
)

type ShimSession struct {
	features ac.FeatureSet

	conn net.Conn

	listenerErrChan chan error
	listenerEvents  chan []byte
	sessionClosed   atomic.Bool

	cancel context.CancelFunc

	id         *xsync.Counter
	requestMap *xsync.MapOf[int64, chan commands.CommandRawData]

	store sstore.SessionStore

	sync.Mutex
}

const (
	ShimInitErrTimeout  = 1 * time.Second
	ShimCmdReplyTimeout = 5 * time.Second
)

// Start attempts to initialize a session with the system's Bluetooth daemon or service.
// Upon complete initialization, it returns the session descriptor, and capabilities of
// the application.
func (s *ShimSession) Start(authHandler bluetooth.SessionAuthorizer, cfg config.Configuration) (ac.FeatureSet, error) {
	var ce ac.Errors

	var initialized bool
	defer func() {
		if !initialized {
			s.Stop()
		}
	}()

	if authHandler == nil {
		authHandler = bluetooth.DefaultAuthorizer{}
	}

	if cfg.SocketPath == "" {
		t, err := os.CreateTemp("", "shim_sock_")
		t.Close()
		if err != nil {
			return ac.NilFeatureSet(),
				fault.Wrap(err,
					fctx.With(context.Background(), "error_at", "create-socket"),
					ftag.With(ftag.Internal),
					fmsg.With("Cannot create socket file"),
				)
		}

		cfg.SocketPath = t.Name()
	}

	ctx := s.reset(false)

	session := exec.CommandContext(
		ctx, cfg.ShimPath,
		commands.StartRpcServer(cfg.SocketPath).Slice()...,
	)
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	if err := session.Start(); err != nil {
		return ac.NilFeatureSet(),
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "start-shim"),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot start RPC session with shim"),
			)
	}

	if err := s.waitForInitErrors(ctx, session); err != nil {
		return ac.NilFeatureSet(),
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "exec-shim"),
				ftag.With(ftag.Internal),
				fmsg.With("Shim process exited with errors"),
			)
	}

	if err := s.startListener(ctx, cfg.SocketPath); err != nil {
		return ac.NilFeatureSet(),
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "listener-shim"),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot start listener on provided socket"),
			)
	}

	features, _, err := commands.GetFeatureFlags().ExecuteWith(s.executor)
	if err != nil {
		return ac.NilFeatureSet(),
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "shim-features"),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot get advertised features from shim"),
			)
	}

	if err := s.refreshStore(); err != nil {
		return ac.NilFeatureSet(),
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "shim-features"),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot initialize the new session store"),
			)
	}

	initialized = true
	s.features = ac.NewFeatureSet(features, ce)

	for _, absentFeatures := range s.features.Supported.AbsentFeatures() {
		ce.Append(ac.NewError(absentFeatures, errorkinds.ErrNotSupported))
	}

	return s.features, nil
}

// Stop attempts to stop a session with the system's Bluetooth daemon or service.
func (s *ShimSession) Stop() error {
	if s.sessionClosed.Load() {
		return errorkinds.ErrSessionNotExist
	}

	var err error
	if s.conn != nil {
		_, _, err = commands.StopRpcServer().ExecuteWith(s.executor)
	}
	s.reset(true)

	return err
}

// Adapters returns a list of known adapters.
func (s *ShimSession) Adapters() []bluetooth.AdapterData {
	return s.store.Adapters()
}

// Adapter returns a function call interface to invoke adapter related functions.
func (s *ShimSession) Adapter(adapterAddress bluetooth.MacAddress) bluetooth.Adapter {
	return &adapter{s, adapterAddress}
}

// Device returns a function call interface to invoke device related functions.
func (s *ShimSession) Device(deviceAddress bluetooth.MacAddress) bluetooth.Device {
	return &device{s, deviceAddress}
}

// Obex returns a function call interface to invoke obex related functions.
func (s *ShimSession) Obex(deviceAddress bluetooth.MacAddress) bluetooth.Obex {
	return &obex{s, deviceAddress}
}

// Network returns a function call interface to invoke network related functions.
func (s *ShimSession) Network(deviceAddress bluetooth.MacAddress) bluetooth.Network {
	return &network{}
}

// MediaPlayer returns a function call interface to invoke media player/control
// related functions on a device.
func (s *ShimSession) MediaPlayer(deviceAddress bluetooth.MacAddress) bluetooth.MediaPlayer {
	return &mediaPlayer{}
}

func (s *ShimSession) adapter() *adapter {
	return &adapter{}
}

func (s *ShimSession) device() *device {
	return &device{}
}

func (s *ShimSession) refreshStore() error {
	adapters, _, err := commands.GetAdapters().ExecuteWith(s.executor)
	if err != nil {
		return err
	}

	for _, adapter := range adapters {
		newAdapter, err := s.adapter().appendProperties(adapter)
		if err != nil {
			return err
		}
		s.store.AddAdapter(newAdapter)

		devices, _, err := commands.GetPairedDevices(adapter.Address).ExecuteWith(s.executor)
		if err != nil {
			return err
		}
		for _, device := range devices {
			newDevice, err := s.device().appendProperties(device, adapter)
			if err != nil {
				return err
			}

			s.store.AddDevice(newDevice)
		}
	}

	return nil
}

func (s *ShimSession) waitForInitErrors(ctx context.Context, cmd *exec.Cmd) error {
	go func() {
		if err := cmd.Wait(); err != nil && ctx.Err() != nil {
			s.Stop()
		}
	}()

	select {
	case err := <-s.listenerErrChan:
		return err

	case <-ctx.Done():
		return errorkinds.ErrSessionNotExist

	case <-time.NewTimer(ShimInitErrTimeout).C:
	}

	return nil
}

func (s *ShimSession) startListener(ctx context.Context, socketpath string) error {
	socket, err := net.Dial("unix", socketpath)
	if err != nil {
		return err
	}

	s.conn = socket
	go s.listenForEvents(ctx)

	return nil
}

func (s *ShimSession) listenForEvents(ctx context.Context) {
	sendData := func(c chan commands.CommandRawData, m commands.CommandRawData) {
		select {
		case c <- m:
			close(c)
		default:
		}
	}

	for {
		select {
		case <-ctx.Done():
			return

		default:
		}

		if s.sessionClosed.Load() {
			return
		}

		replyHeader := commands.RawCommandHeaderBuffer{}
		headerBytes, err := s.conn.Read(replyHeader[:])
		if err != nil {
			s.handleListenerError(err)
			continue
		}
		if headerBytes != len(replyHeader) {
			continue
		}

		header, err := commands.UnpackReplyHeader(replyHeader)
		if err != nil {
			s.handleListenerError(err)
			continue
		}

		buf := make([]byte, header.ContentSize)
		_, err = io.ReadFull(s.conn, buf)
		if err != nil {
			s.handleListenerError(err)
			continue
		}

		fmt.Println(string(buf))

		if header.EventID > 0 {
			s.handleListenerEvent(buf)
			continue
		}

		replyChan, ok := chan commands.CommandRawData(nil), false
		if header.IsOperationComplete {
			replyChan, ok = s.requestMap.LoadAndDelete(header.RequestId)
		} else {
			replyChan, ok = s.requestMap.Load(header.RequestId)
		}

		if ok {
			sendData(replyChan, commands.CommandRawData{
				CommandMetadata: commands.CommandMetadata{
					OperationId: header.OperationId,
					RequestId:   header.RequestId,
				},
				RawData: buf,
			})
		}
	}
}

func (s *ShimSession) handleListenerEvent(ev []byte) {

}

func (s *ShimSession) handleListenerError(err error) {

}

func (s *ShimSession) executor(params ...string) (chan commands.CommandRawData, error) {
	if s.sessionClosed.Load() {
		return nil, errorkinds.ErrSessionNotExist
	}

	s.id.Inc()
	replyChan := make(chan commands.CommandRawData, 1)
	s.requestMap.Store(s.id.Value(), replyChan)

	command := map[string]any{
		"command":    params,
		"request_id": s.id.Value(),
	}

	commandBytes, err := serde.MarshalJson(command)
	if err != nil {
		return nil, err
	}

	if _, err = s.conn.Write(commandBytes); err != nil {
		return nil, err
	}

	return replyChan, nil
}

func (s *ShimSession) reset(isClosed bool) context.Context {
	s.Lock()
	defer s.Unlock()

	s.features = ac.NilFeatureSet()

	s.sessionClosed.Store(isClosed)
	if isClosed {
		if s.cancel != nil {
			s.cancel()
		}

		if s.conn != nil {
			s.conn.Close()
		}

		return context.Background()
	}

	s.id = xsync.NewCounter()
	s.requestMap = xsync.NewMapOf[int64, chan commands.CommandRawData]()

	s.listenerErrChan = make(chan error, 10)
	s.listenerEvents = make(chan []byte, 1)

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.store = sstore.NewSessionStore()

	return ctx
}
