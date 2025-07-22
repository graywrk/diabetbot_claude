# Настройка Cloudflare Proxy для DiabetBot

## Проблема "Server is down"

Если Cloudflare показывает "Server is down", проверьте следующие настройки:

## 1. DNS Settings

В панели Cloudflare → **DNS** → **Records**:

```
Type: A
Name: diabetbot (или @)
Content: [IP адрес вашего сервера]
Proxy status: Proxied (оранжевое облако) ✅
TTL: Auto
```

## 2. SSL/TLS Settings

В панели Cloudflare → **SSL/TLS** → **Overview**:

- **Encryption mode: Flexible** 
  (так как ваш сервер работает по HTTP на порту 8080)

## 3. Port Configuration

Убедитесь, что:
- ✅ Ваш сервер слушает на порту **8080**
- ✅ Cloudflare проксирует **HTTPS** → **HTTP:8080**
- ✅ Приложение доступно по адресу `your-server-ip:8080`

## 4. Firewall Rules

Проверьте, что порт 8080 открыт:

```bash
# Ubuntu/Debian
sudo ufw allow 8080

# CentOS/RHEL
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

## 5. Health Check

Проверьте доступность вашего сервера:

```bash
curl -I http://your-server-ip:8080/webapp/
```

Должен вернуть HTTP 200 OK.

## 6. Troubleshooting

### A. Проверьте DNS
```bash
nslookup diabetbot.graywrk.ru
```

### B. Проверьте статус Cloudflare
- Перейдите на https://status.cloudflare.com/
- Убедитесь, что нет проблем с сервисом

### C. Временно отключите proxy
- В DNS настройках нажмите на оранжевое облако
- Сделайте его серым (DNS-only)
- Проверьте прямой доступ к `http://diabetbot.graywrk.ru:8080`

## 7. Page Rules (опционально)

Создайте Page Rule для принудительного HTTPS:

```
URL: diabetbot.graywrk.ru/*
Settings: Always Use HTTPS
```

## Архитектура

```
Пользователь → https://diabetbot.graywrk.ru
     ↓
Cloudflare Proxy (SSL termination)  
     ↓
http://your-server:8080
     ↓
Caddy → приложение:8080
```