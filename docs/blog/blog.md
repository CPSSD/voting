# Blog: AnonEvote

**MICHAEL-WALL**

## Functional Specification

The Functional Specification was submitted on 24th November. It highlights the goals for the system along with the proposed functionality and architecture that the system will use. The following are the main points in summary.

AnonEvote aims to provide three core functions for voters:
- Tamper-proof ballots
- Voter anonymity
- Integrity of counting votes

The system will be built using a blockchain database which is distributed over a peer-to-peer network. This will prevent the votes from being tampered with. Cryptography will be used to encrypt voter ballots in such a way that they are not tied to any user to keep their identity hidden. The votes should also be countable in some verifiable manner. The counting function should idealy be publicly usable so that any voter can verify the results of the vote.

For more information please see the functional spec document on [GitLab](https://gitlab.computing.dcu.ie/wallm22/2017-ca400-wallm22/raw/master/docs/functional-spec/func-spec.pdf) or [GitHub](https://github.com/CPSSD/voting/blob/master/docs/functional-spec/func-spec.pdf).
