A fix release: the error messages sent to a client no longer reveal where on
the server the home directory lives. A failed operation now names the file by
the path the client used, never by its full path on disk.

    curl -LO https://github.com/benelog/one-ftpserver/releases/download/v2.0.1/one-ftpserver-linux-amd64
    mv one-ftpserver-linux-amd64 one-ftpserver
    chmod +x one-ftpserver

Assets: `one-ftpserver-linux-amd64`, `one-ftpserver-linux-arm64`,
`one-ftpserver-darwin-amd64`, `one-ftpserver-darwin-arm64`,
`one-ftpserver-windows-amd64.exe`.
