# Algorand Esports Match Aggregator (AEMA) 

This project contains the source for an oracle on the Algorand blockchain that provides recent esports match results. It is licensed under the permissive zlib license. 

### Supported Games

* [x] Counter-Strike: Global Offensive
* [ ] League of Legends (next)
* [ ] Dota 2
* [ ] Valorant

### Project Structure
The repo is divided into modules: one core module and one module for each game. Each game module contains functions to fetch and aggregate match data and publish them on the blockchain. 

### Relevant Resources
* [What is Algorand?](https://developer.algorand.org/docs/get-started/basics/why_algorand/)
* [Smart Contracts](https://developer.algorand.org/docs/get-details/dapps/smart-contracts/)