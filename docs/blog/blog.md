# Blog: AnonEvote

**MICHAEL-WALL**

## 001 - Functional Specification

The Functional Specification was submitted on 24th November. It highlights the goals for the system along with the proposed functionality and architecture that the system will use. The following are the main points in summary.

AnonEvote aims to provide three core functions for voters:
- Tamper-proof ballots
- Voter anonymity
- Integrity of counting votes

The system will be built using a blockchain database which is distributed over a peer-to-peer network. This will prevent the votes from being tampered with. Cryptography will be used to encrypt voter ballots in such a way that they are not tied to any user to keep their identity hidden. The votes should also be countable in some verifiable manner. The counting function should idealy be publicly usable so that any voter can verify the results of the vote.

For more information please see the functional spec document on [GitLab](https://gitlab.computing.dcu.ie/wallm22/2017-ca400-wallm22/raw/master/docs/functional-spec/func-spec.pdf) or [GitHub](https://github.com/CPSSD/voting/blob/master/docs/functional-spec/func-spec.pdf).

## 002 - Research on Homomorphic encryption schemes

One candidate encryption system I came across was the [Pailler Cryptosystem](http://www.cs.tau.ac.il/~fiat/crypt07/papers/Pai99pai.pdf) which has homomorphic properties that lend themselves well to use in electronic voting.

In short, the product of a set of ciphertexts will decrypt to the sum of their corresponding plaintexts. This assumes that all the messages have been encrypted using the same public key. The cryptosystem is non-deterministic, meaning that the same plaintext encrypted using the same public key will not give the same ciphertext. This is implemented through the use of a random value r during the encryption process which should ensure a negligible chance that a collision will occur.

Some issues which arise from the use of this cryptosystem:
- With access to the private key, an attacker could simply decrypt the individual votes, potentially allowing them to link a vote to a voter. We cannot allow this.
- A user who is encrypting 1 vote for a candidate could encrypt a vote with the value of 2 votes for the candidate, and this would not be evident in the decryption of the combined ciphertexts.

Some areas to research to potentially solve the above issues:
- The use of a sharded key system which would allow the decryption to be performed incrementally; using the shards of a private key owned by N parties where the private key is never created from its component parts.
- The use of some zero-knowledge proof could verify that the vote cast contains either a 1 or a 0 vote (for or against), without revealling which way the user has voted. This could possibly be implemented using the Fiat-Shamir heuristic.

## 003 - Research on sharded and shared keys

One of the viable schemes to accomplish this goal is [Shamir's Secret Sharing scheme](https://cs.jhu.edu/~sdoshi/crypto/papers/shamirturing.pdf). It involves creating a polynomial of K degrees to represent the secret (in this case the private key), and then constructing N points from the polynomial for N participants (where N >= K). A polynomial of K degrees will require K of these points in order to retrieve the secret.

The idea behind this is that a treshold can be set in the case of one or more of the participants not being present to reconstruct the secret.

The [Lagrange basis polynomial](https://en.wikipedia.org/wiki/Lagrange_polynomial) can be constructed using the any subset of K points in order to get the secret value, represented by the constant a0 in the polynomial:
f(x) = a0 + a1.x + a2.x^2 + ... + aK-1.x^k-1

This will allow a key to be distributed over a number of clients, but brings with it the limitation of requiring a trusted user to generate and susbsequently destroy the original version of the private key.

## 004 - Optimizations of Lagrange Polynomial Interpolations

While looking at the formula for interpolating Lagrange Polynomials, I have realized that some of the calculations do not need to be performed in order to retrieve the secret value.

Since the secret value is represented by f(x) where x = 0, all of the components with a variable x will be essentially removed, so we will not need to produce them in the first place.

This will reduce the number of multiplications required to retrieve the secret. I am not sure of the total savings of this optimization, but I belive it would be an exponential saving as the number of points required for the interpolation increases.

## 005 - Revised System Architecture

After discussion with my supervisor we finalized the high level design for the system. As expected, there were some of the initial goals which I had set out in my functional specification document which turned out to be impractical to implement.

These limitations were mainly due to:
- conflicting requirements
- aspects of the cryptography which simply were not possible to satisfy simultaneously
- compromising other core goals

Below is a diagram indicating the main components of the system. Two sources of the image are included.

![Image of System Architecture - GitHub](https://github.com/CPSSD/voting/blob/master/docs/blog/images/high-level-system-architecture.png)
![Image of System Architecture - GitLab](https://gitlab.computing.dcu.ie/wallm22/2017-ca400-wallm22/raw/306a63569cc5066772c26d6c27eb2ee3d510914b/docs/blog/images/high-level-system-architecture.png)

The image depicts the following system:
- Each user is assigned a randomized unique string from a list of known strings, in such a way that no one knows who is assigned which string
- A trusted individual creates a keypair for the election
- The trusted individual constructs a number of shareable keys from the private key, as described in Shamir's secret sharing scheme previously
- These keys are distributed to nodes of the system, and the original private key is destroyed
- The public key for the election is publicly available
- A user can now create a vote, and encrypt it with the public election key
- The user sends both their encrypted vote, along with their unique token which verifies their eligibility to vote, to a number of nodes in the system
- Each node verifies that they have not seen the voting token before, and that the token is one from the predefined list
- The node then signs the encrypted vote, and competes to create a block to add to the blockchain
- The nodes are dissuaded from misbehaving as if they begin to change votes or tokens, their behaviour will differ from the other nodes in the system, thus identifying themselves as untrustworthy
- Once the election is over, nodes will collaborate to reform the secret private key, and publish it
- Any user can then tally the results, and perform analyses on the blockchain to calculate statistics relating to discrepencies and such

This does introduce vulnerabilites to the system, such as the user being able to prove that they voted one way or another (through accurately predicting the decryption of their vote on the chain). However this system prioritises the anonymity of a user. If a user wishes to remain anonymous, there is no way to tie a vote back to a user. This fact is based on the assumption that the voting tokens are distributed in such a way that they are not ever tied to an individual.

It is in the best interest of the individual nodes to behave. This is intended to prevent attempts to tamper with ballots, perform ballot stuffing, or to misplace ballots.

The use of homomorphic encryption will allow nodes to perform the potentially compute expensive task of tallying votes in their encrypted form before the election has ended. They will only being able to reveal the true value of the election at the end once the secret key has been reconstructed.

## 006 - Implementing the blockchain core functions

Implementing my own blockchain was one of the most challenging aspects of the project so far. The main components of the blockchain are: transactions, blocks of transactions, the chain of blocks, peers, proof of work (or hashing), and consensus forming. The consensus algorithm will be addressed in further detail in another post.

I first implemented broadcasting of all the transactions to the network, and the subsequent collection of those transactions. If a valid transaction is received, it is stored in a pool of uncommited transactions and then broadcast to all peers. Transactions are valid if they do not exist in (our version of) the blockchain, and if we do not currently have the transaction in our pool. Once a node has enough transactions in its pool, it will begin to work on creating a block.

Creating a block involves performing a proof of work. This proof of work involves computing hashes of the block of candidate transactions while incrementing a nonce value contained in the block header. This process is repeated until a partial hash collision is found by the node. This partial collision is a matching number of zeros at the beginning of the hex representation of the hash of the block. An N-bit collision requires on average 2^(N/2) hashes to have a greater than 50% chance of occuring, so in our case, the dificulty of computing an N-character collision (in hex) is 2^(N*8). When the proof of work is completed, the new block is stored in our blockchain, and is susbsequently broadcast to all known peers. This is the behaviour of each node.

At any point, a node may receive from its peers: a transaction (as described above), a block update, or a request for peers.

Because calculating a proof of work is computationally expensive, we do not want to waste time working on a block if a valid block is received or if someone else has a longer chain than us. Once a block update is received, it is verified and validated concurrently while hashing is ongoing. If the new block is a valid addition to our chain, we send a signal to stop hashing, and add the new block to our chain. Then we can broadcast it to our peers and start working on the next new block. If the block is not the next one in our chain, we check to see if it belongs to a chain that claims to be longer than ours. If it comes from a longer chain, we will stop working on our hash, and then get the new chain from the node that sent us the block update. We will then send the block update which caused us to update to the longer chain to our peers. If none of these situations occur, we can be confident that the block is invalid or that it comes from a chain that is shorter than ours, so we do not aid in propagating it through the network. Because of the way this consensus operates (adopting the longest chain we have seen) it becomes more and more difficult to have invalid transactions or modified blocks accepted in the network.

If a request for peers is received, the two nodes will merge their peers-list together to keep up-to-date on the newest peers in the network. This means that an original source node is only required to bootstrap the launch of the distributed network. Once enough peers have connected (and enough stay online) you can get in contact with peers simply by connecting to one initial node that you are aware of on the network.

One thing to note, if our node happens to accept a new longer chain, we must take measures to ensure that no transactions from our chain are lost in the process of adopting the new chain. This involves re-broadcasting any transactions that were unique to our chain to the network.

## 007 - The consensus algorithm

At the core of managing a distributed database is agreeing on which version of the data is the (most) correct version. Because there is no central authority we must take measures to keep everyone on the same page and in agreement about the version of the chain that is (most) valid. One way we could do this is by allowing each node to "vote" on what version they think is valid, and then taking the most popular choice as the correct version. This may seem like a good idea, but when all users can be anonymous, it brings up more problems than it solves. However we would still like to perform some sort of vote or consensus on which version to accept.

The idea behind the consensus algorithm is to solve this issue and prevent a number of attacks on the network. One of the attacks we are trying to mitigate is a Sybil attack. This is where an attacker invents an overwhelming number of identities to cast votes for their invalid version of the blockchain. Since all users can be anonymous, this cannot be helped. Another attack which might occur is for an attacker to replace transactions or blocks, or even outright refuse to broadcast them.

Nodes express their "vote" for a version of the chain, or display their acceptance of it, by building new blocks on top of it. Because creating a proof of work requires computational power, an attacker cannot "invent" machines to perform its hashing for free. The proof of work is "proof" of a certain amount of effort having been expended to create a chain of blocks. For an attacker to modify an old block, it would need enough computing power to create blocks faster than the network can, hence being able to surpass the natural growth of the chain. In practice this would require the attacker to control 51% of the computational power in the network. This becomes infeasible to do as the network grows in size. Also, when a node sees a longer chain that is valid, this is proof that more effort had been expended to create the longer chain than their own, so this is what they should accept. If they do not accept it, they will be creating blocks for their own chain and broadcasting them, but the majority of nodes on the network will not accept these blocks because they are from a shorter chain.

While forks in the chain can occur, they usually do not last very long. As the two versions grow, one will eventually grow quicker, since there is a greater total of hashing power working towards extending this version of the chain. If a node refuses to broadcast other transactions, this is okay because the majority of nodes will broadcast it. As long as there is a majority of honest nodes in the network, then the chain can be considered secure.
