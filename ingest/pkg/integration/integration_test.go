package integration

import (
	"github.com/neongreen/mono/ingest/pkg/git"
	"github.com/neongreen/mono/ingest/pkg/testutil"
	"testing"
)

func TestGitIngestion(t *testing.T) {
	testutil.WithTempHome(t)

	testRepo := testutil.NewTempGitRepo(t, []testutil.GitCommit{
		{
			Message: "Initial commit",
			Files: map[string]string{
				"test.txt": "Hello, World!",
			},
		},
		{
			Message: "Second commit",
			Files: map[string]string{
				"test2.txt": "Second file",
			},
		},
	})

	db := testutil.OpenDatabase(t)

	runID, err := db.CreateRun(testRepo, "git")
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	commits, err := git.WalkRepository(testRepo, nil)
	if err != nil {
		t.Fatalf("WalkRepository: %v", err)
	}

	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}

	for _, commit := range commits {
		commitID, err := db.CreateCommit(
			runID,
			commit.Hash,
			commit.Author,
			commit.AuthorEmail,
			commit.Committer,
			commit.CommitterEmail,
			commit.Date,
			commit.Message,
			commit.ParentHashes,
		)
		if err != nil {
			t.Fatalf("CreateCommit: %v", err)
		}

		for _, file := range commit.Files {
			var blobID *int64
			if len(file.Content) > 0 {
				id, err := db.GetOrCreateBlob(file.Content, file.SHA256Hash)
				if err != nil {
					t.Fatalf("GetOrCreateBlob: %v", err)
				}
				blobID = &id
			}

			if err := db.CreateFile(commitID, file.Path, file.Size, file.Mode, blobID); err != nil {
				t.Fatalf("CreateFile: %v", err)
			}
		}
	}

	if err := db.UpdateRunItemCount(runID); err != nil {
		t.Fatalf("UpdateRunItemCount: %v", err)
	}

	if err := db.FinishRun(runID, "completed"); err != nil {
		t.Fatalf("FinishRun: %v", err)
	}

	testutil.AssertSingleRun(t, db, "git", 2)
}
