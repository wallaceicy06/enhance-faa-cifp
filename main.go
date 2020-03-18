// enhance_faa_cifp augments FAA CIFP (coded instrument flight procedure) data
// for flight simulators.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/wallaceicy06/enhance-faa-cifp/enhance"
)

var outFile = flag.String("output", "", "path of the file to output augmented procedures")

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

	inReader, err := os.Open(cifpFile)
	if err != nil {
		log.Fatalf("Could not open CIFP file: %v", err)
	}
	defer inReader.Close()
	outWriter := os.Stdout
	if *outFile != "" {
		outWriter, err = os.Create(*outFile)
		if err != nil {
			log.Fatalf("Could not open output file: %v", err)
		}
		defer outWriter.Close()
	}

	if err := enhance.Process(inReader, outWriter); err != nil {
		log.Fatalf("Could not process data: %v", err)
	}
	log.Printf("Processed data.")
}
