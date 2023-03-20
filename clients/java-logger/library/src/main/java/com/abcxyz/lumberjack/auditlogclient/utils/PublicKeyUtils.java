package com.abcxyz.lumberjack.auditlogclient.utils;

import java.io.FileReader;
import org.json.simple.*;
import org.json.simple.parser.*;

public class PublicKeyUtils {

  public static String parsePublicKey() {
    JSONParser parser = new JSONParser();
    try {
      Object obj = parser.parse(new FileReader("/etc/lumberjack/public_key.json"));
      JSONObject jsonObject = (JSONObject) obj;
      String decoded = (String) jsonObject.get("decoded");
      return decoded;
    } catch (Exception e) {
      return "";
    }
  }
}
