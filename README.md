# Slack Tracer

Slack editing & removing tracer

## Usage

```sh
slack-tracer -f CONFIG_FILE
slack-tracer -t TOKEN [-l HISTORY_LENGTH]
  -f string
        Path of toml file for configure (default "config.tml")
  -t string
        Bearer token
  -l int
        Length of history
```

### Prepare bearer token

[Create app](https://api.slack.com/apps), [add bot user](https://api.slack.com/bot-users) or [generate legacy tokens](https://api.slack.com/custom-integrations/legacy-tokens)

### Put config file (optional)

```toml
token = "xoxp-XXXXX"
length = 500
```

### Run

```sh
./slack-tracer -t xoxp-XXX -l 300 # Option arguments overwrite config from file.
```

## Dependency

- [BurntSushi/toml](https://github.com/BurntSushi/toml)
- [fatih/color](https://github.com/fatih/color)
- [nlopes/slack](https://github.com/nlopes/slack)
- [sergi/go-diff](https://github.com/sergi/go-diff)
