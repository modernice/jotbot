package opendocs_test

import (
	"os"
	"path/filepath"
	"testing"

	opendocs "github.com/modernice/opendocs/go"
	"github.com/modernice/opendocs/go/internal"
)

var (
	repoRoot = filepath.Join(internal.Must(os.Getwd()), "testdata", "gen", "finder")
	repoFS   = os.DirFS(repoRoot)
)

func init() {
	internal.InitRepo(repoRoot)
}

func TestFinder_Uncommented(t *testing.T) {
	f := opendocs.NewFinder(repoFS)

	result, err := f.Uncommented()
	if err != nil {
		t.Fatal(err)
	}

	for path, findings := range result {
		t.Logf("%s: %d findings", path, len(findings))
		for _, finding := range findings {
			t.Logf("\t%s: %s", finding.Path, finding.Identifier)
		}
	}

	t.Error()
}
