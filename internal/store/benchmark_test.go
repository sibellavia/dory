package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sibellavia/dory/internal/models"
)

func BenchmarkWrite(b *testing.B) {
	for _, count := range []int{10, 100, 500} {
		b.Run(fmt.Sprintf("items=%d", count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				dir := b.TempDir()
				root := filepath.Join(dir, ".dory")
				s := NewSingle(root)
				s.Init("benchmark", "")
				for j := 0; j < count; j++ {
					s.Learn(fmt.Sprintf("Lesson %d", j), "topic", models.SeverityNormal, "", "", nil)
				}
				s.Close()
			}
		})
	}
}

func BenchmarkOpen(b *testing.B) {
	for _, count := range []int{10, 100, 500} {
		b.Run(fmt.Sprintf("items=%d", count), func(b *testing.B) {
			// Setup: create store with items
			dir := b.TempDir()
			root := filepath.Join(dir, ".dory")
			s := NewSingle(root)
			s.Init("benchmark", "")
			for j := 0; j < count; j++ {
				s.Learn(fmt.Sprintf("Lesson %d", j), "topic", models.SeverityNormal, "", "", nil)
			}
			s.Close()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s2 := NewSingle(root)
				s2.List("", "", "", time.Time{}, time.Time{}) // Forces open
				s2.Close()
			}
		})
	}
}

func BenchmarkRandomAccess(b *testing.B) {
	dir := b.TempDir()
	root := filepath.Join(dir, ".dory")
	s := NewSingle(root)
	s.Init("benchmark", "")
	for j := 0; j < 100; j++ {
		s.Learn(fmt.Sprintf("Lesson %d", j), "topic", models.SeverityNormal, "", "", nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("L%03d", (i%100)+1)
		s.Show(id)
	}
	s.Close()
}

func BenchmarkList(b *testing.B) {
	for _, count := range []int{10, 100, 500} {
		b.Run(fmt.Sprintf("items=%d", count), func(b *testing.B) {
			dir := b.TempDir()
			root := filepath.Join(dir, ".dory")
			s := NewSingle(root)
			s.Init("benchmark", "")
			for j := 0; j < count; j++ {
				s.Learn(fmt.Sprintf("Lesson %d", j), "topic", models.SeverityNormal, "", "", nil)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s.List("", "", "", time.Time{}, time.Time{})
			}
			s.Close()
		})
	}
}

func countFiles(root string) int {
	count := 0
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}
