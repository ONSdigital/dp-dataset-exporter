package filter

import (
	"fmt"
	filterCli "github.com/ONSdigital/dp-api-clients-go/filter"
	"math"
)

type Splitter struct {
	batches int
}

type Option struct {
	DimensionName string
	Options       []string
}

func NewSplitter(batchSize int) *Splitter {
	return &Splitter{
		batches: batchSize,
	}
}

func (s *Splitter) Split(filter *filterCli.Model) []*filterCli.Model {

	dimensionToSplitIndex := 0
	numberOfOptions := 0

	// determine the dimension with the largest cardinality
	for i, dimension := range filter.Dimensions {

		if len(dimension.Options) > numberOfOptions {
			dimensionToSplitIndex = i
			numberOfOptions = len(dimension.Options)
		}
	}


	optionsPerBatch := int(math.Ceil(float64(numberOfOptions) / float64(s.batches)))
	dimensionToSplit := filter.Dimensions[dimensionToSplitIndex]

	fmt.Println("found largest dimension: ", dimensionToSplit.Name)
	fmt.Println("batches: ", s.batches)
	fmt.Println("number per batch: ", optionsPerBatch)

	// take the split options and create new split dimensions for each batch
	var splitFilters []*filterCli.Model
	for i := 0; i < numberOfOptions; i += optionsPerBatch {

		// if we are on the last batch, it may have less options in it.
		endIndex := i + optionsPerBatch
		if endIndex > numberOfOptions {
			endIndex = numberOfOptions
		}

		fmt.Println("Start index: ", i)
		fmt.Println("End index: ", endIndex)

		options := dimensionToSplit.Options[i:endIndex]
		fmt.Println("Options in batch: ", options)

		splitFilter := &filterCli.Model{
			FilterID:   filter.FilterID,
			Dimensions: []filterCli.ModelDimension{
				{
					Name:    dimensionToSplit.Name,
					Options: options,
				},
			},
		}



		// add the other dimensions that have not been split
		for j, dimension := range filter.Dimensions {
			if j != dimensionToSplitIndex{
				splitFilter.Dimensions = append(splitFilter.Dimensions, dimension)
			}
		}

		splitFilters = append(splitFilters, splitFilter)

	}

	return splitFilters
}
