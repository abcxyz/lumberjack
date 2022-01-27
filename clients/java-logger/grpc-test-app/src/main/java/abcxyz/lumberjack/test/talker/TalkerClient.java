package abcxyz.lumberjack.test.talker;

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
/*
 * Copyright 2015 The gRPC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import com.abcxyz.lumberjack.test.talker.HelloRequest;
import com.abcxyz.lumberjack.test.talker.HelloResponse;
import com.abcxyz.lumberjack.test.talker.TalkerGrpc;
import com.abcxyz.lumberjack.test.talker.WhisperRequest;
import com.abcxyz.lumberjack.test.talker.WhisperResponse;
import com.google.auth.oauth2.AccessToken;
import com.google.auth.oauth2.GoogleCredentials;
import com.google.cloud.bigquery.BigQuery;
import com.google.cloud.bigquery.BigQueryOptions;
import com.google.cloud.bigquery.Job;
import com.google.cloud.bigquery.JobId;
import com.google.cloud.bigquery.JobInfo;
import com.google.cloud.bigquery.QueryJobConfiguration;
import com.google.cloud.bigquery.TableResult;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;
import io.grpc.auth.MoreCallCredentials;
import java.io.IOException;
import java.util.Calendar;
import java.util.UUID;
import java.util.concurrent.TimeUnit;
import java.util.logging.Level;
import java.util.logging.Logger;
import java.util.stream.StreamSupport;

/** A simple client that requests a greeting from the {@link TalkerService} with TLS. */
public class TalkerClient {
  private static final Logger logger = Logger.getLogger(TalkerClient.class.getName());
  private final TalkerGrpc.TalkerBlockingStub blockingStub;

  private static final int INITIAL_AUDIT_LOG_WAIT_SECONDS = 60;
  private static final int RETRY_DELAY_SECONDS = 10;
  private static final int RETRY_COUNT = 5;

  /** Construct client for accessing {@link TalkerService} using the existing channel. */
  public TalkerClient(ManagedChannel channel, GoogleCredentials credentials) throws IOException {
    blockingStub =
        TalkerGrpc.newBlockingStub(channel)
            .withCallCredentials(MoreCallCredentials.from(credentials));
  }

  /** Say hello to server. */
  public void greet(String name, UUID target) {
    logger.info("Will try to greet " + name + " ...");
    HelloRequest request =
        HelloRequest.newBuilder().setMessage(name).setTarget(target.toString()).build();
    HelloResponse response;
    try {
      response = blockingStub.hello(request);
    } catch (StatusRuntimeException e) {
      logger.log(Level.WARNING, "RPC failed: {0}", e.getStatus());
      throw e;
    }
    logger.info("Greeting: " + response.getMessage());
  }

  /** Whisper secrets to server. */
  public void whisper(String secret, UUID target) {
    logger.info("Will try to whisper " + secret + " ...");
    WhisperRequest request =
        WhisperRequest.newBuilder().setMessage(secret).setTarget(target.toString()).build();
    WhisperResponse response;
    try {
      response = blockingStub.whisper(request);
    } catch (StatusRuntimeException e) {
      logger.log(Level.WARNING, "RPC failed: {0}", e.getStatus());
      throw e;
    }
    logger.info("Greeting: " + response.getMessage());
  }

  /**
   * Greet server. If provided, the first element of {@code args} is the name to use in the
   * greeting. First element can either be a list ( in format '["a", "b"]') or a singular host
   */
  public static void main(String[] args) throws Exception {
    // this turns an array string into an array. e.g. "["a", "b"]" -> ["a","b"]
    String hostList = args[0];
    String[] hosts =
        hostList
            .replaceAll("\\[", "")
            .replaceAll("\\]", "")
            .replace("https://", "")
            .replaceAll("\\s", "")
            .replaceAll("\"", "")
            .split(",");

    int port = Integer.parseInt(args[1]);

    GoogleCredentials credentials;
    if (args.length >= 3) {
      // token explicitly added to args, use that.
      logger.info("Using explicit token");
      String token = args[2];
      Calendar currentTime = Calendar.getInstance();
      currentTime.add(Calendar.MINUTE, 15);
      credentials = GoogleCredentials.create(new AccessToken(token, currentTime.getTime()));
    } else {
      logger.info("Attempting to use default credentials");
      // try to use the application default credentials if no token is specified.
      credentials = GoogleCredentials.getApplicationDefault();
    }

    String projectId = System.getenv().get("BACKEND_PROJECT_ID");
    String datasetQuery = System.getenv().get("BIGQUERY_DATASET_QUERY");

    BigQuery bigquery = BigQueryOptions.newBuilder().setProjectId(projectId).build().getService();

    for (String host : hosts) {
      ManagedChannel channel = ManagedChannelBuilder.forAddress(host, port).build();

      try {
        TalkerClient client = new TalkerClient(channel, credentials);
        UUID target = UUID.randomUUID();
        client.greet(host, target);
        QueryJobConfiguration queryConfig = createQuery(projectId, datasetQuery, target.toString());

        UUID target2 = UUID.randomUUID();
        client.whisper("This is a secret! Don't audit log this string", target2);
        QueryJobConfiguration queryConfig2 =
            createQuery(projectId, datasetQuery, target2.toString());

        logger.info("Sleeping and waiting for logs to propogate...");
        TimeUnit.SECONDS.sleep(INITIAL_AUDIT_LOG_WAIT_SECONDS);
        validate(queryConfig, bigquery);
        validate(queryConfig2, bigquery);
      } finally {
        channel.shutdownNow().awaitTermination(5, TimeUnit.SECONDS);
      }
    }
  }

  private static QueryJobConfiguration createQuery(
      String project, String datasetQuery, String target) {
    String query =
        String.format(
            "SELECT * FROM %s.%s WHERE jsonPayload.resource_name=\"%s\" LIMIT 1",
            project, datasetQuery, target);
    return QueryJobConfiguration.newBuilder(query).setUseLegacySql(false).build();
  }

  private static void validate(QueryJobConfiguration queryConfig, BigQuery bigQuery)
      throws InterruptedException {
    JobId jobId = JobId.of(UUID.randomUUID().toString());
    Job queryJob = bigQuery.create(JobInfo.newBuilder(queryConfig).setJobId(jobId).build());

    for (int i = 0; i < RETRY_COUNT; i++) {
      TableResult result = queryJob.getQueryResults();
      if (StreamSupport.stream(result.getValues().spliterator(), false).count() == 1) {
        logger.info("Found a matching audit log!");
        return;
      }
      logger.info("Not found, will retry in " + RETRY_DELAY_SECONDS + " seconds...");
      TimeUnit.SECONDS.sleep(RETRY_DELAY_SECONDS);
    }
    throw new RuntimeException("Unable to find matching audit log in BigQuery.");
  }
}
