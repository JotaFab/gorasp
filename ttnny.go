package main

import (
	"machine"
	"time"
	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func main() {
	// Inicializa el UART para depuración
	machine.Serial.Configure(machine.SerialConfig{BaudRate: 115200})

	// Inicializa el adaptador Bluetooth
	must("enable BLE stack", adapter.Enable())

	// Inicia el escaneo de dispositivos
	machine.Serial.Write([]byte("Escaneando dispositivos Bluetooth...\n"))
	err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		msg := "Dispositivo encontrado: " + device.Address.String() + " - RSSI: " + string(device.RSSI) + "\n"
		machine.Serial.Write([]byte(msg))
	})
	must("start scan", err)

	// Mantiene el programa en ejecución
	for {
		time.Sleep(time.Second)
	}
}

// must ayuda a manejar errores y depurarlos por serial
func must(action string, err error) {
	if err != nil {
		machine.Serial.Write([]byte("Error en " + action + ": " + err.Error() + "\n"))
		for {
			time.Sleep(time.Second)
		}
	}
}
