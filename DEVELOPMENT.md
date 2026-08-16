Developing One FTP Server
=========
Go 1.25 or later is the only requirement to build and test.

Build

    make build      # go build -o one-ftpserver .
    make run        # build and serve the current directory on port 2121


Quality tools
---------

Following https://blog.benelog.net/go-quality-tools :

    make fmt        # goimports -w .
    make lint       # golangci-lint run ./...
    make test       # go test ./...
    make check      # fmt + lint + test, before committing
    make ci         # lint + test, what CI runs

`goimports` and `golangci-lint` (v2) are needed for `fmt` and `lint`:

    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

The enabled linters are declared in `.golangci.yml`. The same checks run on
every push and pull request through `.github/workflows/ci.yml`.


Layout
---------

    main.go                                 entry point, every flag is declared here
    internal/oneftpserver/config.go         Config, its defaults and its validation
    internal/oneftpserver/server.go         Server and Run, the loop main.go blocks on
    internal/oneftpserver/driver.go         authentication and the file system given to a client
    internal/oneftpserver/fs.go             file system wrapper that keeps the home path out of errors
    internal/oneftpserver/log.go            the logger, and the log file that follows the day
    internal/oneftpserver/logfs.go          file system wrapper that logs what a client does
    internal/oneftpserver/tls.go            certificate generated for the run, local address lookup
    internal/oneftpserver/usage.go          Summary, printed as text or as JSON

[ftpserverlib](https://github.com/fclairamb/ftpserverlib) speaks the protocol;
what this repository adds is the single user `driver`, the certificate, and the
summary. `driver.AuthUser` returns an `afero.NewBasePathFs` rooted at the home
directory, which is what keeps a client from reaching anything above it. Keep
it that way when changing the file system. It comes back wrapped twice, by
`pathHidingFs` and then by `loggingFs`, so an operation worth logging is added
to `logfs.go` rather than to the library.

`Config.Prepare` is where a bad combination of flags has to be caught. It runs
before anything is listening, so an error there costs the user nothing.


Dependencies
---------
Two of them: `ftpserverlib` for the protocol and `afero` for the file system it
hands to a client. Both end up in the binary; there is nothing to install
beside it.


Tests
---------
`server_test.go` starts a real server on a port the OS picks and talks to it
with the few commands of RFC 959 the tests need, so an upload and a download go
over a real socket. `--port=0` exists partly for this: several tests can run at
once without agreeing on ports.

Testing FTPS goes through the same client over `tls.Dial`, since the server
runs implicit FTPS: the connection is encrypted from its first byte, with no
`AUTH TLS` to negotiate.


Releasing
---------

Build one binary per platform, tag, then publish the assets. macOS is left
out on purpose: an unsigned download is blocked by Gatekeeper, so macOS users
build the binary themselves, as the README explains.

    rm -rf dist && mkdir -p dist
    for target in linux/amd64 linux/arm64 windows/amd64 windows/arm64; do
      GOOS=${target%/*}; GOARCH=${target#*/}
      out=dist/one-ftpserver-$GOOS-$GOARCH
      [ "$GOOS" = windows ] && out=$out.exe
      GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $out .
    done

    git tag -a vX.Y.Z -m "vX.Y.Z" && git push origin vX.Y.Z
    gh release create vX.Y.Z dist/* --title "vX.Y.Z" --notes-file RELEASE.md

Add the section of the new version to `CHANGELOG.md` first, then write
`RELEASE.md` for the release page: it holds the notes of the version being
released only, while `CHANGELOG.md` keeps every version. The download
instructions in `README.md` point at `releases/latest`, so they need no update;
only the pinned example does.
