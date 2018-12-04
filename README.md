# Connection Metadata (CNXMD) Protocol

The CNXMD protocol defines a header that a client proxy can send to a
CNXMD-aware server proxy in order to pass arbitrary key-value pairs containing
information about the connection to be proxied. It has similarities to
the [PROXY protocol](https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt)
and could be viewed as a generalization of it.

## Typical Flow

- A client attempts to communicate with a destination server.
- Between the client and server reside a client-side proxy and a server-side proxy.
- The client connects to the client-side proxy.
- The client-side proxy connects to the server-side proxy.
- The client-side proxy sends a CNXMD header to the server-side proxy. The header contains a list of KV pairs.
- The server-side proxy locates the destination server, possibly using information contained in the KV pairs.
- The client-side proxy pipes the connection from the client.
- The server-side proxy pipes the connection to the destination server.
- The client and server can now communicate.

## Specification

The CNXMD header is composed of a sequence of newline ('\n') terminated lines.
The first line must be "CONNECTION_METADATA/1.1\n"
What follows is zero or mnore key-value pairs formatted as "key=value\n"
All characters following the '=' character are interpreted as belonging to the value.
The header is terminated with the blank line "\n"
The specification does not define a maximum line length or number of lines.
However, specific implementations may impose their own limits.

Example:
```
CONNECTION_METADATA/1.1\n
foo=bar\n
jane=john=jack\n
\n
```
The above header sends the key-values: {"foo":"bar", "jane":"john=jack"}

## Use case(s)

One use case is to augment the
[Server Name Indication](https://tools.ietf.org/html/rfc6066#section-3) feature
of TLS. A server-side TLS proxy can use the Server Name (SN) specified in the SNI
header to locate the correct destination server.

There may be situations where a TLS client does not set the Server Name correctly,
or does not set it at all, making it impossible for the TLS proxy to route
the connection.

This problem can be solved by a combination of:
1. inserting a client-side CNXMD proxy between the client and TLS proxy
2. enhancing the TLS proxy to recognize the CNXMD header as an alternative
to the SNI header for routing purposes

Assuming the client-side proxy knows the appropriate SN for connections
initiated by the client, it inserts the SN into the CNXMD header. The SN can
be specified using a CNXMD kv entry (for e.g. using a key named "host").
Upon detecting a CNXMD header with the correct key, the TLS proxy removes
the header from the TLS stream and proxies the connection to the appropriate
destination server.

## Repo contents
This repo contains a go library with two useful functions:
- Parse() attempts to decode a CNXMD header from a byte slice. If successful,
it returns the length of the header and a map containing the KV pairs.
- ServeClientProxy() implements a client-side proxy. A CNXMD header is generated
from the supplied map containing KV pairs.