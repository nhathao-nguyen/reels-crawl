package capcut

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateUUID(t *testing.T) {
	uuid1 := GenerateUUID()
	uuid2 := GenerateUUID()

	if len(uuid1) != 36 {
		t.Errorf("Expected UUID length 36, got %d (%s)", len(uuid1), uuid1)
	}
	if uuid1 == uuid2 {
		t.Errorf("UUIDs should be unique, got duplicate: %s", uuid1)
	}
}

func TestDeepCopyMap(t *testing.T) {
	src := map[string]any{
		"name": "test",
		"nums": []any{1, 2, 3},
		"nested": map[string]any{
			"key": "val",
		},
	}

	dst, err := DeepCopyMap(src)
	if err != nil {
		t.Fatalf("DeepCopyMap failed: %v", err)
	}

	// Mutate dst and ensure src is unchanged
	dst["name"] = "modified"
	if src["name"] == "modified" {
		t.Errorf("Source mutated after modifying destination copy")
	}
}

func TestFindCapCutExecutable(t *testing.T) {
	exe := FindCapCutExecutable()
	if exe != "" {
		if _, err := os.Stat(exe); err != nil {
			t.Errorf("Returned CapCut exe does not exist: %s", exe)
		}
	}
}

func TestCleanCapCutProcesses(t *testing.T) {
	// Should execute without panicking or erroring
	err := CleanCapCutProcesses()
	if err != nil {
		t.Errorf("CleanCapCutProcesses failed: %v", err)
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "source.txt")
	dst := filepath.Join(tmpDir, "sub", "dest.txt")

	_ = os.WriteFile(src, []byte("hello world"), 0644)
	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile dst failed: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("Content mismatch: got %s, want hello world", string(content))
	}
}

func TestEnsureTemplateDirectory(t *testing.T) {
	tmpDrafts := t.TempDir()
	tplDir := filepath.Join(tmpDrafts, "template-fb")
	_ = os.MkdirAll(tplDir, 0755)

	found, err := EnsureTemplateDirectory(tmpDrafts, "template-fb")
	if err != nil {
		t.Fatalf("EnsureTemplateDirectory failed: %v", err)
	}
	if found != tplDir {
		t.Errorf("Expected %s, got %s", tplDir, found)
	}

	// Fallback to "template"
	tmpDrafts2 := t.TempDir()
	tplDir2 := filepath.Join(tmpDrafts2, "template")
	_ = os.MkdirAll(tplDir2, 0755)

	found2, err := EnsureTemplateDirectory(tmpDrafts2, "template-fb")
	if err != nil {
		t.Fatalf("EnsureTemplateDirectory fallback failed: %v", err)
	}
	if found2 != tplDir2 {
		t.Errorf("Expected %s, got %s", tplDir2, found2)
	}
}
