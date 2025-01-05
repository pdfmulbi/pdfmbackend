package fpdf

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pkg/errors"
)

// MergePDFBytes merges two PDF files provided as []byte and returns the merged result as []byte.
func MergePDFBytes(pdf1, pdf2 []byte) ([]byte, error) {
	// Create in-memory buffers for input PDFs
	input1 := bytes.NewReader(pdf1)
	input2 := bytes.NewReader(pdf2)

	// Create temporary files to save in-memory PDFs
	tmpFile1, err := os.CreateTemp("", "pdf1_*.pdf")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp file for pdf1")
	}
	defer os.Remove(tmpFile1.Name())

	tmpFile2, err := os.CreateTemp("", "pdf2_*.pdf")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp file for pdf2")
	}
	defer os.Remove(tmpFile2.Name())

	// Write the in-memory bytes to temporary files
	if _, err := io.Copy(tmpFile1, input1); err != nil {
		return nil, errors.Wrap(err, "failed to copy pdf1 data to temp file")
	}
	if _, err := io.Copy(tmpFile2, input2); err != nil {
		return nil, errors.Wrap(err, "failed to copy pdf2 data to temp file")
	}

	// Close the files so they can be read later by pdfcpu
	if err := tmpFile1.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close temp file for pdf1")
	}
	if err := tmpFile2.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close temp file for pdf2")
	}

	// Debugging: Check file sizes
	info1, _ := os.Stat(tmpFile1.Name())
	info2, _ := os.Stat(tmpFile2.Name())
	fmt.Printf("Temp file 1: %s, Size: %d bytes\n", tmpFile1.Name(), info1.Size())
	fmt.Printf("Temp file 2: %s, Size: %d bytes\n", tmpFile2.Name(), info2.Size())

	// Create another temporary file to store the merged output
	mergedFile, err := os.CreateTemp("", "merged_*.pdf")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temp file for merged PDF")
	}
	defer os.Remove(mergedFile.Name())

	// Prepare the input files for merging
	inputFiles := []string{tmpFile1.Name(), tmpFile2.Name()}

	// Merge the PDFs using output writer
	outFile, err := os.Create(mergedFile.Name())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create output file for merged PDF")
	}
	defer outFile.Close()

	err = api.Merge(outFile.Name(), inputFiles, nil, nil, false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to merge PDFs")
	}

	// Read the merged PDF into memory and return it as []byte
	mergedPDF, err := os.ReadFile(mergedFile.Name())
	if err != nil {
		return nil, errors.Wrap(err, "failed to read merged PDF file")
	}

	return mergedPDF, nil
}
