# Algorand Esports Match Aggregator (AEMA) 

This project contains the source for an oracle on the Algorand blockchain that provides recent esports match results. It is licensed under the permissive zlib license. 

### Supported Games

* [x] Counter-Strike: Global Offensive
* [ ] League of Legends (next)
* [ ] Dota 2
* [ ] Valorant

### Project Structure
The repo is divided into modules: a `core` module, and a module for each game (e.g. `csgo`, `lol`, ...).
The core module contains a `Buffer` interface that provides persistent storage, while abstracting away 
the storage mechanism. Because flushing changes to the buffer into the blockchain is expensive, the
buffers can implement write-limits and publish health metrics.

The game modules fetch and aggregate match data, and try to keep the buffer up-to-date. The proposed 
changes to the buffer can be regularly submitted. The game modules are designed to be resilient. They
must deal with latency, timeouts, failure to write, crashes and recovery. 

### Relevant Resources
* [What is Algorand?](https://developer.algorand.org/docs/get-started/basics/why_algorand/)
* [Smart Contracts](https://developer.algorand.org/docs/get-details/dapps/smart-contracts/)