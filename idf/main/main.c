#include <stdio.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "esp_bt.h"
#include "esp_log.h"
#include "esp_err.h"
#include "esp_bt_main.h"
#include "esp_gap_bt_api.h"
#include "driver/uart.h"
#include <string.h>

#define UART_NUM UART_NUM_1  // UART1 para salida de datos
#define TX_PIN 17
#define RX_PIN 16

static const char *TAG = "BT_SCAN";

void bt_scan_callback(esp_bt_gap_cb_event_t event, esp_bt_gap_cb_param_t *param) {
    if (event == ESP_BT_GAP_DISC_RES_EVT) {
        char bda_str[18];
        for (int i = 0; i < param->disc_res.num_prop; i++) {
            if (param->disc_res.prop[i].type == ESP_BT_GAP_DEV_PROP_BDNAME) {
                snprintf(bda_str, sizeof(bda_str), "%02X:%02X:%02X:%02X:%02X:%02X",
                         param->disc_res.bda[0], param->disc_res.bda[1], param->disc_res.bda[2],
                         param->disc_res.bda[3], param->disc_res.bda[4], param->disc_res.bda[5]);
                
                printf("Dispositivo encontrado: %s - %s\n", bda_str, (char *)param->disc_res.prop[i].val);

                // Enviar los datos por UART
                uart_write_bytes(UART_NUM, bda_str, strlen(bda_str));
                uart_write_bytes(UART_NUM, "\n", 1);
            }
        }
    }
}

void app_main(void) {
    ESP_LOGI(TAG, "Inicializando Bluetooth...");
    
    // Configurar UART
    uart_config_t uart_config = {
        .baud_rate = 115200,
        .data_bits = UART_DATA_8_BITS,
        .parity = UART_PARITY_DISABLE,
        .stop_bits = UART_STOP_BITS_1,
        .flow_ctrl = UART_HW_FLOWCTRL_DISABLE
    };
    uart_param_config(UART_NUM, &uart_config);
    uart_set_pin(UART_NUM, TX_PIN, RX_PIN, UART_PIN_NO_CHANGE, UART_PIN_NO_CHANGE);
    uart_driver_install(UART_NUM, 1024, 0, 0, NULL, 0);

    // Inicializar Bluetooth
    esp_bt_controller_mem_release(ESP_BT_MODE_BLE);
    esp_bt_controller_config_t bt_cfg = BT_CONTROLLER_INIT_CONFIG_DEFAULT();
    esp_bt_controller_init(&bt_cfg);
    esp_bt_controller_enable(ESP_BT_MODE_CLASSIC_BT);
    esp_bluedroid_init();
    esp_bluedroid_enable();

    // Iniciar escaneo de dispositivos Bluetooth
    esp_bt_gap_register_callback(bt_scan_callback);
    esp_bt_gap_start_discovery(ESP_BT_INQ_MODE_GENERAL_INQUIRY, 10, 0);

    ESP_LOGI(TAG, "Escaneando Bluetooth...");
}
