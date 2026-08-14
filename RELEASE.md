One-FTPServer is now a single binary. No JVM, no jar, nothing to install
beside it — and still no configuration file to write.

    curl -LO https://github.com/benelog/one-ftpserver/releases/latest/download/one-ftpserver-linux-amd64
    mv one-ftpserver-linux-amd64 one-ftpserver
    chmod +x one-ftpserver
    ./one-ftpserver

That serves the current directory on port 2121, open to anonymous logins, and
prints the `curl` and `wget` commands to reach it with the credentials already
filled in.

Assets: `one-ftpserver-linux-amd64`, `one-ftpserver-linux-arm64`,
`one-ftpserver-darwin-amd64`, `one-ftpserver-darwin-arm64`,
`one-ftpserver-windows-amd64.exe`.

## Highlights

- **FTPS without a key store.** `--ssl` is all there is to it. Version 0.1
  shipped a `ftpkeystore.jks` inside the jar and copied it into the working
  directory on every start, with its password written in the source. The
  certificate is now generated in memory for the run and nothing is written to
  disk.
- **A password is no longer optional by accident.** `--id` without
  `--password` used to leave the password equal to the user name. It now
  generates one for the run and prints it.
- **`--timeout`.** The server can stop by itself after a given duration, such
  as `--timeout=30m`, which suits a server started to move a single file.
- **`--json`.** The startup summary can be printed as one JSON object, so a
  script can pick up the address and the credentials without parsing text.
- **`--port=0`.** The operating system picks a free port and the summary
  reports it, which allows several servers to run without agreeing on ports.
- **`--publicHost`.** The address handed to clients for passive transfers can
  be set, which is what a server behind NAT or inside a container needs.

## Fixes

- A client can no longer reach anything above the home directory. The served
  file system is now rooted at it.
- Credentials are compared in constant time, so a password cannot be found one
  character at a time by measuring how long the answer takes.
- The printed client commands quote a password holding spaces or shell
  characters, and pass credentials to `wget` as options instead of embedding
  them in the URL.
- A passive port range that cannot be parsed is reported before the server
  starts, instead of being ignored.
- A home directory that does not exist, or that is a file, is refused at
  startup instead of failing on the first client.

## Upgrading from 0.1

`key=value` pairs are replaced by flags. Everything else keeps its name and its
default.

| 0.1 | v2.0.0 |
| --- | --- |
| `java -jar one-ftpserver.jar` | `./one-ftpserver` |
| `port=10021` | `--port=10021` |
| `passivePorts=10125-10199` | `--passivePorts=10125-10199` |
| `id=benelog password=1234` | `--id=benelog --password=1234` |
| `home=/srv/files` | `--home=/srv/files` |
| `ssl=true` | `--ssl` |

Anonymous logins now accept `anonymous` and `ftp` as the user name, with any
password, which is what clients send.

`--help` lists every flag; see [README.md](README.md) for the rest.
