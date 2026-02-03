package store

import (
	"fmt"
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
				s := New(root)
				if err := s.Init("benchmark", ""); err != nil {
					b.Fatal(err)
				}
				for j := 0; j < count; j++ {
					if _, err := s.Learn(fmt.Sprintf("Lesson %d", j), "topic", models.SeverityNormal, "", nil); err != nil {
						b.Fatal(err)
					}
				}
				if err := s.Close(); err != nil {
					b.Fatal(err)
				}
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
			s := New(root)
			if err := s.Init("benchmark", ""); err != nil {
				b.Fatal(err)
			}
			for j := 0; j < count; j++ {
				if _, err := s.Learn(fmt.Sprintf("Lesson %d", j), "topic", models.SeverityNormal, "", nil); err != nil {
					b.Fatal(err)
				}
			}
			if err := s.Close(); err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s2 := New(root)
				if _, err := s2.List("", "", "", time.Time{}, time.Time{}); err != nil {
					b.Fatal(err)
				}
				if err := s2.Close(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkRandomAccess(b *testing.B) {
	dir := b.TempDir()
	root := filepath.Join(dir, ".dory")
	s := New(root)
	if err := s.Init("benchmark", ""); err != nil {
		b.Fatal(err)
	}
	var ids []string
	for j := 0; j < 100; j++ {
		id, err := s.Learn(fmt.Sprintf("Lesson %d", j), "topic", models.SeverityNormal, "", nil)
		if err != nil {
			b.Fatal(err)
		}
		ids = append(ids, id)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := ids[i%len(ids)]
		if _, err := s.Show(id); err != nil {
			b.Fatal(err)
		}
	}
	if err := s.Close(); err != nil {
		b.Fatal(err)
	}
}

func BenchmarkList(b *testing.B) {
	for _, count := range []int{10, 100, 500} {
		b.Run(fmt.Sprintf("items=%d", count), func(b *testing.B) {
			dir := b.TempDir()
			root := filepath.Join(dir, ".dory")
			s := New(root)
			if err := s.Init("benchmark", ""); err != nil {
				b.Fatal(err)
			}
			for j := 0; j < count; j++ {
				if _, err := s.Learn(fmt.Sprintf("Lesson %d", j), "topic", models.SeverityNormal, "", nil); err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := s.List("", "", "", time.Time{}, time.Time{}); err != nil {
					b.Fatal(err)
				}
			}
			if err := s.Close(); err != nil {
				b.Fatal(err)
			}
		})
	}
}
