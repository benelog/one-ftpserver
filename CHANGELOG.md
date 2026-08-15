# Changelog

## v2.1.0

- `--cert` and `--key` serve FTPS with a certificate of your own, such as one
  from Let's Encrypt, instead of the one generated for the run. Giving the
  pair implies `--ssl`, a pair that cannot be loaded is refused at startup,
  and the printed `curl` commands drop `-k`.

## v2.0.1

- The error messages sent to a client no longer reveal where on the server the
  home directory lives; a failed operation names the file by the path the
  client used.

## v2.0.0

One-FTPServer is rewritten in Go. It is now a single binary: no JVM, no jar,
and nothing to install beside it.

### Highlights

- **Single binary.** Download it, run it. Published for Linux, macOS and
  Windows, on amd64 and arm64.
- **FTPS without a key store.** `--ssl` is now all there is to it. The previous
  version shipped a `ftpkeystore.jks` inside the jar and copied it into the
  working directory on every start, with its password written in the source.
  The certificate is now generated in memory for the run and nothing is written
  to disk.
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

### Fixes

- A client can no longer reach anything above the home directory. The served
  file system is now rooted at it.
- Credentials are compared in constant time, so a password cannot be found one
  character at a time by measuring how long the answer takes.
- The printed client commands quote a password holding spaces or shell
  characters, and pass credentials to `wget` as options instead of embedding
  them in the URL.
- A passive port range that cannot be parsed is now reported before the server
  starts, instead of being ignored.
- A home directory that does not exist, or that is a file, is refused at
  startup instead of failing on the first client.

### Breaking changes

- **Deployment.** `java -jar one-ftpserver.jar` is replaced by a native binary
  per platform, published as a release asset. The Maven build is gone.
- **Arguments.** `key=value` pairs are replaced by flags: `port=10021` becomes
  `--port=10021`, and so on for `passivePorts`, `id`, `password`, `home` and
  `ssl`. `--ssl` takes no value. `--help` lists them all.
- **Anonymous logins** accept `anonymous` and `ftp` as the user name, and any
  password, which is what clients send.

### Development

Quality tooling follows https://blog.benelog.net/go-quality-tools :
`make check` runs goimports, golangci-lint and the tests; `make ci` is what the
GitHub Actions workflow runs on every push and pull request.


## 0.1

The Java version, a wrapper around
[Apache FtpServer](http://mina.apache.org/ftpserver/) packaged as an executable
jar.
