scantopc
==========

***

This program provides the "ScanToComputer" functionality given by HP for they multi-functions printers for windows users.

Place your document on the platten or in the automatic document feeder (Adf). Select Scan on printer display, then select a computer (a destination) and file format (PDF or JPEG).

Double side scanning with single side ADF: 
Place your original pile of document in the ADF, scan it. Flip them and scan it again.
If the second batch starts before 2 minutes from the last one, both batches will be merged into a single document, with correct page order.


Tested with printer model Officejet 6700 on linux, freebsd.


Usage of ./scantopc:  
>   -d="": shorthand for -destination  
>   -destination="": Folder where images are strored (see help for tokens)  
>   -name="localhost": Name of the computer visible on the printer (default: $hostname)  
>   -printer="": Printer URL like http://1.2.3.4:8080, when omitted, the device is searched on the network  
>   -trace=false: Enable traces  

Allowed tokens for dir / file name are:  
	%Y  Year (4 digits):      2014  
	%y  Year (2 digits):      14	     
	%d  Day (2 digits):       03  
	%A  Weekday (long):       Monday  
	%a  Weekday (short):      Mon  
	%m  Month (2 digits):       
	%I  Hour (12 hour):       05  
	%H  Hour (24 hour):       17  
	%M  Minute (2 digits):    54  
	%S  Second (2 digits):    20  
	%p  AM / PM:              PM  


This litle piece of code is my first programming experience with Go language and my first coding experience since a decade.

# Known problems
- Fails on Qnap arm system; somthing is wrong with PDF generation.
- Doesn't recover connection after  printer switch on/off
- when using platen, the user MUST terminate scaning job properly on the printer. Otherwise, the job is never terminated



# TODO: 
- Fix Qnap problem
- Manage several MFP on the network.
- Permit "soft" duplex scanning 
- better error management

# CHANGE LOG

* 0.2
	- Code reorganisation
	- Multi settings: creation of several "destinations", each destination uses different settings:
		HOSTNAME (Normal) : Same settings as before
		HOSTNAME (LowRes) : 75DPI scanning 
	- Fix: uncontroled go routine number
* 0.1
	- Fix endless scanning when using flat bed

