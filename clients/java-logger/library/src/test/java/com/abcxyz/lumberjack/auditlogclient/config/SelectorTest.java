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

package com.abcxyz.lumberjack.auditlogclient.config;

import static org.assertj.core.api.Assertions.assertThat;

import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogType;
import java.util.ArrayList;
import java.util.List;
import org.junit.jupiter.api.Test;

public class SelectorTest {
  @Test
  public void getsLogType() {
    Selector selector = new Selector("", null, LogType.ADMIN_ACTIVITY);
    assertThat(selector.getLogType()).isEqualTo(LogType.ADMIN_ACTIVITY);
  }

  @Test
  public void getsLogType_default() {
    Selector selector = new Selector("", null, null);
    assertThat(selector.getLogType()).isEqualTo(LogType.DATA_ACCESS);
  }

  @Test
  public void getsSelectorlength() {
    Selector selector = new Selector("com.example", null, null);
    assertThat(selector.getLength()).isEqualTo(11);
  }

  @Test
  public void getsDirective() {
    Selector selector = new Selector("", Directive.AUDIT_REQUEST_ONLY, null);
    assertThat(selector.getDirective()).isEqualTo(Directive.AUDIT_REQUEST_ONLY);
  }

  @Test
  public void getsDirective_default() {
    Selector selector = new Selector("", null, null);
    assertThat(selector.getDirective()).isEqualTo(Directive.AUDIT);
  }

  @Test
  public void isApplicable() {
    Selector selector = new Selector("com.example", null, null);
    assertThat(selector.isApplicable("com.example")).isTrue();
    assertThat(selector.isApplicable("com.example.Hello")).isTrue();
    assertThat(selector.isApplicable("com.other.Hello")).isFalse();
  }

  @Test
  public void isApplicable_endWild() {
    Selector selector = new Selector("com.example.*", null, null);
    assertThat(selector.isApplicable("com.example.Hello")).isTrue();
    assertThat(selector.isApplicable("com.example")).isFalse();
    assertThat(selector.isApplicable("com.other.Hello")).isFalse();
  }

  @Test
  public void isApplicable_wild() {
    Selector selector = new Selector("*", null, null);
    assertThat(selector.isApplicable("com.example.Hello")).isTrue();
    assertThat(selector.isApplicable("com.example")).isTrue();
    assertThat(selector.isApplicable("com.other.Hello")).isTrue();
  }

  @Test
  public void getsMostRelevant() {
    List<Selector> selectors = new ArrayList<>();
    Selector selector1 = new Selector("*", null, null);
    selectors.add(selector1);
    Selector selector2 = new Selector("com.example.a", null, null);
    selectors.add(selector2);
    Selector selector3 = new Selector("com.example.a.*", null, null);
    selectors.add(selector3);
    Selector selector4 = new Selector("com.example.a.stuff", null, null);
    selectors.add(selector4);

    assertThat(Selector.returnMostRelevant("", selectors).get()).isEqualTo(selector1);
    assertThat(Selector.returnMostRelevant("other.stuff", selectors).get()).isEqualTo(selector1);
    assertThat(Selector.returnMostRelevant("com.example", selectors).get()).isEqualTo(selector1);
    assertThat(Selector.returnMostRelevant("com.example.a", selectors).get()).isEqualTo(selector2);
    assertThat(Selector.returnMostRelevant("com.example.a.other", selectors).get())
        .isEqualTo(selector3);
    assertThat(Selector.returnMostRelevant("com.example.a.stuff.method", selectors).get())
        .isEqualTo(selector4);
  }

  @Test
  public void getsMostRelevant_none() {
    List<Selector> selectors = new ArrayList<>();
    selectors.add(new Selector("com.example.a", null, null));
    assertThat(Selector.returnMostRelevant("com.other", selectors).isEmpty()).isTrue();
  }
}
