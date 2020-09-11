package cmd

import "pandora_notify_channels/pandora_manager"

var managedDevices map[int64]*pandora_manager.PandoraAccountManager

func Run() error {
	managedDevices := make(map[int64]*pandora_manager.PandoraAccountManager, 0)
}
