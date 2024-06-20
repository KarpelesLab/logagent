# logagent

Simple log daemon running and gathering logs. It has multiple modes of operation:

* It creates a unix socket in /run/logagent (if root) or /tmp/.logagent-1000.sock if non-root.
* Each connection must follow the logagent protocol
* A connected client can request a logging descriptor that will be sent to it as a pipe, and can be used to log messages coming from a command (useful when running a command)

We provide a go client that can perform all those things in logclient.

Messages are sent to either the klab system (log or slog messages) or to any specified compatible sink, based on the parameters or environment.

Logagent can update itself while keeping all sockets as is. Clients talking with logagent won't notice an update happened.

## Upgrade procedure

Upon start, the daemon will attempt to connect to itself first, if it manages to, it will try to request a takeover, in which case all sockets will be passed from the current daemon to the new one.
