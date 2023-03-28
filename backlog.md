Things still to do:

Legacy TODO:

    Discuss whether we can actually delete transactions from the queue...
    Potential issue with someone sending an old transaction that has already been done
    ^ Old transaction Id's can be found in alreadyBroadcasted

    If transactions arrive faster than sequencer can add them to
    the block, it will add transactions forever, potential issue?
    Maybe insert some locks in startBroadcastingBlock?
    ^We make a new list with the length, m, being whatever we read it to be,
    then we take the first m Id's and add those to the block.

Questions for TA:

    none

Already answered questions:

For special RSA key pair:

    Does sequencers public key, and the public key used for the block, need to be different?
    aka: does sequencer need 2 different pairs of keys, or can we duplicate them?
    This does simplify things, IF we are allowed to do this!

        Answer:
        Only problem is that we use it multiple times, but that is OK for this handin

For blocks order:

    Should we handle if block number 2 arrives before block number 1 in some instance where the
    network is slow enough for this to happen?

        Answer:
        This cannot happen if we implemented the system correctly, since everyone that
        has seen block 2 has always seen block 1, works somewhat recursively

For receivedTransaction:

    Ask if OK to check that a transaction is done here?

        Answer:
        Yep its perfectly fine, we need to do it there

For receivedBlock:

    There might be an issue here: if a transaction has not made it into the queue,
    we will not be doing it even if our block says we should be doing it, ask a TA or something...

        Answer:
        This cannot happen if we implemented the system correctly, since the
        sequencer should also have broadcasted that transaction before broadcasting
        the block containing the order to do the transaction.
