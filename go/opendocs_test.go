package opendocs_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	opendocs "github.com/modernice/opendocs/go"
	"github.com/modernice/opendocs/go/internal"
	"github.com/modernice/opendocs/go/patch"
)

func TestRepository_Patch(t *testing.T) {
	root := filepath.Join(internal.Must(os.Getwd()), "testdata", "gen", "opendocs-patch")
	internal.WithRepo("basic", root, func(repoFS fs.FS) {
		wantPatch := patch.New(repoFS)

		repo := opendocs.Repo(root)
		patch := repo.Patch()

		internal.AssertPatch(t, wantPatch, patch)
	})
}

// TODO: Implement documentation generation
// func TestRepository_Generate(t *testing.T) {
// 	sourceRoot := filepath.Join(internal.Must(os.Getwd()), "testdata", "gen", "opendocs-generate-source")
// 	internal.InitRepo("basic", sourceRoot)

// 	expectedPatch := opendocs.Repo(sourceRoot).Patch()

// 	root := filepath.Join(internal.Must(os.Getwd()), "testdata", "gen", "opendocs-generate")
// 	internal.InitRepo("basic", root)

// 	repo := opendocs.Repo(root)

// 	patch, err := repo.Generate()
// 	if err != nil {
// 		t.Fatalf("generate documentation: %v", err)
// 	}

// 	internal.AssertPatch(t, expectedPatch, patch)
// }
