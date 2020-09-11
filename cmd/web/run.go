package web

import (
	"context"
	"github.com/pkg/errors"
	"pandora_notify_channels/cmd/web/db"
	"pandora_notify_channels/pandora_manager"
	"sync"
)

var managedDevices map[int64]*pandora_manager.PandoraAccountManager
var deviceManagerMutex sync.Mutex

func Run() error {
	managedDevices = make(map[int64]*pandora_manager.PandoraAccountManager, 0)

	return nil
}

func acquireDeviceManager(deviceId int64) (*pandora_manager.PandoraAccountManager, error) {
	deviceManagerMutex.Lock()
	defer deviceManagerMutex.Unlock()

	if v, ok := managedDevices[deviceId]; ok {
		return v, nil
	}

	return runDeviceManagerById(deviceId)
}

func runDeviceManagerById(deviceId int64) (*pandora_manager.PandoraAccountManager, error) { // should be called only in lock context
	dbConn, err := db.GetDb()

	if err != nil {
		return nil, errors.WithStack(err)
	}

	dbSes := dbConn.NewSession()
	defer dbSes.Close()

	device := db.PandoraDevice{}
	found, err := dbSes.Where("id = ?", deviceId).Get(&device)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if !found {
		return nil, errors.New("device with such id not found")
	}

	if !device.IsActive {
		return nil, errors.New("device is not active")
	}

	acc := db.PandoraAccount{}

	found, err = dbSes.Where("id = ?", device.PandoraAccountId).Get(&acc)

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, errors.New("account for that device not found")
	}

	return runDeviceManager(device, acc)
}

func runDeviceManager(device db.PandoraDevice, account db.PandoraAccount) (*pandora_manager.PandoraAccountManager, error) {
	deviceManagerMutex.Lock()
	defer deviceManagerMutex.Unlock()

	if v, ok := managedDevices[device.Id]; ok {
		v.Close()
	}

	mgr := pandora_manager.NewPandoraDeviceManager(account.Login, account.Password, device.Id, context.Background())

	if err := mgr.Authorize(); err != nil {
		return nil, err
	}

	managedDevices[device.Id] = mgr

	return mgr, nil
}
