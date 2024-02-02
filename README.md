# jenkins-log-streamer
Stream Jenkins job logs to the console

## Usage

```shell
NAME:
   jenkins-log-streamer - Stream console log from a Jenkins job

USAGE:
   jenkins-log-streamer [global options] command [command options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --url URL      Jenkins job URL
   --user value   Jenkins user [$JENKINS_USER]
   --token value  Jenkins API token [$JENKINS_TOKEN]
   --help, -h     show help
```

## Keyboard Shortcuts

- `g`/`Home`: Go to top
- `G`/`End`: Go to bottom
- `q`/`Escape`: Quit

While at the bottom, the log will automatically scroll for new data. Otherwise, it will stay at the current position.
