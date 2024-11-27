//go:build linux

package linux

import (
	"context"
	"path/filepath"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"github.com/Southclaws/fault/ftag"
	ac "github.com/bluetuith-org/api-native/api/appfeatures"
	bluetooth "github.com/bluetuith-org/api-native/api/bluetooth"
	"github.com/bluetuith-org/api-native/api/config"
	errorkinds "github.com/bluetuith-org/api-native/api/errorkinds"
	sstore "github.com/bluetuith-org/api-native/api/helpers/sessionstore"
	dbh "github.com/bluetuith-org/api-native/linux/internal/dbushelper"
	mp "github.com/bluetuith-org/api-native/linux/mediaplayer"
	nm "github.com/bluetuith-org/api-native/linux/networkmanager"
	"github.com/bluetuith-org/api-native/linux/obex"
	"github.com/godbus/dbus/v5"
)

// BluezSession describes a Linux Bluez DBus session.
type BluezSession struct {
	systemBus  *dbus.Conn
	sessionBus *dbus.Conn

	netman *nm.NetManager

	store sstore.SessionStore
}

// Start attempts to initialize and start interfacing with the Bluez daemon via DBus.
func (b *BluezSession) Start(
	authHandler bluetooth.SessionAuthorizer,
	cfg config.Configuration,
) (ac.FeatureSet, error) {
	var capabilities ac.Features
	var ce ac.Errors

	if authHandler == nil {
		authHandler = &bluetooth.DefaultAuthorizer{}
	}

	systemBus, err := dbus.SystemBus()
	if err != nil {
		return ac.NilFeatureSet(),
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "start-systembus"),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot initialize system DBus"),
			)
	}

	sessionBus, err := dbus.SessionBus()
	if err != nil {
		return ac.NilFeatureSet(),
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "start-sessionbus"),
				ftag.With(ftag.Internal),
				fmsg.With("Cannot start session DBus"),
			)
	}

	*b = BluezSession{
		systemBus:  systemBus,
		sessionBus: sessionBus,
		store:      sstore.NewSessionStore(),
	}

	if err := b.refreshStore(); err != nil {
		return ac.NilFeatureSet(),
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "refresh-sessionstore"),
				ftag.With(ftag.Internal),
				fmsg.With("Error while initializing object cache"),
			)
	}

	if err := setupAgent(systemBus, authHandler, cfg.AuthTimeout); err != nil {
		return ac.NilFeatureSet(),
			fault.Wrap(err,
				fctx.With(context.Background(), "error_at", "agent-initialize"),
				ftag.With(ftag.Internal),
				fmsg.With("Error while initializing Bluez agent"),
			)
	}

	capabilities.Add(
		ac.FeatureConnection,
		ac.FeaturePairing,
		ac.FeatureMediaPlayer,
	)

	obexcap, cerr := b.obex().Initialize(authHandler, cfg.AuthTimeout)
	if cerr != nil {
		ce.Append(cerr)
	}

	netman, netcap, cerr := nm.Initialize()
	if cerr != nil {
		ce.Append(cerr)
	} else {
		b.netman = netman
	}

	capabilities |= obexcap | netcap

	return ac.NewFeatureSet(capabilities, ce), nil
}

// Stop attempts to stop interfacing with the Bluez daemon.
func (b *BluezSession) Stop() error {
	_ = removeAgent()

	if err := b.sessionBus.Close(); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(), "error_at", "stop-sessionbus"),
			ftag.With(ftag.Internal),
			fmsg.With("Error while closing session bus"),
		)
	}

	if err := b.systemBus.Close(); err != nil {
		return fault.Wrap(err,
			fctx.With(context.Background(), "error_at", "stop-systembus"),
			ftag.With(ftag.Internal),
			fmsg.With("Error while closing system bus"),
		)
	}

	return nil
}

// Adapters returns a list of known adapters.
func (b *BluezSession) Adapters() []bluetooth.AdapterData {
	return b.store.Adapters()
}

// Adapter returns a function call interface to invoke adapter related functions.
func (b *BluezSession) Adapter(adapterAddress bluetooth.MacAddress) bluetooth.Adapter {
	return &adapter{b: b, Address: adapterAddress}
}

// Device returns a function call interface to invoke device related functions.
func (b *BluezSession) Device(deviceAddress bluetooth.MacAddress) bluetooth.Device {
	return &device{b: b, Address: deviceAddress}
}

// Obex returns a function call interface to invoke obex related functions.
func (b *BluezSession) Obex(deviceAddress bluetooth.MacAddress) bluetooth.Obex {
	return &obex.Obex{SessionBus: b.sessionBus, Address: deviceAddress}
}

// Network returns a function call interface to invoke network related functions.
func (b *BluezSession) Network(deviceAddress bluetooth.MacAddress) bluetooth.Network {
	return &nm.Network{NetManager: b.netman, Address: deviceAddress}
}

// MediaPlayer returns a function call interface to invoke mediaplayer related functions.
func (b *BluezSession) MediaPlayer(deviceAddress bluetooth.MacAddress) bluetooth.MediaPlayer {
	return &mp.MediaPlayer{SystemBus: b.systemBus, Address: deviceAddress}
}

// adapter returns an adapter-related function call interface for internal use.
// This is used primarily to initialize adapter objects.
func (b *BluezSession) adapter(path dbus.ObjectPath) *adapter {
	return &adapter{b: b, path: path}
}

// device returns an device-related function call interface for internal use.
// This is used primarily to initialize device objects.
func (b *BluezSession) device(path dbus.ObjectPath) *device {
	return &device{b: b, path: path}
}

// obex returns an obex-related function call interface for internal use.
// This is used primarily to initialize obex objects.
func (b *BluezSession) obex() *obex.Obex {
	return &obex.Obex{SessionBus: b.sessionBus}
}

// mediaPlayer returns an mediaplayer-related function call interface for internal use.
// This is used primarily to initialize mediaPlayer objects.
func (b *BluezSession) mediaPlayer() *mp.MediaPlayer {
	return &mp.MediaPlayer{SystemBus: b.systemBus}
}

// refreshStore refreshes the global session store with adapter and device objects
// that are retrieved from the Bluez DBus interface (system bus).
func (b *BluezSession) refreshStore() error {
	b.store.WaitInitialize()
	go b.watchBluezSystemBus()
	defer b.store.DoneInitialize()

	objects := make(map[dbus.ObjectPath]map[string]map[string]dbus.Variant)
	if err := b.systemBus.Object(dbh.BluezBusName, "/").
		Call(dbh.DbusObjectManagerIface, 0).
		Store(&objects); err != nil {
		return err
	}

	for path, object := range objects {
		for iface, values := range object {
			var err error

			switch iface {
			case dbh.BluezAdapterIface:
				err = b.adapter(path).convertAndStoreObjects(values)

			case dbh.BluezDeviceIface:
				err = b.device(path).convertAndStoreObjects(values)
			}

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// watchBluezSystemBus will register a signal to receive events from the bluez dbus interface.
func (b *BluezSession) watchBluezSystemBus() {
	signalMatch := "type='signal', sender='org.bluez'"
	b.systemBus.BusObject().Call(dbh.DbusSignalAddMatchIface, 0, signalMatch)

	ch := make(chan *dbus.Signal, 1)
	b.systemBus.Signal(ch)

	for signal := range ch {
		b.parseSignalData(signal)
	}
}

// parseSignalData parses bluez DBus signal data.
//
//gocyclo:ignore
func (b *BluezSession) parseSignalData(signal *dbus.Signal) {
	switch signal.Name {
	case dbh.DbusSignalPropertyChangedIface:
		objectInterfaceName, ok := signal.Body[0].(string)
		if !ok {
			return
		}

		propertyMap, ok := signal.Body[1].(map[string]dbus.Variant)
		if !ok {
			return
		}

		switch objectInterfaceName {
		case dbh.BluezAdapterIface:
			dbh.PublishAdapterUpdateEvent(&b.store, signal, propertyMap)

		case dbh.BluezDeviceIface:
			dbh.PublishDeviceUpdateEvent(&b.store, signal, propertyMap)

		case dbh.BluezMediaPlayerIface:
			devicePath := dbus.ObjectPath(filepath.Dir(string(signal.Path)))

			address, ok := dbh.PathConverter.Address(dbh.DbusPathDevice, devicePath)
			if !ok {
				dbh.PublishSignalError(errorkinds.ErrDeviceNotFound, signal,
					"Bluez event handler error",
					"error_at", "pchanged-mediaplayer-address",
				)

				return
			}

			properties, err := b.mediaPlayer().ParseMap(propertyMap)
			if err != nil {
				dbh.PublishSignalError(err, signal,
					"Bluez event handler error",
					"error_at", "pchanged-mediaplayer-address",
				)

				return
			}

			bluetooth.MediaEvent(bluetooth.EventActionUpdated).Publish(bluetooth.MediaEventData{
				Address:   address,
				MediaData: properties,
			})

		case dbh.BluezBatteryIface:
			percentage := -1

			if v, ok := propertyMap["Percentage"]; ok {
				if p, ok := v.Value().(byte); ok {
					percentage = int(p)
				}
			}

			if percentage < 0 {
				dbh.PublishSignalError(errorkinds.ErrEventDataParse, signal,
					"Bluez event handler error",
					"error_at", "pchanged-batterypct-decode",
				)

				return
			}

			dbh.PublishDeviceUpdateEvent(&b.store, signal, propertyMap)
		}

	case dbh.DbusSignalInterfacesAddedIface:
		if len(signal.Body) != 2 {
			return
		}

		objectPath, ok := signal.Body[0].(dbus.ObjectPath)
		if !ok {
			return
		}

		nestedPropertyMap, ok := signal.Body[1].(map[string]map[string]dbus.Variant)
		if !ok {
			return
		}

		for iftype := range nestedPropertyMap {
			mergedPropertyMap, ok := nestedPropertyMap[iftype]
			if !ok {
				return
			}

			for key, values := range nestedPropertyMap {
				if key == iftype {
					continue
				}

				for k, v := range values {
					mergedPropertyMap[k] = v
				}
			}

			switch iftype {
			case dbh.BluezAdapterIface:
				var adapter bluetooth.AdapterData

				if err := dbh.DecodeVariantMap(mergedPropertyMap, &adapter); err != nil {
					dbh.PublishSignalError(err, signal,
						"Bluez event handler error",
						"error_at", "padded-adapter-decode",
					)

					continue
				}

				b.store.AddAdapter(adapter)
				dbh.PathConverter.AddDbusPath(dbh.DbusPathAdapter, objectPath, adapter.Address)

				bluetooth.AdapterEvent(bluetooth.EventActionAdded).
					Publish(adapter.AdapterEventData)

			case dbh.BluezDeviceIface:
				var device bluetooth.DeviceData

				if err := dbh.DecodeVariantMap(mergedPropertyMap, &device); err != nil {
					dbh.PublishSignalError(err, signal,
						"Bluez event handler error",
						"error_at", "padded-device-decode",
					)

					continue
				}

				b.store.AddDevice(device)
				dbh.PathConverter.AddDbusPath(dbh.DbusPathDevice, objectPath, device.Address)

				bluetooth.DeviceEvent(bluetooth.EventActionAdded).
					Publish(device.DeviceEventData)

			case dbh.BluezBatteryIface:
				percentage := -1

				propertyMap := nestedPropertyMap[iftype]

				if v, ok := propertyMap["Percentage"]; ok {
					if p, ok := v.Value().(byte); ok {
						percentage = int(p)
					}
				}

				if percentage < 0 {
					dbh.PublishSignalError(errorkinds.ErrEventDataParse, signal,
						"Bluez event handler error",
						"error_at", "padded-batterypct-decode",
					)

					return
				}

				dbh.PublishDeviceUpdateEvent(&b.store, signal, propertyMap)
			}
		}

	case dbh.DbusSignalInterfacesRemovedIface:
		objectPath, ok := signal.Body[0].(dbus.ObjectPath)
		if !ok {
			return
		}

		ifaceNames, ok := signal.Body[1].([]string)
		if !ok {
			return
		}

		for _, ifaceName := range ifaceNames {
			switch ifaceName {
			case dbh.BluezAdapterIface:
				address, ok := dbh.PathConverter.Address(dbh.DbusPathAdapter, objectPath)
				if !ok {
					dbh.PublishSignalError(errorkinds.ErrAdapterNotFound, signal,
						"Bluez event handler error",
						"error_at", "premoved-adapter-address",
					)

					return
				}

				adapter := bluetooth.AdapterEventData{Address: address}
				bluetooth.AdapterEvent(bluetooth.EventActionRemoved).Publish(adapter)

				b.store.RemoveAdapter(adapter.Address)
				dbh.PathConverter.RemoveDbusPath(dbh.DbusPathAdapter, objectPath)

			case dbh.BluezDeviceIface:
				address, ok := dbh.PathConverter.Address(dbh.DbusPathDevice, objectPath)
				if !ok {
					dbh.PublishSignalError(errorkinds.ErrDeviceNotFound, signal,
						"Bluez event handler error",
						"error_at", "premoved-device-address",
					)

					return
				}

				adapterPath := dbus.ObjectPath(filepath.Dir(string(signal.Path)))

				adapterAddress, ok := dbh.PathConverter.Address(dbh.DbusPathAdapter, adapterPath)
				if !ok {
					dbh.PublishSignalError(errorkinds.ErrAdapterNotFound, signal,
						"Bluez event handler error",
						"error_at", "premoved-device-adapter",
					)

					return
				}

				device := bluetooth.DeviceEventData{
					Address:           address,
					AssociatedAdapter: adapterAddress,
				}
				bluetooth.DeviceEvent(bluetooth.EventActionRemoved).Publish(device)

				b.store.RemoveDevice(device.Address)
				dbh.PathConverter.RemoveDbusPath(dbh.DbusPathDevice, objectPath)
			}
		}
	}
}
