# NebulaGraph Console

This repository contains the NebulaGraph Console for NebulaGraph 3.x. NebulaGraph Console (Console for short) is a console for NebulaGraph. With Console, you can create a graph schema, import the demonstration `basketballplayer` dataset, and retrieve data.

## Compatibility Matrix

|                       Console version                                  | NebulaGraph Version |
| :--------------------------------------------------------------------: | :-----------------: |
| **[v2.0.1](https://github.com/vesoft-inc/nebula-console/tree/v2.0.1)** |        2.0.x        |
| **[v2.5.0](https://github.com/vesoft-inc/nebula-console/tree/v2.5.0)** |        2.5.x        |
| **[v2.6.0](https://github.com/vesoft-inc/nebula-console/tree/v2.6.0)** |        2.6.x        |
| **[v3.0.0](https://github.com/vesoft-inc/nebula-console/tree/v3.0.0)** |         3.x         |
| **[v3.1.x](https://github.com/vesoft-inc/nebula-console/tree/v3.1.0)** |         3.x         |
| **[v3.2.x](https://github.com/vesoft-inc/nebula-console/tree/v3.2.0)** |         3.x         |
| **[v3.3.x](https://github.com/vesoft-inc/nebula-console/tree/v3.3.0)** |         3.x         |
| **[v3.4.x](https://github.com/vesoft-inc/nebula-console/tree/v3.4.0)** |         3.x         |
| **[v3.5.x](https://github.com/vesoft-inc/nebula-console/tree/v3.4.0)** |         3.x         |
| **[master](https://github.com/vesoft-inc/nebula-console/tree/master)** |       nightly       |


## Features

- Supports interactive and non-interactive mode.
- Supports viewing the history statements.
- Supports autocompletion.
- Supports multiple OS and architecture (We recommend Linux/AMD64).

## How to Install

### From Source Code

1. Build NebulaGraph Console

    To build NebulaGraph Console, make sure that you have installed [Go](https://golang.org/doc/install).

    > NOTE: Go version provided with apt on ubuntu is usually "outdated".

    Run the following command to examine if Go is installed on your machine.

    ```bash
    $ go version
    ```

    The version should be newer than 1.13.

    Use Git to clone the source code of NebulaGraph Console to your host.

    ```bash
    $ git clone https://github.com/vesoft-inc/nebula-console
    ```

    Run the following command to build NebulaGraph Console.

    ```bash
    $ cd nebula-console
    $ make
    ```
    You can find a binary named `nebula-console`.

2. Connect to NebulaGraph

    To connect to your Nebula Graph services, use the following command.

    ```bash
    $ ./nebula-console -addr <ip> -port <port> -u <username> -p <password>
        [-t 120] [-e "nGQL_statement" | -f filename.nGQL]
    ```

    | Option          | Description         |
    | ------------    | --------------------------------------------------------------------------------------------------------------------------------------------- |
    | `-h`            | Shows the help menu.      |
    | `-addr/-address`| Sets the IP/HOST address of the graphd service.      |
    | `-P/-port`      | Sets the port number of the graphd service.       |
    | `-u/-user`      | Sets the username of your NebulaGraph account. See [authentication](https://docs.nebula-graph.io/2.0/7.data-security/1.authentication/1.authentication/).      |
    | `-p/-password`  | Sets the password of your NebulaGraph account.   |
    | `-t/-timeout`   | Sets an integer-type timeout threshold for the connection. The unit is millisecond. The default value is 120.    |
    | `-e/-eval`      | Sets a string-type nGQL statement. The nGQL statement is executed once the connection succeeds. The connection stops after the result is returned.   |
    | `-f/-file`      | Sets the path of an nGQL file. The nGQL statements in the file are executed once the connection succeeds. You'll get the return messages and the connection stops then.      |
    | `-enable_ssl`   | Enable SSL when connecting to NebulaGraph |
    | `-ssl_root_ca_path` | Sets the path of the certification authority file |
    | `-ssl_cert_path` | Sets the path of the certificate file |
    | `-ssl_private_key_path` | Sets the path of the private key file |
    | `-ssl_insecure_skip_verify` | Controls whether a client verifies the server's certificate chain and host name |


    E.g.,
    ```bash
    $./nebula-console -addr=192.168.10.111 -port 9669 -u root -p nebula
    2021/03/15 15:21:43 [INFO] connection pool is initialized successfully
    Welcome to NebulaGraph!
    ```

    Check options for `./nebula-console -h`:

    - try `./nebula-console` in interactive mode directly.

    - And try `./nebula-console -e 'show hosts'` for the direct script mode.

    - And try `./nebula-console -f demo.nGQL` for the script file mode.

### From Binary

- Download the binaries on the [Releases page](https://github.com/vesoft-inc/nebula-console/releases)

- Add execute permissions to the binary file of NebulaGraph

- Connect to your NebulaGraph services:

```bash
$ ./<$YOUR_BINARY> -addr <ip> -port <port> -u <username> -p <password>
        [-t 120] [-e "nGQL_statement" | -f filename.nGQL]
```

### Docker

Assumed we would like to run console in docker attached to NebulaGraph's docker-compose network, which is by default `nebula-docker-compose_nebula-net` and we would like to use the master console version: `nightly`.

> note: we could replace `nightly` with i.e. `v3.0.0` for specific console version.

Option 0: we could access the container's shell with nebulagraph console installed with:

```bash
$ docker run --rm -ti --network nebula-docker-compose_nebula-net --entrypoint=/bin/sh vesoft/nebula-console:nightly
```

And then call it like:

```bash
docker> nebula-console -u <user> -p <password> --address=<graphd> --port=9669
```

Option 1: or call console directly with:

```bash
docker run --rm -ti --network nebula-net vesoft/nebula-console:nightly -addr graphd -port 9669 -u root -p nebula
```

## Console side commands:

> **NOTE**: The following commands are case insensitive.

* Export the result of the following statement to a csv file:

```nGQL
nebula> :csv a.csv
```

* Export the execution plan in graphviz format to a dot file when profiling a statement with format "dot" or "dot:struct":

```nGQL
nebula> :dot a.dot
nebula> PROFILE FORMAT="dot" GO FROM "player102" OVER serve YIELD dst(edge);
```
You can paste the content in the dot file to `https://dreampuf.github.io/GraphvizOnline/` to show the execution plan.

* Export the execution plan in ASCII Table to a file when profiling a statement :

```nGQL
nebula> :profile profile.log
nebula> PROFILE GO FROM "player102" OVER serve YIELD dst(edge);
nebula> :explain explain.log
nebula> EXPLAIN GO FROM "player102" OVER serve YIELD dst(edge);
```

* Load the demonstration `basketballplayer` dataset:

```ngql
nebula> :play basketballplayer
Start loading dataset basketballplayer...

Load dataset succeeded!
```

* Repeat to execute a statement n times, the average execution time will also be printed:

```ngql
nebula> :repeat 3
```

* Sleep for some seconds, it's just used in `:play basketballplayer`:

```nGQL
nebula> :sleep 3
```

* Exit the console

You can use `:EXIT` or `:QUIT` to disconnect from NebulaGraph. For convenience, nebula-console supports using these commands in lower case without the colon (":"), such as `quit`.

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
