
# Demo of an Interbank Transfer Broker

## Requirements

We need a working Redis database. There are several ways to do this, the easiest
one is to run Redis in Docker:

https://redis.io/docs/latest/operate/oss_and_stack/install/install-stack/docker/

The advantage of this method is that, in addition to having the server on the
standard port (_6379_, without username/password), we have Redis
[Insight](https://redis.io/insight/) on port _8001_ to view the contents of our
database.

## Description

There is a ./scripts/gen.py to generate random account files stored in the
./data directory.

Then we have 3 scripts in the ./cmd directory to run the test:

**./cmd/init/main.go**
: Clears the Redis database, reads the CSV file from the ./data directory, and
loads this content into Redis. This would be the first script to run to test the
connection with Redis. We can do it by running:
> make init

**./cmd/broker/main.go**
: This script creates the number of parallel processes according to the
configuration. Each process establishes a connection to Redis and generates a
number of random transactions between accounts. Each transaction performs the
following steps in Redis:

1. It makes a *Get* to obtain the current amount of money in the source account
and checks if it has enough money to send.

2. If the account has enough money, it goes to step 3. If not, the transaction
stops, and an event is sent to the Redis queue indicating that the transaction
was canceled. This step is critical because it needs to ensure that there are no
race conditions. That is, no other process should withdraw money from the
account in the time we read the available amount and deduct the money.

3. If the destination and origin banks are the same, it does nothing with the
bank balance keys. However, if they are different, it deducts the amount from
the origin bank, adds it to the destination bank, and calculates the balances
between the two banks.

4. Afterward, it *deposits* the money into the destination account.

These would be the Redis commands in a successful transaction:

```redis
$ MULTI
$ get from_account
$ incrby from_account -amount
$ EXEC
$ incrby from_bank -amount
$ incrby to_bank amount
$ incrby fromBank_toBank -amount
$ incrby toBank_fromBank amount
$ incrby to_account amount
```

5. At the end of the process, an event is sent to the Redis stream that receives
the transactions.

We can run this script by executing:

> make broker

**./cmd/consume/main.go**
: This process consumes events from the Redis stream. Since this stream contains
all successful transactions and errors, it can calculate the balance of
individual accounts, total balances, and interbank balances. So it performs
these calculations and checks that the values in Redis are correct. It also
verifies that the sum of the bank balances equals zero and that the sum of the
interbank balances equals zero as well. The idea behind this is to ensure that
there were no errors in parallel transfers and no race condition errors. It is
best to run this command simultaneously with the broker so that it consumes the
stream as it is generated. We can run it like this:

> make consume
