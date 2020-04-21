# Nebula Graph Console

The console for Nebula Graph 2.0

# Build

Install golang by https://golang.org/doc/install, then try `go build`

# Usage

Check options for `./nebula-console -h`, try `./nebula-console` in interactive mode directly.
And try `./nebula-console -e 'exit'` for the direct script mode.
And try `./nebula-console -f demo.nGQL` for the script file mode.

# Feature

- Interactive and non-interactive
- History
- Autocompletion
- Multiple OS and arch supported (linux/amd64 recommend)

# TODO

- CI/CD
- package to RPM/DEB/DOCKER
- batch process to reduce memory consumption and speed up IO
