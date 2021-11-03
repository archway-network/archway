# Governance

Governance is the process from which users in the Archway network can come to consensus on software upgrades, parameters of the mainnet or signaling mechanisms through text proposals. 
This is done through voting on proposals, which will be submitted
by `ARCH` holders on the mainnet.

Some considerations about the voting process:

- Voting is done by bonded `ARCH` holders on a `1` bonded `ARCH` 1 vote basis
- Delegators inherit the vote of their validator if they don't vote
- Votes are tallied at the end of the voting period (2 weeks on mainnet) where
  each address can vote multiple times to update its `Option` value (paying the transaction fee each time),
  only the most recently cast vote will count as valid
- Voters can choose between options `Yes`, `No`, `NoWithVeto` and `Abstain`
- At the end of the voting period, a proposal is accepted iff:
  - `(YesVotes / (YesVotes+NoVotes+NoWithVetoVotes)) > 1/2`
  - `(NoWithVetoVotes / (YesVotes+NoVotes+NoWithVetoVotes)) < 1/3`
  - `((YesVotes+NoVotes+NoWithVetoVotes) / totalBondedStake) >= quorum`

For more information about the governance process and how it works, please check
out the Governance module [specification](https://github.com/cosmos/cosmos-sdk/tree/master/x/gov/spec).

## Create a Governance Proposal

In order to create a governance proposal, you must submit an initial deposit along with a title and description. 
Various modules outside of governance may
implement their own proposal types and handlers (eg. parameter changes), where
the governance module itself supports `Text` proposals. Any module
outside of governance has it's command mounted on top of `submit-proposal`.

To submit a `Text` proposal:

```bash
archwayd tx gov submit-proposal \
  --title=<title> \
  --description=<description> \
  --type="Text" \
  --deposit="1000000uARCH" \
  --from=<name> \
  --chain-id=<chain_id>
```

You may also provide the proposal directly through the `--proposal` flag which points to a JSON file containing the proposal.

To submit a parameter change proposal, you must provide a proposal file as its contents are less friendly to CLI input:

```bash
archwayd tx gov submit-proposal param-change <path/to/proposal.json> \
  --from=<name> \
  --chain-id=<chain_id>
```

Where `proposal.json` contains the following:

```json
{
  "title": "Param Change",
  "description": "Update max validators",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": 105
    }
  ],
  "deposit": [
    {
      "denom": "stake",
      "amount": "10000000"
    }
  ]
}
```

>**Warning:**
Currently parameter changes are _evaluated_ but not _validated_, so it is very important that any `value` change is valid \(i.e. correct type and within bounds\) for its respective parameter, e.g. `MaxValidators` should be an integer and not a decimal.

Proper vetting of a parameter change proposal should prevent this from happening
(no deposits should occur during the governance process), but it should be noted
regardless.

>**Tip:**
The `SoftwareUpgrade` is currently not supported as it's not implemented and currently does not differ from the semantics of a `Text` proposal.


### Query Proposals

Once created, you can now query information of the proposal:

```bash
archwayd query gov proposal <proposal_id>
```

Or query all available proposals:

```bash
archwayd query gov proposals
```

You can also query proposals filtered by `voter` or `depositor` by using the corresponding flags.

To query for the proposer of a given governance proposal:

```bash
archwayd query gov proposer <proposal_id>
```

## Increase Deposit

In order for a proposal to be broadcasted to the network, the amount deposited must be above a `minDeposit` value (initial value: `512000000uARCH`). If the proposal you previously created didn't meet this requirement, you can still increase the total amount deposited to activate it. Once the minimum deposit is reached, the proposal enters voting period:

```bash
archwayd tx gov deposit <proposal_id> "10000000uARCH" \
  --from=<name> \
  --chain-id=<chain_id>
```

>**Note:** Proposals that don't meet this requirement will be deleted after `MaxDepositPeriod` is reached.

### Query Deposits

Once a new proposal is created, you can query all the deposits submitted to it:

```bash
archwayd query gov deposits <proposal_id>
```

You can also query a deposit submitted by a specific address:

```bash
archwayd query gov deposit <proposal_id> <depositor_address>
```

## Vote on a Proposal

After a proposal's deposit reaches the `MinDeposit` value, the voting period opens. Bonded `ARCH` holders can then cast vote on it:

```bash
archwayd tx gov vote <proposal_id> <Yes/No/NoWithVeto/Abstain> \
  --from=<name> \
  --chain-id=<chain_id>
```

### Query Votes

Check the vote with the option you just submitted:

```bash
archwayd query gov vote <proposal_id> <voter_address>
```

You can also get all the previous votes submitted to the proposal with:

```bash
archwayd query gov votes <proposal_id>
```

## Query proposal tally results

To check the current tally of a given proposal you can use the `tally` command:

```bash
archwayd query gov tally <proposal_id>
```

## Query Governance Parameters

To check the current governance parameters run:

```bash
archwayd query gov params
```

To query subsets of the governance parameters run:

```bash
archwayd query gov param voting
archwayd query gov param tallying
archwayd query gov param deposit
```
