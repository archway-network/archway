<!--
order: 2
-->

# Begin-Block

Section describes the operations performed in the ABCI begin block call.

## Mint Calculation

Amount of tokens to be minted are calculated as follows:

### 1. Calculate time elapsed since last mint
$$ 
\tag{1}
timeElapsed = currentBlockTime - time_{lastBlockInfo} 
$$

where:

* $timeElapsed$ - the calculated time since the last time tokens were minted;
* $currentBlockTime$ - current block time;
* $time_{lastBlockInfo}$ - the time when last inflation was minted;

### 2. Calculate inflation for current block

$$
\tag{2}
 inflation = inflation_{lastBlockInfo} + 
\begin{cases}
    inflationChange_{params} \times timeElapsed, & \text bondedRatio < minBondedRatio_{params}\\
    -1 \times inflationChange_{params} \times timeElapsed, & \text bondedRatio > maxBondedRatio_{params}\\
    0 ,              & \text {otherwise}
\end{cases}
$$

where:

* $inflation$ - the inflation for the current block;
* $inflation_{lastBlockInfo}$ - the inflation for the last block mint;
* $inflationChange_{params}$ - the inflation change rate as configured as module param;
* $timeElapsed$ - the calculated time since the last time tokens were minted $(1)$;
* $bondedRatio$ - the amount of total supply which is currently staked;
* $minBondedRatio_{params}$ - minimum bonded ratio to aim for as configured as module param;
* $maxBondedRatio_{params}* - maximum bonded ratio to aim for as configured as module param;

### 3. Calculate amount of tokens to mint in the block based on the block inflation

$$
\tag{3}
tokens = inflation \times bondedTokenSupply \times timeElapsed_{(seconds)}/Year_{(seconds)}
$$

where:

* $tokens$ - the total amount of tokens to mint in the current block;
* $inflation$ - the inflation calculated for the current block $(2)$;
* $bondedTokenSupply$ - total amount of tokens currently staked;
* $timeElapsed_{(seconds)}$ - the calculated time since the last time tokens were minted $(1)$, measured in seconds,
* $Year_{(seconds)}$ - the number of seconds in a year

## Mint Distribution

With the calculated $tokens$ value from $(3)$, 
Next we loop through the configured inflation recipients from the module params. For each recipient we calculate the mint distribution based on the inflation ratio set.

$$
\tag{4}
distributionAmount_{recipient} = tokens \times ratio_{recipient}
$$

where:

* $distributionAmount_{recipient}$ - the amount of tokens to distribute to the recipient;
* $tokens$ - the total amount of tokens to mint in the current block $(3)$;
* $ratio_{recipient}$ - the percentage share of the inflation to be distributed to the recipient;


After the mint and distribution distribution, the latest mint inflation and timestamp is updated in the module store to be used the next block.
The distribution of the tokens is also stored in state such that any recipient module can access how many tokens it recieved in the given block