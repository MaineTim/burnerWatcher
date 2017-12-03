# burnerWatcher
A simple daemon to monitor an oil burner's firing times.

I use this on a Raspberry Pi Model B which sits in my basement, monitoring various things including the oil burner.
As I didn't want to log the data on the Pi itself, it sends the data via HTTP to a logging daemon on another machine.

The burner is monitored using an MID400 opto-isolator chip tied to the thermostat control line. When the oil burner
is idle, the line is at 24 VAC, and the output of the MID400 is low. When the burner fires, the control line goes low
(<1 VAC) and the MID400 output goes high.

This was written in part as an exercise in learning Golang, and for a really specific use-case.
It's likely pretty bad Golang, but it works, and if it can be of any use to anyone, help yourself...

The config file must be in the same directory as the executable.
The "signal" URL starts a logging session on the logging server, which records temperature entries from sensors
around the house during and after the run.
The "time" URL signals the end of the run, and gives the time data to the logging server.
The "temperature" URL is unsupported on the logging server at present, but is intended for "one-shot" temperature
entries.

The "partner" logging server is at: https://github.com/MaineTim/httpLogger
