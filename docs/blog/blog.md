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
