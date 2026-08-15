One FTP Server
=========
One binary, one command, one user: a complete FTP server with nothing to
configure.


Introduction
---------

	one-ftpserver

That is a complete, running server: the current directory shared over FTP on
port 2121, open to anonymous logins. One FTP Server is configured entirely by
command line arguments — no configuration file is read, and none is written.

Most FTP servers ask for a configuration file before they will start, and a
second one to list their users. One FTP Server has neither:

- **No configuration file.** Every setting is a flag. What you typed is the
  whole state of the server, so `--help` is the entire reference.
- **No installation.** One binary with no runtime to install first: no JVM, no
  Python, no package manager.
- **No key store for FTPS.** `--ssl` is the only thing to do. The certificate is
  generated in memory when the server starts, so there is no `keytool` to run,
  no `.jks` or `.pem` to create, and no key file left behind. A certificate of
  your own goes in as `--cert` and `--key`.
- **No user database.** The single user is `--id` and `--password`. Give an
  `--id` without a password and one is generated for that run and printed.
- **No leftovers.** The server writes nothing but the files clients upload.

Once it is listening, it prints the settings it ended up with and the client
commands that match them, credentials already filled in:

	FTP server started : ftp://192.168.0.10:10021

	# Settings
	- ssl : false
	- port : 10021
	- passivePorts : [none]
	- id : benelog
	- password : 1234
	- home : /srv/files
	- timeout : [none]

	# Client commands
	- upload : curl -T [filename] ftp://192.168.0.10:10021/ -u benelog:1234
	- download : curl -O ftp://192.168.0.10:10021/[filename] -u benelog:1234
	- download : wget --user=benelog --password=1234 ftp://192.168.0.10:10021/[filename]


Usage
---------

### Download

Get the binary for your platform from the
[latest release](https://github.com/benelog/one-ftpserver/releases/latest):

	curl -LO https://github.com/benelog/one-ftpserver/releases/latest/download/one-ftpserver-linux-amd64
	mv one-ftpserver-linux-amd64 one-ftpserver
	chmod +x one-ftpserver

The asset carries the platform in its name; renaming it to `one-ftpserver` is
what makes the commands below work as they are written.

Available assets: `one-ftpserver-linux-amd64`, `one-ftpserver-linux-arm64`,
`one-ftpserver-darwin-amd64`, `one-ftpserver-darwin-arm64`,
`one-ftpserver-windows-amd64.exe`. To pin a version, replace `latest/download`
with `download/v2.1.0`.

### Options

	one-ftpserver --port=10021 --id=benelog --password=1234 --home=/srv/files

| Option | Default | What it does |
| --- | --- | --- |
| `--port` | `2121` | Control port. `0` lets the OS pick one, which is then printed. |
| `--home` | `.` | Directory to serve. A client cannot reach anything above it. |
| `--id` | `anonymous` | The one user allowed to log in. |
| `--password` | | Password of that user. Generated when `--id` is given without one. |
| `--passivePorts` | | Ports for passive transfers, as `10125-10199`. The OS picks when unset. |
| `--ssl` | `false` | Serve over TLS with a certificate generated for this run. |
| `--cert` | | Certificate of your own as a PEM file. Implies `--ssl`. |
| `--key` | | Private key of that certificate, as a PEM file. |
| `--timeout` | `0` | Stop by itself after this long, such as `30m`. `0` keeps it running. |
| `--publicHost` | | Address to show clients. Detected when unset. |
| `--json` | `false` | Print the settings as JSON instead of text. |

### FTPS

	one-ftpserver --ssl

The mode is implicit FTPS, which is what an `ftps://` URL means. The
certificate is generated at startup, in memory: it lasts a year, covers
`localhost` and the addresses of the machine, and is replaced by a new one on
every start. It is signed by nobody, so clients have nothing to verify it
against; that is why the printed `curl` commands carry `-k`, and why FileZilla
or WinSCP ask once whether to trust it:

	curl -O ftps://192.168.0.10:2121/report.pdf -k

The generated certificate encrypts the transfer, but it cannot prove who the
server is. That suits a transfer over a network you do not control, between
two people who already know each other. For a server that strangers have to
trust, bring a certificate of your own, as the usual pair of PEM files.
Giving the pair implies `--ssl`, and the printed commands drop `-k`:

	one-ftpserver --cert=fullchain.pem --key=privkey.pem

The easiest authority to get one from is [Let's Encrypt](https://letsencrypt.org/):
free, automated, and trusted by clients out of the box. On a machine that a
domain points to, with port 80 reachable for the issuance:

	sudo certbot certonly --standalone -d ftp.example.com

	one-ftpserver --cert=/etc/letsencrypt/live/ftp.example.com/fullchain.pem \
	              --key=/etc/letsencrypt/live/ftp.example.com/privkey.pem

Three things to know when serving a certificate of your own:

- Clients verify the name they connect with against it, so hand them the
  domain it was issued for, not the IP address.
- The files under `/etc/letsencrypt` are readable by root only; run the server
  as root, or copy the pair somewhere it may read.
- The files are read once, at startup. A Let's Encrypt certificate lasts 90
  days and `certbot` renews the files on its own, so restart the server after
  a renewal.

A certificate from any other authority works the same, as long as it comes as
PEM: the certificate with its chain in one file, the private key in another.

### Stopping by itself

A server started to move one file rarely needs to outlive it. `--timeout` sets
how long it stays up, after which it stops on its own:

	one-ftpserver --timeout=30m

`Ctrl-C` stops it at any time.

### Behind NAT

Passive transfers tell the client which address to connect back to, and the one
the server sees is not always the one the client can reach. Two flags cover
that case: `--publicHost` for the address to hand out, and `--passivePorts` for
a range narrow enough to open on a firewall or to publish from a container.

	one-ftpserver --publicHost=203.0.113.10 --passivePorts=10125-10199

	docker run -p 2121:2121 -p 10125-10199:10125-10199 ...

### Scripting

`--json` replaces the printed summary with a single object holding the same
information, so the address and the credentials can be picked up without
parsing text:

	one-ftpserver --id=benelog --json

	{
	  "address": "ftp://192.168.0.10:2121",
	  "protocol": "ftp",
	  "host": "192.168.0.10",
	  "port": 2121,
	  "id": "benelog",
	  "password": "UEPH6KBIXDUIVEKC2Q23U3L4QI",
	  "anonymous": false,
	  "home": "/srv/files",
	  "ssl": false,
	  "upload": "curl -T [filename] ftp://192.168.0.10:2121/ -u benelog:UEPH6KBIXDUIVEKC2Q23U3L4QI",
	  "download": "curl -O ftp://192.168.0.10:2121/[filename] -u benelog:UEPH6KBIXDUIVEKC2Q23U3L4QI",
	  "get": "wget --user=benelog --password=UEPH6KBIXDUIVEKC2Q23U3L4QI ftp://192.168.0.10:2121/[filename]",
	  "warnings": [
	    "no --password was given, so this one was generated for this run"
	  ]
	}

Combine it with `--port=0` to run several servers at once without picking ports
for them: each one reports the port it was given.


---------

Changes of each version are listed in [CHANGELOG.md](CHANGELOG.md).
To build or modify the server, see [DEVELOPMENT.md](DEVELOPMENT.md).
