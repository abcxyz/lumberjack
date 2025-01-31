package com.abcxyz.lumberjack.loggingshell;

import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/.well-known/jwks")
public class JwkController {
  private static final Logger log = LoggerFactory.getLogger(JwkController.class);

  @GetMapping
  public ResponseEntity<String> handle() {

    try {
      Path testJwksPath = Paths.get("integration/testrunner/test_jwks");
      byte[] publicKey = Files.readAllBytes(testJwksPath);
      return ResponseEntity.ok(new String(publicKey));
    } catch (Exception e) {
      log.error("Failed to read public key from file.", e);
      return ResponseEntity.internalServerError().body("Failed to read public key from file.");
    }
  }
}
