# Настройка YandexGPT для DiabetBot

## Получение API ключа

1. Перейдите в [Yandex Cloud Console](https://console.cloud.yandex.ru/)
2. Создайте или выберите существующий каталог (folder)
3. В разделе "Сервисы" найдите **Foundation Models** (YandexGPT)
4. Активируйте сервис, если он еще не активирован

## Получение Folder ID

1. В Yandex Cloud Console выберите ваш каталог
2. В адресной строке скопируйте folder ID (например: `b1g2abc3def4ghi5jklm`)
3. Или используйте CLI: `yc resource-manager folder list`

## Создание API ключа

1. Перейдите в раздел **Service Accounts** (Сервисные аккаунты)
2. Создайте новый сервисный аккаунт или используйте существующий
3. Назначьте роль `ai.languageModels.user` для сервисного аккаунта
4. Создайте API ключ:
   - Выберите сервисный аккаунт
   - Перейдите в раздел "API-ключи"
   - Нажмите "Создать API-ключ"
   - Скопируйте полученный ключ

## Настройка DiabetBot

Добавьте в файл `.env`:

```bash
# YandexGPT Configuration (primary AI service)
YANDEXGPT_API_KEY=AQVN...your_api_key_here
YANDEXGPT_FOLDER_ID=b1g2abc3def4ghi5jklm

# GigaChat остается как fallback
GIGACHAT_API_KEY=your_gigachat_key_here
GIGACHAT_BASE_URL=https://ngw.devices.sberbank.ru:9443
```

## Приоритет AI сервисов

DiabetBot использует следующий порядок приоритета:

1. **YandexGPT** - если настроен API ключ и Folder ID
2. **GigaChat** - если YandexGPT недоступен, но настроен GigaChat
3. **Fallback** - заглушки с рекомендацией обратиться к врачу

## Проверка настройки

После запуска приложения в логах должна появиться строка:
```
Using YandexGPT as primary AI service
```

Если YandexGPT не настроен, появится:
```
Using GigaChat as primary AI service
```

## Тестирование

1. Запустите бота
2. Отправьте значение глюкозы (например: `7.2`)
3. Бот должен вернуть рекомендацию от YandexGPT

## Лимиты и стоимость

- YandexGPT Lite: бесплатно до 1000 запросов в месяц
- Далее: от 1.2 ₽ за 1000 токенов
- Подробнее: https://cloud.yandex.ru/docs/foundation-models/pricing

## Устранение неполадок

### Ошибка аутентификации
- Проверьте правильность API ключа
- Убедитесь, что сервисному аккаунту назначена роль `ai.languageModels.user`

### Ошибка доступа к модели
- Проверьте правильность Folder ID
- Убедитесь, что Foundation Models активирован в каталоге

### Превышение лимитов
- Проверьте квоты в консоли Yandex Cloud
- При превышении лимитов приложение автоматически переключится на GigaChat