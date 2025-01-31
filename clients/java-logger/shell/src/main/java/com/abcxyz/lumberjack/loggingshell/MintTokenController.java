package com.abcxyz.lumberjack.loggingshell;

import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.Duration;
import java.time.Instant;
import java.util.Date;
import java.util.UUID;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/mint-token")
public class MintTokenController {
  private static final Logger log = LoggerFactory.getLogger(MintTokenController.class);

  @GetMapping
  public ResponseEntity<String> mintToken() {
    try {
      Path testPrivateKeyPath = Paths.get("integration/testrunner/test_private_key");
      byte[] privateKey = Files.readAllBytes(testPrivateKeyPath);
      Instant now = Instant.now();
      return ResponseEntity.ok(
          Jwts.builder()
              .setAudience("logging-shell")
              .setExpiration(Date.from(now.plus(Duration.ofHours(1))))
              .setId(UUID.randomUUID().toString())
              .setIssuedAt(Date.from(now))
              .setIssuer("lumberjack-test-runner")
              .setNotBefore(Date.from(now))
              .setSubject(
                  "github-automation-bot@gha-lumberjack-ci-i-9d0848.iam.gserviceaccount.com")
              .signWith(Keys.hmacShaKeyFor(privateKey))
              .compact());
    } catch (Exception e) {
      log.error("Failed to read public key from file.", e);
      return ResponseEntity.internalServerError().body("Failed to read public key from file.");
    }
  }
}
