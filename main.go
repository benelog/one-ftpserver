// Command one-ftpserver runs an FTP server that is configured entirely by
// command line arguments.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/benelog/one-ftpserver/internal/oneftpserver"
)

func main() {
	config := &oneftpserver.Config{}

	flag.IntVar(&config.Port, "port", oneftpserver.DefaultPort, "control port to listen on, 0 to let the OS pick one")
	flag.StringVar(&config.Home, "home", ".", "directory to serve")
	flag.StringVar(&config.ID, "id", oneftpserver.AnonymousID, "user allowed to log in, anonymous by default")
	flag.StringVar(&config.Password, "password", "", "password of that user, generated when --id is given without one")
	flag.StringVar(&config.PassivePorts, "passivePorts", "", "ports for passive transfers, as a range such as 10125-10199")
	flag.BoolVar(&config.SSL, "ssl", false, "serve over TLS with a certificate generated for this run")
	flag.StringVar(&config.Cert, "cert", "", "certificate of your own as a PEM file, implies --ssl")
	flag.StringVar(&config.Key, "key", "", "private key of that certificate as a PEM file")
	flag.DurationVar(&config.Timeout, "timeout", 0, "stop by itself after this long, such as 30m, 0 to keep running")
	flag.StringVar(&config.PublicHost, "publicHost", "", "address to show clients, useful behind NAT, detected by default")
	flag.BoolVar(&config.JSON, "json", false, "print the settings as JSON instead of text")
	flag.StringVar(&config.Log, "log", oneftpserver.DefaultLogFile,
		"file to log the activity to, one per day, "+oneftpserver.LogOff+" to keep no log file")
	flag.BoolVar(&config.Console, "console", false, "log the activity to the console as well")

	flag.Usage = usage
	flag.Parse()

	if err := config.Prepare(); err != nil {
		fmt.Fprintf(os.Stderr, "one-ftpserver: %v\n", err)
		os.Exit(2)
	}

	if err := oneftpserver.Run(config, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "one-ftpserver: %v\n", err)
		os.Exit(1)
	}
}

const (
	header = `one-ftpserver serves a directory over FTP. No configuration file is read or written.

Usage:
  one-ftpserver [options]

Options:
`
	examples = `
Examples:
  one-ftpserver
  one-ftpserver --port=10021 --id=benelog --password=1234
  one-ftpserver --port=10021 --passivePorts=10125-10199 --ssl --home=/srv/files --timeout=30m
  one-ftpserver --cert=fullchain.pem --key=privkey.pem
  one-ftpserver --console --log=off
`
)

func usage() {
	out := flag.CommandLine.Output()

	_, _ = fmt.Fprint(out, header)
	flag.PrintDefaults()
	_, _ = fmt.Fprint(out, examples)
}
