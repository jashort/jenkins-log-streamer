# Jenkins Log Streamer
Stream logs to the console from the latest build of a project in [Jenkins](https://jenkins.io)

Features:
- Shows logs from the latest build, even when a new build starts
- Scrolls automatically if the log is at the bottom
- Scroll forward and back through the log in the terminal with arrow keys or page up/page down
- Supports scrolling with the mouse wheel if your terminal does (tested in [iTerm2](https://iterm2.com/))

https://github.com/jashort/jenkins-log-streamer/assets/1596580/c70ca911-76ab-468f-aea6-648898106265

## Installation

Download the appropriate binary for your operating system from https://github.com/jashort/jenkins-log-streamer/releases 
and place it in the path.

## Usage

```shell
jenkins-log-streamer --url https://jenkins.example.com/job/YourProject/ --user YOUR_USERNAME --token YOUR_TOKEN
```

Parameters:

- `--url`: The URL to your job in the Jenkins UI, without a specific build number. For example:
  - A multibranch pipeline in the "Projects" folder, the "demo" project on the "main" branch: `https://jenkins.example.com/job/Projects/job/demo/job/main/`
  - In a regular project in the root: `https://jenkins.example.com/job/YourProject/`
  - Using http on a nonstandard port: `http://jenkins.example.com:8080/job/YourProject/`
- `--user`: The username you use to log in to Jenkins
- `--token`: Your Jenkins API Token. After logging in to Jenkins, click on your username in the upper right corner, 
             then "Configure", then "Add New Token" under "API Token".

`--user` and `--token` may be set in the environment variables `JENKINS_USER` and `JENKINS_TOKEN` instead of setting
them with command line arguments.

```shell
NAME:
   jenkins-log-streamer - Stream console log from a Jenkins project

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
