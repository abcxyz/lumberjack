package com.abcxyz.lumberjack.auditlogclient.utils;

public class PublicKeysUtil {
  private static final String PUBLIC_JWK =
      "{"
          + "\"crv\": \"P-256\","
          + "\"kid\": \"integ-key\","
          + "\"kty\": \"EC\","
          + "\"x\": \"hBWj8vw5LkPRWbCr45k0cOarIcWgApM03mSYF911de4\","
          + "\"y\": \"atcBji-0fTfKQu46NsW0votcBrDIs_gFp4YWSEHDUyo\""
          + "}";

  public static final String getPublicJWKString() {
    return PUBLIC_JWK;
  }
}
