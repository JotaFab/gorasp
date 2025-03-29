#include <stdio.h>
#include "esp_log.h"
#include "nvs_flash.h"
#include "esp_err.h"
#include "esp_bt.h"
#include "esp_bt_main.h"
#include "esp_gap_bt_api.h"
#include <inttypes.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

static const char *TAG = "BT_SCAN";

void bt_gap_callback(esp_bt_gap_cb_event_t event, esp_bt_gap_cb_param_t *param)
{
    switch (event) {
    case ESP_BT_GAP_DISC_RES_EVT:
        ESP_LOGI(TAG, "Dispositivo encontrado: MAC: %02X:%02X:%02X:%02X:%02X:%02X",
                 param->disc_res.bda[0], param->disc_res.bda[1], param->disc_res.bda[2],
                 param->disc_res.bda[3], param->disc_res.bda[4], param->disc_res.bda[5]);
        break;
    case ESP_BT_GAP_DISC_STATE_CHANGED_EVT:
        if (param->disc_st_chg.state == ESP_BT_GAP_DISCOVERY_STARTED) {
            ESP_LOGI(TAG, "Escaneo iniciado.");
        } else if (param->disc_st_chg.state == ESP_BT_GAP_DISCOVERY_STOPPED) {
            ESP_LOGI(TAG, "Escaneo finalizado.");
        }
        break;
    default:
        ESP_LOGW(TAG, "Evento GAP no manejado: %d", event);
        break;
    }
}

void app_main(void)
{
    esp_err_t ret;

    // Inicializar NVS (necesario para la calibración RF y datos persistentes)
    ret = nvs_flash_init();
    if (ret == ESP_ERR_NVS_NO_FREE_PAGES || ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        ESP_ERROR_CHECK(nvs_flash_erase());
        ESP_ERROR_CHECK(nvs_flash_init());
    }
    ESP_LOGI(TAG, "NVS inicializado.");

    ESP_LOGI(TAG, "Inicializando Bluetooth...");

    // Configurar el controlador BT con la configuración por defecto
    esp_bt_controller_config_t bt_cfg = BT_CONTROLLER_INIT_CONFIG_DEFAULT();

    // Si el controlador ya está inicializado, deshabilitarlo y desinicializarlo
    if (esp_bt_controller_get_status() != ESP_BT_CONTROLLER_STATUS_IDLE) {
        ESP_LOGW(TAG, "El controlador Bluetooth ya está inicializado, desinicializando...");
        ESP_ERROR_CHECK(esp_bt_controller_disable());
        ESP_ERROR_CHECK(esp_bt_controller_deinit());
        vTaskDelay(pdMS_TO_TICKS(100));
    }

    // Inicializar y habilitar el controlador BT en modo dual (BTDM)
    ESP_LOGI(TAG, "Inicializando el controlador BT...");
    ESP_ERROR_CHECK(esp_bt_controller_init(&bt_cfg));
    ESP_LOGI(TAG, "Habilitando el controlador BT...");
    ESP_ERROR_CHECK(esp_bt_controller_enable(ESP_BT_MODE_BTDM));
    if (esp_bt_controller_get_status() != ESP_BT_CONTROLLER_STATUS_ENABLED) {
        ESP_LOGE(TAG, "El controlador Bluetooth NO se pudo habilitar");
        return;
    }
    ESP_LOGI(TAG, "Controlador BT habilitado.");

    // Inicializar y habilitar la pila Bluedroid
    ESP_LOGI(TAG, "Inicializando Bluedroid...");
    ESP_ERROR_CHECK(esp_bluedroid_init());
    ESP_ERROR_CHECK(esp_bluedroid_enable());
    ESP_LOGI(TAG, "Bluedroid habilitado correctamente.");

    // Registrar el callback GAP
    ESP_ERROR_CHECK(esp_bt_gap_register_callback(bt_gap_callback));
    ESP_LOGI(TAG, "Callback GAP registrado.");

    // Pequeño retraso para estabilidad antes de iniciar el escaneo
    vTaskDelay(pdMS_TO_TICKS(500));

    ESP_LOGI(TAG, "Iniciando escaneo de dispositivos Bluetooth...");
    ESP_ERROR_CHECK(esp_bt_gap_start_discovery(ESP_BT_INQ_MODE_GENERAL_INQUIRY, 10, 0));

    // Mantener la tarea en ejecución para recibir eventos
    while (1) {
        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}
