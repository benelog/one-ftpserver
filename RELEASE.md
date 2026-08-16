One FTP Server tells you what its clients are doing again:

    one-ftpserver --console

    time=2026-08-16T10:27:25 level=INFO msg=login client=1 from=192.168.0.20:51000 id=benelog
    time=2026-08-16T10:27:25 level=INFO msg=download client=1 from=192.168.0.20:51000 id=benelog path=/report.pdf
    time=2026-08-16T10:27:26 level=INFO msg="download done" client=1 from=192.168.0.20:51000 id=benelog path=/report.pdf bytes=182004

- **Every connection, login, transfer, listing, deletion and rename is logged**,
  transfers with the number of bytes that moved. Passwords are never written,
  not even the ones a refused login sent.
- **`--console`** writes those lines to the terminal as they happen, which is
  what [#3](https://github.com/benelog/one-ftpserver/issues/3) asked for. They
  go to the standard error stream, so the startup summary, and the object
  `--json` prints, stay on the standard output stream for a script to read.
- **A log file comes back**, `one-ftpserver.log` by default. The date of the day
  goes into the name, as in `one-ftpserver-2026-08-16.log`, and the file of the
  next day is opened when the date changes. Nothing is renamed and no old file
  is removed, so a day that is written to is never moved and a day that is done
  stays where it is.
- **`--log`** moves that file elsewhere, and `--log=off` keeps none. A path that
  cannot be written is refused at startup, before anything listens.

The file and the console are independent: either one, both, or neither. The
startup summary, and the `--json` object, name the log file of the day.

    curl -LO https://github.com/benelog/one-ftpserver/releases/latest/download/one-ftpserver-linux-amd64
    mv one-ftpserver-linux-amd64 one-ftpserver
    chmod +x one-ftpserver

Assets: `one-ftpserver-linux-amd64`, `one-ftpserver-linux-arm64`,
`one-ftpserver-windows-amd64.exe`, `one-ftpserver-windows-arm64.exe`. On macOS,
build the binary yourself with `go build`, as the Download section of
[README.md](README.md) explains; an unsigned download would be blocked by
Gatekeeper.
