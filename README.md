# blockpit-lido

<!--toc:start-->
- [blockpit-lido](#blockpit-lido)
  - [Building](#building)
  - [Usage](#usage)
<!--toc:end-->

A simple cli tool to fetch the rewards history from lido's api and

transform them into a csv that blockpit can understand.

## Building

`make`

## Usage

`./bpl --address 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045`

Additionally see "bpl -h":

```
Usage of ./bpl:
  -address string
        address of the wallet you wish to generate a blockpit compatible csv for
  -archive-rate string
         (default "false")
  -asset string
         (default "stETH")
  -currency string
         (default "USD")
  -integration string
         (default "Lido stETH")
  -label string
         (default "Staking")
  -lido-api-url string
         (default "https://stake.lido.fi/api/rewards")
  -only-rewards string
         (default "true")
  -out string
        --out <filename> -- defaults to '<year>-report.csv'
  -year string
         (default "2023")
```

You can probably leave most of the stuff as is, apart from the year.

The currency flag doesn't change anything in the output, because blockpit seems

to be able to infer the value from whatever data they got.
