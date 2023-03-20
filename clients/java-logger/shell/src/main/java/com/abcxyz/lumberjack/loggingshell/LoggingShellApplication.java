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

import com.abcxyz.lumberjack.auditlogclient.utils.PublicKeyUtils;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpServer;
import java.io.IOException;
import java.io.OutputStream;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/** Entry point for the Logging Shell/Test app. */
@SpringBootApplication
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
    // Matching private key here:
    // https://github.com/abcxyz/lumberjack/blob/main/integration/testrunner/grpcrunner/grpc.go#L59
    // private static final String PUBLIC_JWK =
    //     "{"
    //         + "\"crv\": \"P-256\","
    //         + "\"kid\": \"integ-key\","
    //         + "\"kty\": \"EC\","
    //         + "\"x\": \"hBWj8vw5LkPRWbCr45k0cOarIcWgApM03mSYF911de4\","
    //         + "\"y\": \"atcBji-0fTfKQu46NsW0votcBrDIs_gFp4YWSEHDUyo\""
    //         + "}";
    private static String PUBLIC_JWK = PublicKeyUtils.parsePublicKey();

    @Override
    public void handle(HttpExchange t) throws IOException {
      String response = String.format("{\"keys\": [%s]}", PUBLIC_JWK);
      t.sendResponseHeaders(200, response.length());
      OutputStream os = t.getResponseBody();
      os.write(response.getBytes());
      os.close();
    }
  }
}
