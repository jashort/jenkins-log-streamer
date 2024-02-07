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
- `Page Down`/`f`/`space`: Page down
- `Page Up`/`b`: Page up
- `u`/`ctrl+u`: Half page up
- `d`/`ctrl+d`: Half page down
- `up`/`k`: Scroll up
- `down`/`j`: Scroll down
- `g`/`Home`: Go to top
- `G`/`End`: Go to bottom
- `q`/`Escape`/`ctrl+c`: Quit

While at the bottom, the log will automatically scroll for new data. Otherwise, it will stay at the current position.
