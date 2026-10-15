module github.com/rfwlab/bench/todomvc

go 1.26.9

require github.com/rfwlab/rfw/v2 v2.0.0-beta.19

// Benchmark builds against the local checkout.
replace github.com/rfwlab/rfw/v2 => ../../..
