#include <stdio.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "esp_bt.h"
#include "esp_log.h"
#include "esp_err.h"
#include "esp_bt_main.h"
#include "nvs_flash.h"

static const char *TAG = "BT_MINIMAL";

void app_main(void) {
  esp_err_t ret;

  // Inicializar NVS
  ret = nvs_flash_init();
  if (ret == ESP_ERR_NVS_NO_FREE_PAGES || ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
    ESP_ERROR_CHECK(nvs_flash_erase());
    ret = nvs_flash_init();
  }
  ESP_ERROR_CHECK(ret);

  ESP_LOGI(TAG, "Inicializando Bluetooth...");

  // Inicializar el controlador Bluetooth
  esp_bt_controller_config_t bt_cfg = BT_CONTROLLER_INIT_CONFIG_DEFAULT();
  ret = esp_bt_controller_init(&bt_cfg);
  if (ret != ESP_OK) {
    ESP_LOGE(TAG, "Error al inicializar el controlador Bluetooth: %s", esp_err_to_name(ret));
    return;
  }

  // Habilitar el controlador Bluetooth
  ret = esp_bt_controller_enable(ESP_BT_MODE_CLASSIC_BT);
  if (ret != ESP_OK) {
    ESP_LOGE(TAG, "Error al habilitar el controlador Bluetooth: %s", esp_err_to_name(ret));
    return;
  }

  // Inicializar Bluedroid (pila Bluetooth)
  ret = esp_bluedroid_init();
  if (ret != ESP_OK) {
    ESP_LOGE(TAG, "Error al inicializar Bluedroid: %s", esp_err_to_name(ret));
    return;
  }

  // Habilitar Bluedroid
  ret = esp_bluedroid_enable();
  if (ret != ESP_OK) {
    ESP_LOGE(TAG, "Error al habilitar Bluedroid: %s", esp_err_to_name(ret));
    return;
  }

  ESP_LOGI(TAG, "Bluetooth inicializado correctamente");
}
