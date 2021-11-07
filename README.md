# Schulze method Go library

[![Go](https://github.com/janos/schulze/workflows/Go/badge.svg)](https://github.com/janos/schulze/actions)
[![PkgGoDev](https://pkg.go.dev/badge/resenje.org/schulze)](https://pkg.go.dev/resenje.org/schulze)
[![NewReleases](https://newreleases.io/badge.svg)](https://newreleases.io/github/janos/schulze)

Schulze is a Go implementation of the [Schulze method](https://en.wikipedia.org/wiki/Schulze_method) voting system. The system is developed in 1997 by Markus Schulze. It is a single winner preferential voting. The Schulze method is also known as Schwartz Sequential dropping (SSD), cloneproof Schwartz sequential dropping (CSSD), the beatpath method, beatpath winner, path voting, and path winner.

The Schulze method is a [Condorcet method](https://en.wikipedia.org/wiki/Condorcet_method), which means that if there is a candidate who is preferred by a majority over every other candidate in pairwise comparisons, then this candidate will be the winner when the Schulze method is applied.

White paper [Markus Schulze, "The Schulze Method of Voting"](https://arxiv.org/pdf/1804.02973.pdf).

## Compute

`Compute(v VoteMatrix) (scores []Score, tie bool)` is the core function in the library. It implements the Schulze method on the most compact required representation of votes, called `VoteMatrix`. It returns the ranked list of choices from the matrix, with the first one as the winner. In case that there are multiple choices with the same score, the returned `tie` boolean flag is true.

`VoteMatrix` holds number of votes for every pair of choices. A convenient structure to record this map is implemented as the `Voting` type in this package, but it is not required to be used.

## Voting

`Voting` is the in-memory data structure that allows voting ballots to be submitted, to export the `VoteMatrix` and also to compute the ranked list of choices.

The act of voting represents calling the `Voting.Vote(b Ballot) error` function with a `Ballot` map where keys in the map are choices and values are their rankings. Lowest number represents the highest rank. Not all choices have to be ranked and multiple choices can have the same rank. Ranks do not have to be in consecutive order.

## Example

```go
package main

import (
	"fmt"
	"log"

	"resenje.org/schulze"
)

func main() {
	// Create a new voting.
	e := schulze.NewVoting("A", "B", "C", "D", "E")

	// First vote.
	if err := e.Vote(schulze.Ballot{
		"A": 1,
	}); err != nil {
		log.Fatal(err)
	}

	// Second vote.
	if err := e.Vote(schulze.Ballot{
		"A": 1,
		"B": 1,
		"D": 2,
	}); err != nil {
		log.Fatal(err)
	}

	// Calculate the result.
	result, tie := e.Compute()
	if tie {
		log.Fatal("tie")
	}
	fmt.Println("winner:", result[0].Choice)
}
```

## Alternative voting implementations

Function `Compute` is deliberately left exported with `VoteMatrix` map to allow different voting implementations. The `Voting` type in this package is purely in-memory but in reality, a proper way of authenticating users and storing the voting records are crucial and may require implementation with specific persistence features.

## License

This application is distributed under the BSD-style license found in the [LICENSE](LICENSE) file.
