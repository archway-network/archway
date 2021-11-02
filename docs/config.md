Configuration
=============

```bash
archwayd config <flag> <value>
```

It allows you to set a default value for each given flag.

First, set up the address of the full-node you want to connect to:

```bash
archwayd config node <host>:<port>

# example: archwayd config node https://77.87.106.33:26657
```

If you run your own full-node locally, just use `tcp://localhost:26657` as the address.


```
Finally, let us set the `chain-id` of the blockchain we want to interact with:

```bash
archwayd config chain-id archway
```

To see the value of a config key, just run it without any given value:

```bash
archwayd config chain-id
```