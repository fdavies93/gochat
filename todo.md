# GoChat Todo

## Complete

## Todo

- [ ] Allow getting list of rooms in server
- [ ] Allow changing rooms / username
- [ ] Ensure dropped connections are cleaned up by incorporating pings / timeouts
- [ ] Document with comments / README
- [ ] Split into submodules
- [ ] Setup a Docker / Compose file for easy deployment

## Not Todo (with reasoning)

**Parallelised writes to clients.** The server needs to write the majority of messages via hubs / rooms / channels or directly via PMs. Currently there's one write process which directs to clients as appropriate. At minimum this would need a hierarchical structure of channels to pass messages up / down from clients to rooms and to the server as a whole; this seems over-complex for a demo app.
