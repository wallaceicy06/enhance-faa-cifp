# Enhance FAA Coded Instrument Flight Procedures

This program augments a provided coded instrument flight procedures (CIFP)
database with more accurate [localizer](https://en.wikipedia.org/wiki/Instrument_landing_system_localizer) bearings for flight simulators.

![](https://github.com/wallaceicy06/enhance-faa-cifp/workflows/Go/badge.svg)

## Background

The Federal Aviation Administration (FAA) CIFP file is released every 28 days
and is the basis for all digital navigation products like GPS databases,
ForeFlight, etc. It can also be used in flight simulators like
[X-Plane](https://www.x-plane.com/) in order to provide the most up-to-date
navaids, waypoints, and procedures.

The [ARINC 424](https://en.wikipedia.org/wiki/ARINC_424) specification defines
fields for localizers to provide an accurate bearing, but unfortunately this
field is not included in the FAA's public CIFP release. This results in
simulators displaying localizer courses that are significantly different from
the actual path (hundreds of feet off).

It turns out that a reasonable estimate can be made for the localizer bearing by
computing the bearing between the [final approach
fix](https://en.wikipedia.org/wiki/Final_approach_(aeronautics)#Final_approach_point)
(FAF) and the localizer unit itself. This data is present in the FAA CIFP file,
and this program makes use of that data to augment every localizer with a more
accurate bearing.

## Install

To install this package, you need to install
[Go](https://golang.org/doc/install) on your system. The following commands
are for Unix-like systems. I am very sorry if you use Windows.

1. Fetch the repository.

   ```shell
   go get github.com/wallaceicy06/enhance-faa-cifp
   ```

1. Build the code from source. This will install the program under `$GOPATH/bin`.

   ```shell
   go install github.com/wallaceicy06/enhance-faa-cifp
   ```

1. Make sure that `$GOPATH/bin` is in your `$PATH`.

    ```shell
    export PATH="${PATH}:${GOPATH}/bin"
    ```

## Usage

To use this program, you need to have a CIFP file in the ARINC 424 format. The
FAA publishes the CIFP for download on their
[website](https://www.faa.gov/air_traffic/flight_info/aeronav/digital_products/cifp/download/).

1. To process your CIFP file and output the results to a new file, run:

   ```shell
   enhance-faa-cifp --output=/path/to/FAACIFP_enhanced /path/to/FAACIFP18
   ```

2. If you would like to use this data in X-Plane, locate your X-Plane install
directory, and place the file in the "Custom Data" folder with the name
`earth_424.dat`. The following is an example, but your paths may differ:

    ```shell
    cp /path/to/FAACIFP_enhanced "${HOME}/Library/Application\ Support/Steam/steamapps/common/X-Plane\ 11/Custom\ Data/earth_424.dat"
    ```

By default, the program removes duplicate localizers that are specified for
airports that use another airport's localizer on one of their approaches (e.g.
[KVNY LDA-C](https://skyvector.com/files/tpp/2004/pdf/00552LDAC.PDF)). The
reason these are removed is that it causes unpredictable behavior when using the
localizer in X-Plane. If you would like to turn this functionality off, set the
`remove_duplicate_locs` flag to `false`:

```shell
 enhance-faa-cifp --output=/path/to/FAACIFP_enhanced --remove_duplicate_locs=false /path/to/FAACIFP18
```

### Help

You can print the help for the program by running:

```shell
enhance-faa-cifp --help
```

## Known Issues

This is a project that will likely encounter a handful of known issues as users
test it in practice. The ones I am aware of are listed below:

- Some localizer bearings are more than 5 degrees off the original one published
  by the FAA. The ones I have detected are IBRL, IPIA, IVVS, and IYKM. A
  warning is logged for these localizers, but the new (potentially incorrect)
  value is persisted. See wallaceicy06/enhance-faa-cifp#1.

## Side Note For Pilots

I wrote this enhancer because I discovered the bug myself while practicing the
ILS 28R at
[KMOD](https://skyvector.com/airport/MOD/Modesto-City-Co-Harry-Sham-Field-Airport)
in X-Plane. I popped out of the clouds to see that the runway was offset
hundreds of feet to my left. Having flown the approach in real life, I know
that this is not accurate. This project was my deep dive into learning why that
error happened, and to offer a potential fix.

You can read more about my process solving this bug in my [blog
post](https://seanharger.com/posts/hundredths-of-degrees-from-death/).

## Contributing

Contributions to this repository are welcome, within the scope of the problem
of enhancing the CIFP data. Please submit bugs/suggestions to the issues section
and I will get back to you as soon as possible.
