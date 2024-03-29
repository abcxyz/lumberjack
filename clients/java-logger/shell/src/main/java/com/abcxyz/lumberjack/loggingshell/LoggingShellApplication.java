/*
 * Copyright 2021 Lumberjack authors (see AUTHORS file)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.abcxyz.lumberjack.loggingshell;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;
import java.io.IOException;
import java.io.OutputStream;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.nio.file.Files;
import java.nio.file.Paths;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/** Entry point for the Logging Shell/Test app. */
@SpringBootApplication
@Slf4j
public class LoggingShellApplication {
  public static void main(String[] args) throws IOException {
    HttpServer jwkServer =
        HttpServer.create(new InetSocketAddress(InetAddress.getLocalHost(), 8080), 0);
    jwkServer.createContext("/.well-known/jwks", new JWKHandler());
    jwkServer.setExecutor(null); // creates a default executor
    jwkServer.start();
    SpringApplication.run(LoggingShellApplication.class, args);
  }

  static class JWKHandler implements HttpHandler {
    @Override
    public void handle(HttpExchange t) throws IOException {
      byte[] publicKey;
      try {
        publicKey = Files.readAllBytes(Paths.get("test_jwks"));
      } catch (Exception e) {
        log.error("Failed to read public key from file.", e);
        t.sendResponseHeaders(500, -1);
        return;
      }
      String response = new String(publicKey);
      t.sendResponseHeaders(200, response.length());
      OutputStream os = t.getResponseBody();
      os.write(response.getBytes());
      os.close();
    }
  }
}
