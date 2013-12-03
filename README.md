scantopc
==========

This program provides the "ScanToComputer" functionality given by HP for they multi-functions printers for windows users.

Place your document on the platen or in the automatic document feeder (Adf). Select Scan on printer display, then select a computer (a destination) and file format (PDF or JPEG).

Double side scanning with single side ADF: 
Place your original pile of document in the ADF, scan it. This will produice a pdf file.
Flip them and scan it again. This will produce a second pdf file. If the second one has same page number as previous one, both are merged into a document having same name as firts job, will "-doubleside" into the name.



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

# TODO: 
- Fix Qnap problem
- Manage several MFP on the network.
- better error management

# CHANGE LOG
* 0.2.2
	Fix: better error management 
	Fix: better deconnection handling
	Fix: reconnection error
	Fix: PowerDown event handling
	Code reorganisation: usage of event manager (again)
	Change in double side handling: if the new scan job has same number of pages as previous one, both jobs are merged into one PDF
	Changed Printer Event constant pulling by usage by using timout / etag.
	 
* 0.2.1
	Fix: Error when image takes to long
	Fix: Turn off TRACE mode by default 
* 0.2
	- Code reorganisation
	- Multi settings: creation of several "destinations", each destination uses different settings:
		HOSTNAME (Normal) : Same settings as before
		HOSTNAME (LowRes) : 75DPI scanning 
	- Fix: uncontroled go routine number
	- Soft double side scanning whin sigle side ADF
* 0.1
	- Fix endless scanning when using flat bed (bad checking of AgingStamp)

