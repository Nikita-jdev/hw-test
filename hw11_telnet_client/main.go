package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go-telnet [--timeout=duration] host port")
		os.Exit(1)
	}

	address := net.JoinHostPort(args[0], args[1])

	client := NewTelnetClient(address, *timeout, os.Stdin, os.Stdout)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "...Failed to connect to %s: %v\n", address, err)
		os.Exit(1)
	}
	defer func() { _ = client.Close() }()

	fmt.Fprintf(os.Stderr, "Connected to %s\n", address)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT)
	defer stop()

	sendDone := make(chan error, 1)
	recvDone := make(chan error, 1)

	go func() { sendDone <- client.Send() }()
	go func() { recvDone <- client.Receive() }()

	select {
	case <-ctx.Done():
		fmt.Fprintln(os.Stderr, "...interrupted")
	case err := <-sendDone:
		if err != nil {
			fmt.Fprintf(os.Stderr, "...send: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "...EOF")
		}
	case err := <-recvDone:
		if err != nil {
			fmt.Fprintf(os.Stderr, "...receive error: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
		}
	}
}
