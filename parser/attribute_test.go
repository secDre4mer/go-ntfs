package parser

import (
	"encoding/json"
	"testing"

	"github.com/sebdah/goldie"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	input []Run
	out   []*MappedReader
}

type readerTestCase struct {
	input []Run
	out   []Range
}

var (
	TestCases = []testCase{
		{input: []Run{
			{474540, 47},
			{0, 1},
			{48, 1213},
			{0, 3},
		}, out: []*MappedReader{
			&MappedReader{0, 474540, 32, 0x400, 0, false, nil},
			&MappedReader{32, 474572, 16, 0x400, 15, false, nil},
			&MappedReader{48, 474588, 1200, 0x400, 0, false, nil},
			&MappedReader{1248, 475788, 16, 0x400, 13, false, nil},
		}},
		// A compressed run followed by a sparse run longer
		// than compression size.
		{input: []Run{
			{1940823, 2},
			{0, 30}, // This is really {0, 14}, {0, 16} merged together.
		}, out: []*MappedReader{

			// A compressed run followed by sparse run.
			&MappedReader{0, 1940823, 16, 0x400, 2, false, nil},
			&MappedReader{16, 0, 16, 0x400, 0, true, &NullReader{}},
		}},
	}

	ReaderTestCases = []readerTestCase{
		{input: []Run{
			{474540, 47},
			{0, 1},
			{48, 1213},
			{0, 3},
		}, out: []Range{
			{0, 32 * 0x400, false},
			{32 * 0x400, 16 * 0x400, false},
			{48 * 0x400, 1200 * 0x400, false},
			{1248 * 0x400, 16 * 0x400, false},
		}},

		// A compressed run followed by a sparse run longer
		// than compression size.
		{input: []Run{
			{1940823, 2},
			{0, 30}, // This is really {0, 14}, {0, 16} merged together.
		}, out: []Range{

			// A compressed run followed by sparse run.
			{0, 16 * 0x400, false},
			{16 * 0x400, 16 * 0x400, true},
		}},
	}
)

func TestNewCompressedRunReader(t *testing.T) {
	for _, testcase := range TestCases {
		runs := NewCompressedRangeReader(
			testcase.input, 0x400, nil, 16)
		assert.Equal(t, testcase.out, runs.runs)
	}
}

func TestReaderRanges(t *testing.T) {
	for _, testcase := range ReaderTestCases {
		runs := NewCompressedRangeReader(
			testcase.input, 1024, nil, 16)
		assert.Equal(t, testcase.out, runs.Ranges())
	}
}

func TestMappedReaderRanges(t *testing.T) {
	// Mapped reader contains two real runs with gaps:
	// 0-100 sparse,
	// 100-300 read,
	// 300-500 sparse,
	// 500-500 real  - last region is clipped because the mapping length ins 1000.
	mapped_reader := &MappedReader{
		FileOffset:   0,
		TargetOffset: 0,
		Length:       1000,
		ClusterSize:  1000,
		Reader: &RangeReader{
			runs: []*MappedReader{
				{
					FileOffset:   100,
					TargetOffset: 100,
					ClusterSize:  1000,
					Length:       200,
				},
				{
					FileOffset:   500,
					TargetOffset: 500,
					ClusterSize:  1000,
					Length:       1000,
				},
			},
		},
	}

	result_json, _ := json.MarshalIndent(mapped_reader.Ranges(), " ", " ")
	goldie.Assert(t, "TestMappedReaderRanges", result_json)
}
