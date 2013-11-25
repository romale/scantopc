// scantopc project doc.go

/*
# scantopc

***

This litle piece of code is my first programming experience with Go language and my first coding experience since a decade.

This program provides the "ScanToComputer" functionality given by HP for they multi-functions printers for windows users


Tested with printer model Officejet 6700 on linux, freebsd.
Fails on Qnap arm system.


## Profile
A profile is a set of scanning paramters like
- resolution for PDFs or JPGs
- Color / Gray
- Name of generated file / destination folder
- virtual duplex scanning

Available profiles are listed on printer as different destinations for "Scan To Computer".
On Officejet 6700, you are limited at 15 different destinations, actual or simulated.

# TODO:
- Fix Qnap problem
- Manage several MFP on the network.
- Permit "soft" duplex scanning

# CHANGE LOG

* 0.1
  - Fix endless scanning when using flat bed


*/
package main
