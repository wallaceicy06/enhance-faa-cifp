package enhance

import (
	"bytes"
	"io/ioutil"
	"testing"
)

const testDataFile = "test_data"

func TestProcess(t *testing.T) {
	testData, err := ioutil.ReadFile(testDataFile)
	if err != nil {
		t.Fatalf("Could not read test data file: %v", err)
	}

	in := bytes.NewReader(testData)
	var out bytes.Buffer
	if err := Process(in, &out); err != nil {
		t.Fatalf("Process() = %v want <nil>", err)
	}
	if !bytes.Equal(out.Bytes(), testData) {
		t.Errorf("Process() out content does not equal input")
	}
}
