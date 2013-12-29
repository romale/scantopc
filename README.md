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
- On Qnap ARM TS119:  SGEFAULT with saving as PDF (update: seems to by tied to QNAP and not ARM architecture. Need help here)

# TODO: 
- Service mode: been able to run as a service in *nix world
- Manage several MFP on the network.
- better error management (still in progress)

# CHANGE LOG
* 0.4.0
	- Code reorganisation
		All HP relative code is now placed into a separate package. The package localize the scanner, handle scan jobs and provide Scan To Computer feature. It's now using a DocumentBatcher interface to comunicate whith rest of the code.
	- Fix: Small delay after job termination before handling a new job
	- Fix: SCANTOPC / ScanEvent goroutine not ended properly
	- Change default resolution to 300 dpi. tesseract gives better results. 	

* 0.3.1
	- add posibility to use pdfunite. pdftool option can give tool to be used when joining PDF pages. Accepted values: pdftk and pdfunite
	- add option to enable / disable OCR. External dependancies are checked only when OCR flag is true
	- Changed double side mode: now, the destination called 'OCR verso' must be selected on the printer to merge previous job with this one
	- Change deflaut destinations to OCR and OCR verso. The last is to indicate that the current job is the 2nd side of previous. Whenever  previous job can't be the recto side of the job, the job is considered as recto and not verso. 
	- Fix error management 
* 0.3
	- Fix: better error management 
	- Fix: better deconnection handling  
	- Fix: reconnection error  
	- Fix: PowerDown event handling  
	- Code reorganisation: usage of event manager (again)  
	- Change in double side handling: if the new scan job has same number of pages as previous one, both jobs are merged into one PDF  
	- Changed Printer Event constant pulling by usage by using timout / etag.  
	- Fix: better cleanup images in obvious cases
	 
* 0.2.1
	Fix: Error when image takes to long to be scanned (hi resolution)
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

