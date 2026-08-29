# Schulze method Go library

[![Go](https://github.com/janos/schulze/workflows/Go/badge.svg)](https://github.com/janos/schulze/actions)
[![PkgGoDev](https://pkg.go.dev/badge/resenje.org/schulze)](https://pkg.go.dev/resenje.org/schulze)
[![NewReleases](https://newreleases.io/badge.svg)](https://newreleases.io/github/janos/schulze)

Schulze is a Go implementation of the [Schulze method](https://en.wikipedia.org/wiki/Schulze_method) voting system. The system was developed in 1997 by Markus Schulze. It is a single winner preferential voting system. The Schulze method is also known as Schwartz Sequential dropping (SSD), cloneproof Schwartz sequential dropping (CSSD), the beatpath method, beatpath winner, path voting, and path winner.

The Schulze method is a [Condorcet method](https://en.wikipedia.org/wiki/Condorcet_method), which means that if there is a candidate who is preferred by a majority over every other candidate in pairwise comparisons, then this candidate will be the winner when the Schulze method is applied.

White paper: [Markus Schulze, "The Schulze Method of Voting"](https://arxiv.org/pdf/1804.02973.pdf).

## Usage

`Vote` and `Compute` are the core functions in the library. They implement the Schulze method on the most compact required representation of votes, called preferences, initialized with the `NewPreferences` function. `Vote` writes the `Ballot` values to the provided preferences and `Compute` returns the ranked list of choices from the preferences, with the first one as the winner. In case of a tie for the top position, the returned `tie` boolean flag is true.

The act of voting represents calling the `Vote` function with a `Ballot` map where keys in the map are choices and values are their rankings. Lowest number represents the highest rank. Not all choices have to be ranked and multiple choices can have the same rank. Ranks do not have to be in consecutive order.

Alternatively, `VoteRanked` accepts a pre-ordered slice of ranks (`[][]C`), avoiding map hashing overhead and enabling high-performance, low-allocation ingestion.

### Features

- **Reversible voting**: The `Unvote` function allows rolling back a previously added `Ballot` using its returned `Record`, enabling voters to change their vote without re-tallying all ballots.
- **Dynamic candidate adjustment**: `SetChoices` updates the pairwise preferences if choices need to be added, removed, or rearranged during active voting, while mathematically preserving consistency.
- **Choice Validation**: `ValidateChoices` ensures choices list is not empty and contains no duplicate candidates.
- **Direct ranked slice voting**: `VoteRanked` accepts ranked slices of candidates (`[][]C`).
- **Generic Ballot Ingestion**: `VoteFrom` enables voting directly from arbitrary external structures using a custom ranking function.
- **Type transformations (`Map`)**: `Record`, `Result`, `Duel`, and `Voting` support Go method type parameters via `.Map(...)` to translate between choice identifiers (e.g. `UUID` $\leftrightarrow$ `string` $\leftrightarrow$ `int`).
- **Standard Go iterators**: `Duels` provides a standard `iter.Seq[*Duel[C]]` iterator for range-over-func loops.
- **Hardware SIMD Acceleration**: Uses ARM64 NEON vector assembly kernels and unrolled multi-core pipelines to accelerate inner matrix relaxation.
- **Fast Winner Calculation**: `Winner` returns the election winner using an $O(N^2)$ Condorcet fast-path with fallback to full beatpath computation for cycles.
- **Configurable numeric precision**: Pairwise preferences support generic `Number` types (`int`, `uint32`, `uint16`, etc.) for memory optimization.

## Voting State

### `Voting`

`Voting[C]` holds the number of votes for every pair of choices. It is a convenient construct to use when the preferences slice does not have to be exposed, and should be kept safe from accidental mutation. Methods on the `Voting` type are not safe for concurrent calls.

## Results & Duels

Results are computed by `Compute`, returning the ranked list of choices and an iterator over all pairwise `Duels`. With Go range-over-func iterators, you can iterate over duels directly:

```go
for duel := range v.Duels() {
    winner, defeated := duel.Outcome()
    // analyze duel outcomes...
}
```

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
	v := schulze.NewVoting(choices)

	// First vote using map Ballot.
	if _, err := v.Vote(schulze.Ballot[string]{
		"A": 1,
	}); err != nil {
		log.Fatal(err)
	}

	// Second vote using ranked slices.
	if _, err := v.VoteRanked([][]string{
		{"A", "B"},
		{"C"},
	}); err != nil {
		log.Fatal(err)
	}

	// Calculate the result.
	result, _, tie := v.Compute()
	if tie {
		log.Fatal("tie")
	}
	fmt.Println("winner:", result[0].Choice)
}
```

## License

This application is distributed under the BSD-style license found in the [LICENSE](LICENSE) file.
