# Nebula Graph Console

This repository contains the Nebula Graph Console for Nebula Graph 2.0. Nebula Graph Console (Console for short) is a console for Nebula Graph. With Console, you can create a graph schema, import the demonstration nba dataset, and retrieve data.

## Features

- Supports interactive and non-interactive mode.
- Supports viewing the history statements.
- Supports autocompletion.
- Supports multiple OS and architecture (We recommend Linux/AMD64).

## How to Install

### From Source Code

1. Build Nebula Graph Console
To build Nebula Graph Console, make sure that you have installed [Go](https://golang.org/doc/install).

Run the following command to examine if Go is installed on your machine.

```bash
$ go version
```

Use Git to clone the source code of Nebula Graph Console to your host.

```bash
$ git clone https://github.com/vesoft-inc/nebula-console
```

Run the following command to build Nebula Graph Console.

```bash
$ cd nebula-console
$ make
```
2. Connect to Nebula Graph

To connect to your Nebula Graph services, use the following command.

```bash
$ ./nebula-console -addr <ip> -port <port> -u <username> -p <password>
    [-t 120] [-e "nGQL_statement" | -f filename.nGQL]
```

### Docker

```
$ docker run --rm -ti --network nebula-docker-compose_nebula-net --entrypoint=/bin/sh vesoft/nebula-console:v2-nightly
```

To connect to your Nebula Graph services, run the follow command in the container:

```
docker> nebula-console -u <user> -p <password> --address=graphd --port=3699
```

| Option          | Description                                                                                                                                                                   |
| ------------    | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-h`            | Shows the help menu.                                                                                                                                                          |
| `-addr/-address`| Sets the IP/HOST address of the graphd service. The default address is 127.0.0.1.                                                                                             |
| `-P/-port`      | Sets the port number of the graphd service.                                                                                                                                   |
| `-u/-user`      | Sets the username of your Nebula Graph account.                                                                                                                               |
| `-p/-password`  | Sets the password of your Nebula Graph account.                                                                                                                               |
| `-t/-timeout`   | Sets an integer-type timeout threshold for the connection. The unit is second. The default value is 120.                                                                      |
| `-e/-eval`      | Sets a string-type nGQL statement. The nGQL statement is executed once the connection succeeds. The connection stops after the result is returned.                            |
| `-f/-file`      | Sets the path of an nGQL file. The nGQL statements in the file are executed once the connection succeeds. You'll get the return messages and the connection stops then.       |

Check options for `./nebula-console -h`, try `./nebula-console` in interactive mode directly.
And try `./nebula-console -e 'show hosts'` for the direct script mode.
And try `./nebula-console -f demo.nGQL` for the script file mode.

## Export mode for Nebula Graph Console

When the export mode is enabled, Nebula Graph Console exports all the query results into a CSV file. When the export mode is disabled, the export stops. The syntax is as follows.

> **NOTE**: The following commands are case insensitive.

* Enable nebula-console export mode:

```nGQL
nebula> :set CSV <your_file.csv>
```

* Disable nebula-console export mode:

```nGQL
nebula> :unset CSV
```

## Load nba dataset

To load the demonstration nba dataset, make sure that Console is connected to Nebula Graph.

```ngql
nebula> :play nba
Start loading dataset nba...

Load dataset succeeded!
```

## Export .dot file

To export the graviz text to a `.dot` format, run the following command:

```ngql
nebula> :set dot <filename>
```

For example:

```ngql
nebula> TODO
```

## Disconnect Nebula Graph Console from Nebula Graph

You can use `:EXIT` or `:QUIT` to disconnect from Nebula Graph. For convenience, nebula-console supports using these commands in lower case without the colon (":"), such as `quit`.

```nGQL
nebula> :QUIT

Bye root!

nebula> :EXIT

Bye root!

nebula> quit

Bye root!

nebula> exit

Bye root!
```

## Keyboard Shortcuts

Key Binding                                     | Description
------------------------------------------------|-----------------------------------------------------------
<kbd>Ctrl-A</kbd>, <kbd>Home</kbd>              | Move cursor to beginning of line
<kbd>Ctrl-E</kbd>, <kbd>End</kbd>               | Move cursor to end of line
<kbd>Ctrl-B</kbd>, <kbd>Left</kbd>              | Move cursor one character left
<kbd>Ctrl-F</kbd>, <kbd>Right</kbd>             | Move cursor one character right
<kbd>Ctrl-Left</kbd>, <kbd>Alt-B</kbd>          | Move cursor to previous word
<kbd>Ctrl-Right</kbd>, <kbd>Alt-F</kbd>         | Move cursor to next word
<kbd>Ctrl-D</kbd>, <kbd>Del</kbd>               | (if line is *not* empty) Delete character under cursor
<kbd>Ctrl-D</kbd>                               | (if line *is* empty) End of File --- quit from the console
<kbd>Ctrl-C</kbd>                               | Reset input (create new empty prompt)
<kbd>Ctrl-L</kbd>                               | Clear screen (line is unmodified)
<kbd>Ctrl-T</kbd>                               | Transpose previous character with current character
<kbd>Ctrl-H</kbd>, <kbd>BackSpace</kbd>         | Delete character before cursor
<kbd>Ctrl-W</kbd>, <kbd>Alt-BackSpace</kbd>     | Delete word leading up to cursor
<kbd>Alt-D</kbd>                                | Delete word following cursor
<kbd>Ctrl-K</kbd>                               | Delete from cursor to end of line
<kbd>Ctrl-U</kbd>                               | Delete from start of line to cursor
<kbd>Ctrl-P</kbd>, <kbd>Up</kbd>                | Previous match from history
<kbd>Ctrl-N</kbd>, <kbd>Down</kbd>              | Next match from history
<kbd>Ctrl-R</kbd>                               | Reverse Search history (Ctrl-S forward, Ctrl-G cancel)
<kbd>Ctrl-Y</kbd>                               | Paste from Yank buffer (Alt-Y to paste next yank instead)
<kbd>Tab</kbd>                                  | Next completion
<kbd>Shift-Tab</kbd>                            | (after Tab) Previous completion

## TODO

- CI/CD
- batch process to reduce memory consumption and speed up IO
