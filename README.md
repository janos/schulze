# Schulze method Go library

[![Go](https://github.com/janos/schulze/workflows/Go/badge.svg)](https://github.com/janos/schulze/actions)
[![PkgGoDev](https://pkg.go.dev/badge/resenje.org/schulze)](https://pkg.go.dev/resenje.org/schulze)
[![NewReleases](https://newreleases.io/badge.svg)](https://newreleases.io/github/janos/schulze)

Schulze is a Go implementation of the [Schulze method](https://en.wikipedia.org/wiki/Schulze_method) voting system. The system is developed in 1997 by Markus Schulze. It is a single winner preferential voting. The Schulze method is also known as Schwartz Sequential dropping (SSD), cloneproof Schwartz sequential dropping (CSSD), the beatpath method, beatpath winner, path voting, and path winner.

The Schulze method is a [Condorcet method](https://en.wikipedia.org/wiki/Condorcet_method), which means that if there is a candidate who is preferred by a majority over every other candidate in pairwise comparisons, then this candidate will be the winner when the Schulze method is applied.

White paper [Markus Schulze, "The Schulze Method of Voting"](https://arxiv.org/pdf/1804.02973.pdf).

## Usage

`Vote` and `Compute` are the core functions in the library. They implement the Schulze method on the most compact required representation of votes, here called preferences that is properly initialized with the `NewPreferences` function. `Vote` writes the `Ballot` values to the provided preferences and `Compute` returns the ranked list of choices from the preferences, with the first one as the winner. In case that there are multiple choices with the same score, the returned `tie` boolean flag is true.

The act of voting represents calling the `Vote` function with a `Ballot` map where keys in the map are choices and values are their rankings. Lowest number represents the highest rank. Not all choices have to be ranked and multiple choices can have the same rank. Ranks do not have to be in consecutive order.

### Additional features

This implementation of Schulze voting method adds capabilities to

- remove the ballot from voting results, allowing the vote to be changed
- add, remove or rearrange choices at any time during the voting process, while preserving consistency of the state just as the choices were present from the beginning

`Unvote` function allows to update the pairwise preferences in a way to cancel the previously added `Ballot` to preferences using `Vote` function. It is useful to change the vote without the need to re-vote all ballots.

`SetChoices` allows to update the pairwise preferences if the choices has to be changed during voting. New choices can be added, existing choices can be removed or rearranged. New choices are ranked as previous ballots did not rank them or were ranked the last, as they were present in initial choices but were not ranked in any ballots.

## Voting

`Voting` holds number of votes for every pair of choices. It is a convenient construct to use when the preferences slice does not have to be exposed, and should be kept safe from accidental mutation. Methods on the Voting type are not safe for concurrent calls.

## Results

Results are provided by the `Compute` function which returns the ranked list of choices from the preferences, but also the iterator function over all `Duels` that represent pairwise comparisons between two choices. Duels can be used to represent and analyze results in more details.

## Example

```go
package main

import (
 "fmt"
 "log"

 "resenje.org/schulze"
)

func main() {
 choices := []string{"A", "B", "C"}
 preferences := schulze.NewPreferences(len(choices))

 // First vote.
 if _, err := schulze.Vote(preferences, choices, schulze.Ballot[string]{
  "A": 1,
 }); err != nil {
  log.Fatal(err)
 }

 // Second vote.
 if _, err := schulze.Vote(preferences, choices, schulze.Ballot[string]{
  "A": 1,
  "B": 1,
  "C": 2,
 }); err != nil {
  log.Fatal(err)
 }

 // Calculate the result.
 result, _, tie := schulze.Compute(preferences, choices)
 if tie {
  log.Fatal("tie")
 }
 fmt.Println("winner:", result[0].Choice)
}
```

## License

This application is distributed under the BSD-style license found in the [LICENSE](LICENSE) file.
