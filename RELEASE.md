One FTP Server can now serve FTPS with a certificate of your own:

    one-ftpserver --cert=fullchain.pem --key=privkey.pem

- **`--cert` and `--key`** take the usual pair of PEM files, such as the ones
  Let's Encrypt issues. Giving the pair implies `--ssl`.
- A pair that cannot be loaded is refused at startup, before anything listens.
- With a certificate that clients can verify, the printed `curl` commands drop
  `-k`, and the startup summary shows which certificate is being served.

Without the pair, `--ssl` keeps working as before: a certificate generated in
memory for the run, nothing written to disk. The FTPS section of
[README.md](README.md) now covers both, including how to get a certificate
from Let's Encrypt and what to know when serving one.

    curl -LO https://github.com/benelog/one-ftpserver/releases/latest/download/one-ftpserver-linux-amd64
    mv one-ftpserver-linux-amd64 one-ftpserver
    chmod +x one-ftpserver

Assets: `one-ftpserver-linux-amd64`, `one-ftpserver-linux-arm64`,
`one-ftpserver-darwin-amd64`, `one-ftpserver-darwin-arm64`,
`one-ftpserver-windows-amd64.exe`.
