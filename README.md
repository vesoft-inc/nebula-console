# Nebula Graph Console

The console for Nebula Graph 2.0

# Build

Install golang by https://golang.org/doc/install, then build with following commands:

```shell
$ cd nebula-console
$ make
```

# Usage

```shell
./nebula-console [-address ip] [-port port] [-u use]r [-p password] [-e "nGQL query statement" |  -f file.nGQL]
```

```shell
-h : help
-address : the Nebula Graph IP address, default value is 127.0.0.1
-port : the Nebula Graph Port, default value is 3699
-u : the Nebula Graph login user name, default value is user
-p : the Nebula Graph login password, default value is password
-e : the nGQL directly
-f : the nGQL script file name
```

Check options for `./nebula-console -h`, try `./nebula-console` in interactive mode directly.
And try `./nebula-console -e 'show hosts'` for the direct script mode.
And try `./nebula-console -f demo.nGQL` for the script file mode.

# Local Command
Nebula-console supports 4 local commands now, which starts with a ':',

## Quit from the console

```shell
(root@nebula) [nba]> :exit
(root@nebula) [nba]> :quit
```

## Output the query results to a csv file
The query results will be output to both the console and the csv file

```shell
(root@nebula) [nba]> :set csv filename.csv
```

## Cancel outputting the query results to the csv file
The query results will only be output to the console

```shell
(root@nebula) [nba]> :unset csv
```

# Feature

- Interactive and non-interactive
- History
- Autocompletion
- Multiple OS and arch supported (linux/amd64 recommend)

# Keyboard Shortcuts

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

# TODO

- CI/CD
- package to RPM/DEB/DOCKER
- batch process to reduce memory consumption and speed up IO
