module github.com/rfwlab/examples/dashboard

go 1.26.9

require github.com/rfwlab/rfw/v2 v2.0.0-beta.19

require golang.org/x/net v0.60.0 // indirect

// Always build against the checkout this example ships with.
replace github.com/rfwlab/rfw/v2 => ../..
