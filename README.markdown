# Nudger

Nudger is a New Relic metrics importer for StatusPage.

It periodically queries the New Relic REST API (v2), and sends those metrics to StatusPage.

Nudger is different to the built in New Relic integration because it allows you to:

 - Pull metrics from multiple New Relic accounts into a single StatusPage account.
 - Pull metrics from one New Relic account into multiple StatusPage accounts.
 - Pull metrics from multiple New Relic accounts into multiple StatusPage accounts.

## Using

Nudger can pass the following metrics for an application from New Relic to StatusPage:

 - Response time
 - Throughput
 - Error rate

### StatusPage config

On StatusPage, for each of the metrics you want to display (i.e. response time, throughput, error rate) you need to add a new Public Metric with a custom data source.

Once you set up a custom data source with a _Display Name_ and _Display Suffix_, you'll be provided with:

 - An API key
 - A page id
 - A metric id

You'll need these to configure Nudger.

### New Relic config

To get the application id, from the New Relic console, browse to the application, and pick the application id out of the URL.

For example, the application id in https://rpm.newrelic.com/accounts/1246480/applications/12325670 is `12325670`.

To get the API key, from the New Relic console, go to the _Account Settings_, and under _Integrations_ browse to the _API keys_ section.

Reveal the API key, and note it down into the Nudger config.

### Nudger config

In the Nudger config file, you define New Relic applications that should be scraped, and the StatusPage page + metric that should be updated:

```
[
  {
    "nr_api_key": "b1946ac92492d2347c6235b4d2611184",
    "nr_app_id": 12345678,
    "sp_api_key": "a1b271ae-3444-48ac-9060-a1b3c4444",
    "sp_page_id": "trx08hfqyabc",
    "metrics": {
      "response_time": "abcw0cv8wh6l",
      "error_rate": "defd9hl632ch",
      "throughput": "ghizztk3p4t4"
    }
  }
]
```

You can define multiple applications in this config file by giving each app its own configuration section:

```
[
  {
    "nr_api_key": "b1946ac92492d2347c6235b4d2611184",
    "nr_app_id": 12345678,
    "sp_api_key": "a1b271ae-3444-48ac-9060-a1b3c4444",
    "sp_page_id": "trx08hfqyabc",
    "metrics": {
      "response_time": "abcw0cv8wh6l",
      "error_rate": "defd9hl632ch",
      "throughput": "ghizztk3p4t4"
    }
  },
  {
    "nr_api_key": "591785b794601e212b260e25925636fd",
    "nr_app_id": 98765412,
    "sp_api_key": "a1b3c4444-9060-48ac-3444-a1b271ae",
    "sp_page_id": "qwop8hfqy123",
    "metrics": {
      "response_time": "defd9hl632ch",
      "error_rate": "abcw0cv8wh6l",
      "throughput": "jik123hk3pabc"
    }
  }
]
```

You can omit metrics you don't want to push to StatusPage by not including a definition for them. For example:

```
[
  {
    "nr_api_key": "b1946ac92492d2347c6235b4d2611184",
    "nr_app_id": 12345678,
    "sp_api_key": "a1b271ae-3444-48ac-9060-a1b3c4444",
    "sp_page_id": "trx08hfqyabc",
    "metrics": {
      "response_time": "abcw0cv8wh6l",
    }
  }
]
```

### Running Nudger

Start nudger by running:

```
nudger
```

By default it tries to read a config file from `./nudger.json`

You can change the path to read the config from with:

```
nudger --config=/path/to/my/nudger.json
```

You can enable extra debugging messages with:

```
nudger --debug
```

## Developing

``` bash
git clone git@github.com:ausdto/nudger.git
cd nudger
cp nudger.sample.json nudger.test.json
foreman start
```
