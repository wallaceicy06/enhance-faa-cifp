// enhance_faa_cifp augments FAA CIFP (coded instrument flight procedure) data
// for flight simulators.
package main

import (
	"flag"
	"fmt"
	"log"
)

var outFile = flag.String("output", "FAACIFP_augmented", "path of the file to output augmented procedures")

func init() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: enhance_faa_cifp [options...] <cifp_file>")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatalf("Must specify a CIFP file.")
	}
	cifpFile := flag.Args()[0]
	log.Printf("CIFP file: %q", cifpFile)
	log.Printf("CIFP output file: %q", *outFile)
}
