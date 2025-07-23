#!/usr/bin/env python3
"""
GigaChat API Proxy
Прокси-сервер для обхода проблем с Docker контейнером
"""

import os
import json
import subprocess
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import parse_qs
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class GigaChatProxyHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path == '/oauth':
            try:
                # Читаем тело запроса
                content_length = int(self.headers.get('Content-Length', 0))
                post_data = self.rfile.read(content_length).decode('utf-8')
                
                # Получаем API ключ из заголовка
                auth_header = self.headers.get('Authorization', '')
                if not auth_header.startswith('Basic '):
                    self.send_error(400, "Missing Authorization header")
                    return
                
                api_key = auth_header[6:]  # Убираем "Basic "
                
                logger.info(f"Proxying OAuth request with data: {post_data}")
                
                # Выполняем curl запрос
                curl_cmd = [
                    'curl', '-k', '-L', '-X', 'POST',
                    'https://ngw.devices.sberbank.ru:9443/api/v2/oauth',
                    '-H', 'Content-Type: application/x-www-form-urlencoded',
                    '-H', 'Accept: application/json',
                    '-H', f'Authorization: Basic {api_key}',
                    '-d', post_data
                ]
                
                result = subprocess.run(curl_cmd, capture_output=True, text=True)
                
                logger.info(f"Curl response code: {result.returncode}")
                logger.info(f"Curl stdout: {result.stdout}")
                logger.info(f"Curl stderr: {result.stderr}")
                
                if result.returncode == 0:
                    # Успешный ответ
                    self.send_response(200)
                    self.send_header('Content-Type', 'application/json')
                    self.end_headers()
                    self.wfile.write(result.stdout.encode('utf-8'))
                else:
                    # Ошибка
                    self.send_response(500)
                    self.send_header('Content-Type', 'text/plain')
                    self.end_headers()
                    self.wfile.write(f"Curl error: {result.stderr}".encode('utf-8'))
                    
            except Exception as e:
                logger.error(f"Error processing request: {e}")
                self.send_error(500, str(e))
        else:
            self.send_error(404)

if __name__ == '__main__':
    port = 8888
    server = HTTPServer(('0.0.0.0', port), GigaChatProxyHandler)
    logger.info(f"Starting GigaChat proxy on 0.0.0.0:{port}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        logger.info("Proxy server stopped")
        server.server_close()