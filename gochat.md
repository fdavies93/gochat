# GoChat

## Timeline

Conplete by 2023-12-3 with README and docker packaging.

## Complete

## Todo

- [x] Allow getting list of rooms in server
- [x] Allow changing rooms
- [x] Show id in sent messages
- [ ] Allow changing username
- [ ] Allow listing other users (?)
- [ ] Allow PMs to other users
- [ ] Ensure dropped connections are cleaned up by incorporating pings / timeouts
- [x] Parallelised writes to clients
    - If not implemented, one client not responding can hold up the whole
      server for a stupidly long amount of time
- [ ] Document with comments / README
- [ ] Split into submodules
- [ ] Make the basic chat interface sexier
- [ ] Setup a Docker / Compose file for easy deployment

