package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"unicode"
)

func AutomateAPScan(sh *SerialHandler, messageBuffer *[]string, bufferMutex *sync.Mutex) ([]string, error) {
	var ssids []string
	seen := make(map[string]bool)

	// Send "scanap" command
	err := sh.SendCommand("scanap")
	if err != nil {
		return nil, fmt.Errorf("error sending scanap command: %w", err)
	}

	// Wait for 10 seconds
	time.Sleep(10 * time.Second)

	// Send "stopscan" command
	err = sh.SendCommand("stopscan")
	if err != nil {
		return nil, fmt.Errorf("error sending stopscan command: %w", err)
	}

	// Send "list -a" command
	err = sh.SendCommand("list -a")
	if err != nil {
		return nil, fmt.Errorf("error sending list -a command: %w", err)
	}

	// Wait for a short time to allow messages to be received
	time.Sleep(3 * time.Second)

	// Read messages from the buffer
	bufferMutex.Lock()
	messages := *messageBuffer
	*messageBuffer = []string{} // Clear the buffer
	bufferMutex.Unlock()

	// Parse the AP list from the messages
	for _, msg := range messages {
		// Ignore messages related to "stopscan"
		log.Println(msg)
		if strings.Contains(msg, "Stopping Wi") || strings.Contains(msg, "#stopscan") {
			continue
		}

		if strings.Contains(msg, "[") && strings.Contains(msg, "]") {
			// Example message: ESP32: [0][CH:11] 62:bd:2c:cd:52:39
			parts := strings.Split(msg, "]")
			if len(parts) < 2 {
				continue
			}

			dataPart := parts[1]

			dataPart = strings.TrimSpace(dataPart)

			cleanedSSID := cleanSSID(dataPart)

			if !seen[cleanedSSID] {
				ssids = append(ssids, cleanedSSID)
				seen[cleanedSSID] = true
			}
		}
	}
	log.Println(ssids)

	return ssids, nil
}

func cleanSSID(ssid string) string {
	cleanedSSID := ""
	for _, r := range ssid {
		if unicode.IsPrint(r) {
			cleanedSSID += string(r)
		}
	}
	return strings.TrimSpace(cleanedSSID)
}
