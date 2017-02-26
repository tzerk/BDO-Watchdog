# BDO Watchdog

A simple Go program that monitors the BlackDesert64.exe process and its network status. If the process disconnects it sends a custom Telegram message and optionally kills the process or turns off the computer.

## Screenshot 

![](http://i68.tinypic.com/5tuuld.png)

## How it works

1. Check if the designated [process is currently running](https://github.com/mitchellh/go-ps)
2. Obtain the process ID (PID)
3. Run `cmd.exe netstat -aon` and find the PID its output

If the process is running, but its PID is no longer found in the output of `netstat` it then does the following:

- Send a Telegram message using a [URL query string](https://core.telegram.org/bots/api#making-requests)
- (optionally) tries to kill the process
- (optionally) shuts down the computer (`cmd /C shutdown /s`)

## Configuration

On first startup the program creates the `config.yml` file, where you can provide the details of your Telegram bot. It also contains some program specific options.

Option | Description
-------| -----------
token | The token is a string along the lines of `AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw` that is required to authorize the bot
botid | A unique ID of your bot along the lines of `123456789`
chatid | Unique identifier for the target chat or username of the target supergroup or channel
message | The message your bot sends in case of a disconnect
stayalive | By default, the program closes if it has detected a disconnect
process | The process to be monitored, defaults to `BlackDesert64.exe`
timebetweenchecksins | Time in seconds to wait between each polling interval

# Licence
MIT
